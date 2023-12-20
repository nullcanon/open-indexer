package loader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"open-indexer/model"
	"open-indexer/db"
	"open-indexer/handlers"
)


func LoadDataBase() {
	var insInfos []db.InscriptionInfo
	ins := db.InscriptionInfo{}
	ins.FetchInscriptionInfo(&insInfos)

	for _, tokens := range insInfos {
		handlers.Tokens[tokens.Ticks] = &model.Token{
			Trxs:tokens.Trxs,
			Tick:tokens.Ticks,
			Minted: model.NewDecimalFromStringValue(tokens.Minted),
			Holders:tokens.Holders,
			Max:model.NewDecimalFromStringValue(tokens.Total),
		}
		handlers.GetLogger().Info(tokens.Ticks, tokens.Trxs, tokens.Minted, tokens.Holders, tokens.Total)
	}


	var userBalances []db.UserBalances
	balance := db.UserBalances{}
	balance.FetchUserBalances(&userBalances)
	for _, users := range userBalances {
		amt := model.NewDecimalFromStringValue(users.Amount)
		userBalances := handlers.UserBalances[users.Address]
		if userBalances == nil {
			userBalances = make(map[string]*model.DDecimal)
		}
		userBalances[users.Ticks] = amt
		handlers.UserBalances[users.Address] = userBalances

		holderB := handlers.TokenHolders[users.Ticks]
		if holderB == nil {
			holderB = make(map[string]*model.DDecimal)
		}
		holderB[users.Address] = amt
		handlers.TokenHolders[users.Ticks] = holderB
	}
}

func LoadTransactionData(fname string) ([]*model.Transaction, error) {

	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var trxs []*model.Transaction
	scanner := bufio.NewScanner(file)
	max := 4 * 1024 * 1024
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	for scanner.Scan() {
		line := scanner.Text()
		//log.Printf(line)
		fields := strings.Split(line, ",")

		if len(fields) != 7 {
			return nil, fmt.Errorf("invalid data format", len(fields))
		}

		var data model.Transaction

		data.Id = fields[0]
		data.From = fields[1]
		data.To = fields[2]

		block, err := strconv.ParseUint(fields[3], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Block = block

		idx, err := strconv.ParseUint(fields[4], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Block = block

		data.Idx = uint32(idx)

		blockTime, err := strconv.ParseUint(fields[5], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Timestamp = blockTime
		data.Input = fields[6]

		trxs = append(trxs, &data)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return trxs, nil
}

func DumpTickerInfoToDB(
	blockNumbers uint64,
	tokens map[string]*model.Token,
	userBalances map[string]map[string]*model.DDecimal,
	tokenHolders map[string]map[string]*model.DDecimal,
) {
	var allTickers []string
	for ticker := range tokens {
		allTickers = append(allTickers, ticker)
	}
	sort.SliceStable(allTickers, func(i, j int) bool {
		return allTickers[i] < allTickers[j]
	})

	for _, ticker := range allTickers {
		info := tokens[ticker]

		inscriptionInfo := db.InscriptionInfo{
			Ticks: info.Tick,
			Total:info.Max.String(),
			Minted:info.Minted.String(),
			Holders:int32(len(tokenHolders[ticker])),
		}
		inscriptionInfo.Update(
			map[string]interface{}{
				"trxs": info.Trxs,
				"total":info.Max.String(),
				"minted":info.Minted.String(),
				"holders":len(tokenHolders[ticker]),
			})

		handlers.GetLogger().Info("Update inscriptionInfo secuess")
		handlers.GetLogger().Info("trxs:", info.Trxs,
			" total:",info.Max.String(),
			" minted:",info.Minted.String(),
			" holders:",len(tokenHolders[ticker]))
		
		// holders
		for holder, needUpdate := range handlers.UpdateUsers {
			if !needUpdate {
				continue
			}
			balance := tokenHolders[ticker][holder]
			user := db.UserBalances{
				Ticks: info.Tick,
				Address: holder,
				Amount: balance.String(),
			}
			user.Update(
				map[string]interface{}{
					"amount": balance.String(),
				})
			handlers.GetLogger().Info(ticker, holder)
			handlers.GetLogger().Info("Update balance secuess:", holder," amount: ", balance.String())
			handlers.UpdateUsers[holder] = false;
		}
	}

	blocnscan := db.BlockScan{}
	blocnscan.UptadeBlockNumber(blockNumbers);
}


func DumpTickerInfoMap(fname string,
	tokens map[string]*model.Token,
	userBalances map[string]map[string]*model.DDecimal,
	tokenHolders map[string]map[string]*model.DDecimal,
) {

	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatalf("open block index file failed, %s", err)
		return
	}
	defer file.Close()

	var allTickers []string
	for ticker := range tokens {
		allTickers = append(allTickers, ticker)
	}
	sort.SliceStable(allTickers, func(i, j int) bool {
		return allTickers[i] < allTickers[j]
	})

	for _, ticker := range allTickers {
		info := tokens[ticker]

		fmt.Fprintf(file, "%s trxs: %d, total: %s, minted: %s, holders: %d\n",
			info.Tick,
			info.Trxs,
			info.Max.String(),
			info.Minted,
			len(tokenHolders[ticker]),
		)

		// holders
		var allHolders []string
		for holder := range tokenHolders[ticker] {
			allHolders = append(allHolders, holder)
		}
		sort.SliceStable(allHolders, func(i, j int) bool {
			return allHolders[i] < allHolders[j]
		})

		// holders
		for _, holder := range allHolders {
			balance := tokenHolders[ticker][holder]

			fmt.Fprintf(file, "%s %s  balance: %s, tokens: %d\n",
				info.Tick,
				holder,
				balance.String(),
				len(userBalances[holder]),
			)
		}
	}
}