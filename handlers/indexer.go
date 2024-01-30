package handlers

import (
	"container/list"
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"open-indexer/db"
	"open-indexer/model"
	"open-indexer/rpc"
	"open-indexer/server"
	"open-indexer/utils"
	"strconv"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/ethereum/go-ethereum/common"
)

// var Inscriptions []*model.Inscription

var Tokens = make(map[string]*model.Token)

// address -> ticker -> balnce
var UserBalances = make(map[string]map[string]*model.DDecimal)

// ticker -> address -> number
var TokenHolders = make(map[string]map[string]*model.DDecimal)

// 数据发送变化的用户,刷新数据库
var UpdateUsers = make(map[string]bool)

var BlockNumber uint64 = 0

var InscriptionNumber uint64 = 0

var Ethrpc *rpc.EthBlockChainRPC

// var TradeCache = make(chan *db.TradeHistory, 512)
var TradeCache []*model.HistoryMessage

var ReockP rocketmq.Producer

var RocketQueue *list.List

func ProcessUpdateARC20(trxs []*model.Transaction) error {
	for _, inscription := range trxs {
		err := Inscribe(inscription)
		if err != nil {
			return err
		}
	}
	return nil
}

func notifyNewTick(asc20 *model.Asc20) {
	var tickMsg model.TickMessage
	tickMsg.Limit = asc20.Limit.String()
	tickMsg.Max = asc20.Max.String()
	tickMsg.Tick = asc20.Tick
	server.SubHandler.Publish(server.NewTick, &tickMsg)
}

func PushRockMQ() {
	for RocketQueue.Len() > 0 {
		element := RocketQueue.Front()

		message := element.Value.(string)
		logger.Infof("Rocket message: %s\n", message)
		_, err := ReockP.SendSync(context.Background(), primitive.NewMessage("new_list", []byte(message)))
		if err != nil {
			var parsedData map[string]interface{}
			errj := json.Unmarshal([]byte(message), &parsedData)
			if errj != nil {
				logger.Error("PushRockMQ JSON error:", errj)
			} else {
				if hashValue, ok := parsedData["id"].(string); ok {
					rocketmsg := db.RocketMsg{}
					rocketmsg.Hash = hashValue
					rocketmsg.Message = message
					err = rocketmsg.CreateRocketMsg(rocketmsg)
					logger.Error("rocketmsg.CreateRocketMsg:", errj)
				}
			}

		} else {
			logger.Infof("Rocket seccess: %s\n", message)
		}
		RocketQueue.Remove(element)
	}
}

func NotifyHistory() {
	server.SubHandler.Publish(server.History, TradeCache)
	TradeCache = TradeCache[:0]
}

func appendTradeCache(inscription *model.Inscription, tick string, amount string) {
	var tardeinfo model.HistoryMessage
	tardeinfo.Tick = tick
	tardeinfo.Status = "1"
	tardeinfo.From = inscription.From
	tardeinfo.To = inscription.To
	tardeinfo.Hash = inscription.Id
	tardeinfo.Time = inscription.Timestamp
	tardeinfo.Amount = amount
	tardeinfo.Number = inscription.Number
	if inscription.From == inscription.To {
		tardeinfo.Method = "mint"
	} else {
		tardeinfo.Method = "transfer"
	}

	TradeCache = append(TradeCache, &tardeinfo)
}

func Inscribe(trx *model.Transaction) error {

	hasFix := strings.HasPrefix(trx.Input, "0x646174613a")
	BlockNumber = trx.Block

	// data:,
	// contract address
	if hasFix || trx.To == strings.ToLower("0x368323Fd8b8BaaEC1615E9A78Dac27779F123f0A") {
		bytes, err := hex.DecodeString(trx.Input[2:])
		if err != nil {
			logger.Warn("inscribe err", err, " at block ", trx.Block, ":", trx.Idx)
			return nil
		}

		InscriptionNumber++
		var inscription model.Inscription
		inscription.Number = InscriptionNumber
		inscription.Id = trx.Id
		inscription.From = trx.From
		inscription.To = trx.To
		inscription.Block = trx.Block
		inscription.Idx = trx.Idx
		inscription.Timestamp = trx.Timestamp

		if !hasFix {
			// 获取交易log
			logger.Info("evm log To: ", trx.To)
			if err := handleEvmLog(&inscription); err != nil {
				logger.Info("error at ", inscription.Number)

				return err

			}
			return nil
		} else {

			input := string(bytes)

			sepIdx := strings.Index(input, ",")
			if sepIdx == -1 || sepIdx == len(input)-1 {
				return nil
			}
			contentType := "text/plain"
			if sepIdx > 5 {
				contentType = input[5:sepIdx]
			}
			content := input[sepIdx+1:]
			inscription.ContentType = contentType
			inscription.Content = content
			if err := handleProtocols(&inscription); err != nil {
				logger.Info("error at ", inscription.Number)

				return err
			}
		}
		// Inscriptions = append(Inscriptions, &inscription)
	}

	return nil
}

func handleEvmLog(inscription *model.Inscription) error {
	txReceipt, err := Ethrpc.GetTransactionReceipt(inscription.Id)
	if err == nil {
		subLogs := txReceipt.GetLogs()
		for _, subLog := range subLogs {
			subTopics := subLog.GetTopics()
			if subTopics[0] != "0x11c57dfe5e3d216d9107fdc9e7cca2cf14a175fd4de91d06c28a1ff238dbedce" {
				continue
			}
			subFrom := strings.ToLower(common.HexToAddress(subTopics[1]).String())
			subTo := strings.ToLower(common.HexToAddress(subTopics[2]).String())
			amt_str := subLog.GetData()[2:66]
			b := new(big.Int)
			b, ok := b.SetString(amt_str, 16)
			if !ok {
				continue
			}
			amt := model.NewDecimalFromStringValue(b.String())
			tick_srt := subLog.GetData()[194:]

			hexBytes, err := hex.DecodeString(tick_srt)
			hexBytes = utils.TrimLeftZeroes(hexBytes)
			hexBytes = utils.TrimRightZeroes(hexBytes)
			if err != nil {
				continue
			}
			tick := string(hexBytes)
			var asc20 model.Asc20
			asc20.Number = inscription.Number
			asc20.Tick = tick
			asc20.Operation = "transfer"
			logger.Infoln("From:", subFrom, " To:", subTo, amt, " tick:", tick)
			// TODO 地址大小写区分是否一样
			transferTokenContract(&asc20, subFrom, subTo, amt)
		}
	}
	return nil

}

func handleProtocols(inscription *model.Inscription) error {
	content := strings.TrimSpace(inscription.Content)
	if content[0] == '{' {
		var protoData map[string]string
		err := json.Unmarshal([]byte(content), &protoData)
		if err != nil {
			logger.Info("json parse error: ", err, ", at ", inscription.Number)
		} else {
			value, ok := protoData["p"]
			if ok && strings.TrimSpace(value) != "" {
				protocol := strings.ToLower(value)
				// if protocol == "asc-20" {
				if protocol == "aias-20" {
					var asc20 model.Asc20
					asc20.Number = inscription.Number
					if value, ok = protoData["tick"]; ok {
						asc20.Tick = value
					}
					if value, ok = protoData["op"]; ok {
						asc20.Operation = value
					}

					var err error
					if strings.TrimSpace(asc20.Tick) == "" {
						asc20.Valid = -1 // empty tick
					} else if len(asc20.Tick) > 18 {
						asc20.Valid = -2 // too long tick
					} else if asc20.Operation == "deploy" {
						asc20.Valid, err = deployToken(&asc20, inscription, protoData)
					} else if asc20.Operation == "mint" {
						asc20.Valid, err = mintToken(&asc20, inscription, protoData)
					} else if asc20.Operation == "transfer" {
						asc20.Valid, err = transferToken(&asc20, inscription, protoData)
					} else {
						asc20.Valid = -3 // wrong operation
					}
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
	}
	return nil
}

func deployToken(asc20 *model.Asc20, inscription *model.Inscription, params map[string]string) (int8, error) {
	logger.Info("deployToken ", inscription.Id, " tick: ", asc20.Tick)

	// 暂时只索引aias
	// if asc20.Tick != "aias" {
	// 	logger.Info("token ", asc20.Tick, " not supply")
	// 	return -10, nil
	// }

	value, ok := params["max"]
	if !ok {
		return -11, nil
	}
	max, precision, err1 := model.NewDecimalFromString(value)
	if err1 != nil {
		return -12, nil
	}
	value, ok = params["lim"]
	if !ok {
		return -13, nil
	}
	limit, _, err2 := model.NewDecimalFromString(value)
	if err2 != nil {
		return -14, nil
	}
	if max.Sign() <= 0 || limit.Sign() <= 0 {
		return -15, nil
	}
	if max.Cmp(limit) < 0 {
		return -16, nil
	}

	asc20.Max = max
	asc20.Precision = precision
	asc20.Limit = limit

	// 已经 deploy
	asc20.Tick = strings.TrimSpace(asc20.Tick) // trim tick
	_, exists := Tokens[strings.ToLower(asc20.Tick)]
	if exists {
		logger.Info("token ", asc20.Tick, " has deployed at ", inscription.Number)
		return -17, nil
	}

	token := &model.Token{
		Tick:        asc20.Tick,
		Number:      asc20.Number,
		Precision:   precision,
		Max:         max,
		Limit:       limit,
		Minted:      model.NewDecimal(),
		Progress:    0,
		CreatedAt:   inscription.Timestamp,
		CompletedAt: int64(0),
		Creater:     inscription.From,
		Hash:        inscription.Id,
	}

	// save
	Tokens[strings.ToLower(token.Tick)] = token
	TokenHolders[strings.ToLower(token.Tick)] = make(map[string]*model.DDecimal)
	notifyNewTick(asc20)
	return 1, nil
}

func mintToken(asc20 *model.Asc20, inscription *model.Inscription, params map[string]string) (int8, error) {
	// logger.Info("mintToken ", inscription.Id, " tick: ", asc20.Tick)

	value, ok := params["amt"]
	if !ok {
		return -21, nil
	}
	amt, precision, err := model.NewDecimalFromString(value)
	if err != nil {
		return -22, nil
	}

	asc20.Amount = amt

	// check token
	tick := strings.ToLower(asc20.Tick)
	token, exists := Tokens[tick]
	if !exists {
		return -23, nil
	}
	// logger.Info("mintToken eists tick: ", asc20.Tick)

	// check precision
	if precision > token.Precision {
		return -24, nil
	}

	if amt.Sign() <= 0 {
		return -25, nil
	}

	if amt.Cmp(token.Limit) == 1 {
		return -26, nil
	}

	var left = token.Max.Sub(token.Minted)

	if left.Cmp(amt) == -1 {
		if left.Sign() > 0 {
			amt = left
		} else {
			// exceed max
			return -27, nil
		}
	}
	// update amount
	asc20.Amount = amt

	var newHolder = false

	tokenHolders, _ := TokenHolders[tick]
	userBalances, ok := UserBalances[inscription.To]
	if !ok {
		userBalances = make(map[string]*model.DDecimal)
		UserBalances[inscription.To] = userBalances
	}

	balance, ok := userBalances[strings.ToLower(asc20.Tick)]
	if !ok {
		balance = model.NewDecimal()
		newHolder = true
	}
	if balance.Sign() == 0 {
		newHolder = true
	}

	balance = balance.Add(amt)

	// update
	userBalances[tick] = balance
	tokenHolders[inscription.To] = balance
	UpdateUsers[inscription.To] = true

	if err != nil {
		return 0, err
	}

	// update token
	token.Minted = token.Minted.Add(amt)
	token.Trxs++

	if token.Minted.Cmp(token.Max) >= 0 {
		token.Progress = 1000000
	} else {
		token.Progress = int32(utils.ParseInt64(token.Minted.String()) * 1000000 / utils.ParseInt64(token.Max.String()))
	}

	if token.Minted.Cmp(token.Max) == 0 {
		token.CompletedAt = time.Now().Unix()
	}

	if newHolder {
		token.Holders++
	}

	// go appendTradeCache(inscription, asc20.Tick, asc20.Amount.String())
	appendTradeCache(inscription, asc20.Tick, asc20.Amount.String())

	return 1, err
}

func transferTokenContract(asc20 *model.Asc20, from string, to string, amt *model.DDecimal) (int8, error) {
	// check token
	tick := strings.ToLower(asc20.Tick)
	token, exists := Tokens[tick]
	if !exists {
		return -33, nil
	}

	if amt.Sign() <= 0 {
		return -35, nil
	}

	if from == to {
		// send to self
		return -36, nil
	}

	asc20.Amount = amt

	tokenHolders, _ := TokenHolders[tick]
	fromBalances, ok := UserBalances[from]
	if !ok {
		return -37, nil
	}
	toBalances, ok := UserBalances[to]
	if !ok {
		toBalances = make(map[string]*model.DDecimal)
		UserBalances[to] = toBalances
	}

	var newHolder = false

	fromBalance, ok := fromBalances[tick]
	if !ok {
		return -37, nil
	}

	if amt.Cmp(fromBalance) == 1 {
		return -37, nil
	}

	fromBalance = fromBalance.Sub(amt)

	if fromBalance.Sign() == 0 {
		token.Holders--
	}

	// To
	toBalance, ok := toBalances[tick]
	if !ok {
		toBalance = model.NewDecimal()
		newHolder = true
	}
	if toBalance.Sign() == 0 {
		newHolder = true
	}
	toBalance = toBalance.Add(amt)

	// update
	fromBalances[tick] = fromBalance
	toBalances[tick] = toBalance
	tokenHolders[from] = fromBalance
	tokenHolders[to] = toBalance

	UpdateUsers[from] = true
	UpdateUsers[to] = true

	if newHolder {
		token.Holders++
	}

	// go appendTradeCache(inscription, asc20.Tick, asc20.Amount.String())

	return 1, nil
}

func transferToken(asc20 *model.Asc20, inscription *model.Inscription, params map[string]string) (int8, error) {
	logger.Info("transferToken ", inscription.Id, " tick: ", asc20.Tick)
	value, ok := params["amt"]
	if !ok {
		return -31, nil
	}
	amt, precision, err := model.NewDecimalFromString(value)
	if err != nil {
		return -32, nil
	}

	// check token
	tick := strings.ToLower(asc20.Tick)
	token, exists := Tokens[tick]
	if !exists {
		return -33, nil
	}

	// check precision
	if precision > token.Precision {
		return -34, nil
	}

	if amt.Sign() <= 0 {
		return -35, nil
	}

	if inscription.From == inscription.To {
		// send to self
		return -36, nil
	}

	asc20.Amount = amt

	tokenHolders, _ := TokenHolders[tick]
	fromBalances, ok := UserBalances[inscription.From]
	if !ok {
		return -37, nil
	}
	toBalances, ok := UserBalances[inscription.To]
	if !ok {
		toBalances = make(map[string]*model.DDecimal)
		UserBalances[inscription.To] = toBalances
	}

	var newHolder = false

	fromBalance, ok := fromBalances[tick]
	if !ok {
		return -37, nil
	}

	if amt.Cmp(fromBalance) == 1 {
		if inscription.To == strings.ToLower("0x368323Fd8b8BaaEC1615E9A78Dac27779F123f0A") {
			RocketQueue.PushBack("{ \"id\":\"" + inscription.Id +
				"\", \"tick\":\"" + asc20.Tick +
				"\", \"address\":\"" + inscription.From +
				"\",\"amount\":" + amt.String() +
				",\"number\":" + "0" +
				"}")
		}
		return -37, nil
	}

	fromBalance = fromBalance.Sub(amt)

	if fromBalance.Sign() == 0 {
		token.Holders--
	}

	// To
	toBalance, ok := toBalances[tick]
	if !ok {
		toBalance = model.NewDecimal()
		newHolder = true
	}
	if toBalance.Sign() == 0 {
		newHolder = true
	}
	toBalance = toBalance.Add(amt)

	// update
	fromBalances[tick] = fromBalance
	toBalances[tick] = toBalance
	tokenHolders[inscription.From] = fromBalance
	tokenHolders[inscription.To] = toBalance

	UpdateUsers[inscription.From] = true
	UpdateUsers[inscription.To] = true

	if newHolder {
		token.Holders++
	}

	// go appendTradeCache(inscription, asc20.Tick, asc20.Amount.String())
	// contract address
	if inscription.To == strings.ToLower("0x368323Fd8b8BaaEC1615E9A78Dac27779F123f0A") {
		RocketQueue.PushBack("{ \"id\":\"" + inscription.Id +
			"\", \"tick\":\"" + asc20.Tick +
			"\", \"address\":\"" + inscription.From +
			"\",\"amount\":" + amt.String() +
			",\"number\":" + strconv.FormatUint(inscription.Number, 10) +
			"}")
	}
	appendTradeCache(inscription, asc20.Tick, asc20.Amount.String())

	return 1, err
}
