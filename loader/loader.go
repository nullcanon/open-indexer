package loader

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"open-indexer/db"
	"open-indexer/handlers"
	"open-indexer/model"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// func DumpTradeCache() {
// 	for trade := range handlers.TradeCache {
// 		trade.Update(
// 			map[string]interface{}{
// 				"ticks":        trade.Ticks,
// 				"status":       trade.Status,
// 				"from_address": trade.From,
// 				"to_address":   trade.To,
// 				"hash":         trade.Hash,
// 				"time":         trade.Time,
// 				"amount":       trade.Amount,
// 				"number":       trade.Number,
// 			})
// 	}
// }

func LoadDataBase() {

	// history := db.TradeHistory{}
	// handlers.InscriptionNumber = history.GetInscriptionNumber()
	// handlers.GetLogger().Info("InscriptionNumber load succees ", handlers.InscriptionNumber)

	var insInfos []db.InscriptionInfo
	ins := db.InscriptionInfo{}
	ins.FetchInscriptionInfo(&insInfos)

	for _, tokens := range insInfos {
		handlers.Tokens[tokens.Ticks] = &model.Token{
			Trxs:        tokens.Trxs,
			Tick:        tokens.Ticks,
			Minted:      model.NewDecimalFromStringValue(tokens.Minted),
			Holders:     tokens.Holders,
			Max:         model.NewDecimalFromStringValue(tokens.Total),
			Limit:       model.NewDecimalFromStringValue(tokens.Limit),
			CreatedAt:   tokens.CreatedAt,
			CompletedAt: int64(tokens.CompletedAt),
			Number:      tokens.Number,
			Creater:     tokens.Creater,
			Hash:        tokens.Hash,
		}
		handlers.TokenHolders[tokens.Ticks] = make(map[string]*model.DDecimal)
		// handlers.GetLogger().Info(tokens.Ticks, tokens.Trxs, tokens.Minted, tokens.Holders, tokens.Total)
	}

	handlers.GetLogger().Info("InscriptionInfo load succees")

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
	handlers.GetLogger().Info("UserBalances load succees")
}

func LoadAndPushRocketMsg() {
	var rocketMsg []db.RocketMsg
	rocketmsg := db.RocketMsg{}
	rocketmsg.FetchRocketMsg(&rocketMsg)
	for _, msgs := range rocketMsg {
		handlers.GetLogger().Infof("Rocket message from DB: %s\n", msgs.Message)
		_, err := handlers.ReockP.SendSync(context.Background(), primitive.NewMessage("new_list", []byte(msgs.Message)))
		if err != nil {
			handlers.GetLogger().Infof("Rocket faild  from DB: %s\n", err)
		} else {
			rocketmsg.DelMsg(msgs)
			handlers.GetLogger().Infof("Rocket seccess  from DB: %s\n", msgs.Message)
		}
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
	tokens map[string]*model.Token,
	userBalances map[string]map[string]*model.DDecimal,
	tokenHolders map[string]map[string]*model.DDecimal,
) {
	startTime := time.Now()
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
			Ticks:   info.Tick,
			Total:   info.Max.String(),
			Minted:  info.Minted.String(),
			Holders: int32(len(tokenHolders[ticker])),
		}
		inscriptionInfo.Update(
			map[string]interface{}{
				"trxs":         info.Trxs,
				"total":        info.Max.String(),
				"minted":       info.Minted.String(),
				"holders":      len(tokenHolders[ticker]),
				"mint_limit":   info.Limit.String(),
				"created_at":   info.CreatedAt,
				"completed_at": info.CompletedAt,
				"number":       info.Number,
				"creater":      info.Creater,
				"tx_hash":      info.Hash,
				"prec":         0,
			})

		// handlers.GetLogger().Info("Update inscriptionInfo secuess")
		// handlers.GetLogger().Info("trxs:", info.Trxs,
		// 	" total:", info.Max.String(),
		// 	" minted:", info.Minted.String(),
		// 	" holders:", len(tokenHolders[ticker]))

		// holders
		_, exists := handlers.Tokens[ticker]
		if !exists {
			continue
		}

		for holder, needUpdate := range handlers.UpdateUsers {
			if !needUpdate {
				continue
			}
			balance := tokenHolders[ticker][holder]
			if balance == nil {
				continue
			}
			user := db.UserBalances{
				Ticks:   info.Tick,
				Address: holder,
				Amount:  balance.String(),
			}
			user.Update(
				map[string]interface{}{
					"amount": balance.String(),
				})
			// handlers.GetLogger().Info(ticker, holder)
			// handlers.GetLogger().Info("Update balance secuess:", holder, " amount: ", balance.String())
		}
	}

	for holder, needUpdate := range handlers.UpdateUsers {
		if !needUpdate {
			continue
		}
		handlers.UpdateUsers[holder] = false
	}

	handlers.GetLogger().Info("DumpTickerInfoToDB succees ", time.Since(startTime))
}

func DumpBlockNumber() {
	if handlers.BlockNumber <= 0 {
		return
	}
	blocnscan := db.BlockScan{}
	blocnscan.UptadeBlockNumber(handlers.BlockNumber, handlers.InscriptionNumber)
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
