package main

import (
	"flag"
	"open-indexer/handlers"
	"open-indexer/loader"
	"context"
	"fmt"
	"open-indexer/plugin"
	"open-indexer/structs"
	"open-indexer/model"
	indexer "open-indexer"
	"open-indexer/db"
	"open-indexer/config"

	
)

var (
	inputfile  string
	outputfile string
)

func init() {
	flag.StringVar(&inputfile, "input", "./data/asc20.input.txt", "the filename of input data, default(./data/asc20.input.txt)")
	flag.StringVar(&outputfile, "output", "./data/asc20.output.txt", "the filename of output result, default(./data/asc20.output.txt)")

	flag.Parse()
}

func main() {

	config.InitConfig()
	db.Setup()

	api := "https://aia-dataseed2.aiachain.org"
	w := indexer.NewHttpBasedEthWatcher(context.Background(), api)

	var logger = handlers.GetLogger()
	loader.LoadDataBase()

	logger.Info("start index")
	// we use BlockPlugin here
	w.RegisterBlockPlugin(plugin.NewSimpleBlockPlugin(func(block *structs.RemovableBlock) {
		if block.IsRemoved {
			fmt.Println("Removed >>", block.Hash(), block.Number())
		} else {
			var trxs []*model.Transaction

			fmt.Println("Adding >>", block.Hash(), block.Number())

			for i := 0; i < len(block.GetTransactions()); i++ {
				tx := block.GetTransactions()[i]
				fmt.Println("Adding >>", tx.GetHash(), tx.GetIndex(),tx.GetIndex(),tx.GetInput())

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
			err := handlers.ProcessUpdateARC20(trxs)
			if err != nil {
				logger.Fatalf("process error, %s", err)
			}
		}

		loader.DumpTickerInfoToDB(handlers.BlockNumber, handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)

	}))

	w.RunTillExitFromBlock(25338054)



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
