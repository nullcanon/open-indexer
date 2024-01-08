package main

import (
	"context"
	"fmt"
	indexer "open-indexer"
	"open-indexer/config"
	"open-indexer/db"
	"open-indexer/handlers"
	"open-indexer/loader"
	"open-indexer/model"
	"open-indexer/plugin"
	"open-indexer/rpc"
	"open-indexer/structs"
	"time"
)

// var (
// 	inputfile  string
// 	outputfile string
// )

// func init() {
// 	flag.StringVar(&inputfile, "input", "./data/asc20.input.txt", "the filename of input data, default(./data/asc20.input.txt)")
// 	flag.StringVar(&outputfile, "output", "./data/asc20.output.txt", "the filename of output result, default(./data/asc20.output.txt)")

//		flag.Parse()
//	}

func main() {

	api := config.Global.Api
	handlers.Ethrpc = rpc.NewEthRPC(api)

	config.InitConfig()
	db.Setup()
	blockscan := db.BlockScan{}
	blockNumber := uint64(blockscan.GetNumber()) + 1
	// api := "https://aia-dataseed2.aiachain.org"
	// api := "https://avalanche-mainnet.infura.io/v3/5146553d78104798833c74e20e2d887b"
	// api := "https://avalanche-mainnet.infura.io/v3/5146553d78104798833c74e20e2d8"
	w := indexer.NewHttpBasedEthWatcher(context.Background(), api)

	var logger = handlers.GetLogger()
	loader.LoadDataBase()

	logger.Info("start index ", blockNumber)

	// go loader.DumpTradeCache()

	ticker1 := time.NewTicker(30 * time.Second)
	defer ticker1.Stop()

	go func(t *time.Ticker) {
		for {
			<-t.C
			handlers.DBLock.Lock()
			loader.DumpTickerInfoToDB(handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)
			loader.DumpBlockNumber()
			handlers.DBLock.Unlock()
		}
	}(ticker1)

	// we use BlockPlugin here
	w.RegisterBlockPlugin(plugin.NewSimpleBlockPlugin(func(block *structs.RemovableBlock) {
		if block.IsRemoved {
			fmt.Println("Removed >>", block.Hash(), block.Number())
		} else {
			var trxs []*model.Transaction

			// fmt.Println("Adding >>", block.Hash(), block.Number())

			for i := 0; i < len(block.GetTransactions()); i++ {
				tx := block.GetTransactions()[i]
				// fmt.Println("Adding >>", tx.GetHash(), tx.GetIndex(), tx.GetIndex(), tx.GetInput())

				var data model.Transaction

				data.Id = tx.GetHash()
				data.From = tx.GetFrom()
				data.To = tx.GetTo()
				data.Block = tx.GetBlockNumber()
				data.Idx = uint32(tx.GetIndex())
				data.Timestamp = block.Timestamp()
				data.Input = tx.GetInput()

				trxs = append(trxs, &data)
			}
			handlers.DBLock.Lock()
			err := handlers.ProcessUpdateARC20(trxs)
			if err != nil {
				logger.Fatalf("process error, %s", err)
			}
			handlers.DBLock.Unlock()

		}
	}))

	err := w.RunTillExitFromBlock(blockNumber)
	logger.Errorln(err.Error())

	// trxs, err := loader.LoadTransactionData(inputfile)
	// if err != nil {
	// 	logger.Fatalf("invalid input, %s", err)
	// }

	// err = handlers.ProcessUpdateARC20(trxs)
	// if err != nil {
	// 	logger.Fatalf("process error, %s", err)
	// }

	// print
	// loader.DumpTickerInfoMap(outputfile, handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)

}
