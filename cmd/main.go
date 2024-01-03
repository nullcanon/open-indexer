package main

import (
	"context"
	"flag"
	"fmt"
	indexer "open-indexer"
	"open-indexer/config"
	"open-indexer/db"
	"open-indexer/handlers"
	"open-indexer/loader"
	"open-indexer/model"
	"open-indexer/plugin"
	"open-indexer/server"
	"open-indexer/service"
	"open-indexer/structs"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
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
	blockscan := db.BlockScan{}
	nblockNumber, number := blockscan.GetNumber()
	blockNumber := uint64(nblockNumber) + 1
	handlers.InscriptionNumber = uint64(number)
	api := config.Global.Api
	w := indexer.NewHttpBasedEthWatcher(context.Background(), api)

	var logger = handlers.GetLogger()
	loader.LoadDataBase()

	logger.Info("start index ", blockNumber)

	// go loader.DumpTradeCache()

	ticker1 := time.NewTicker(30 * time.Second)
	defer ticker1.Stop()
	go func(t *time.Ticker) {
		duration := 60 * time.Second
		time.Sleep(duration)
		for {
			<-t.C
			var pong model.PongMessage
			pong.Pong = "pong"
			server.SubHandler.Publish(server.Ping, &pong)
		}
	}(ticker1)

	// go func(t *time.Ticker) {
	// 	for {
	// 		<-t.C
	// 		handlers.DBLock.Lock()
	// 		loader.DumpTickerInfoToDB(handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)
	// 		loader.DumpBlockNumber()
	// 		handlers.DBLock.Unlock()
	// 	}
	// }(ticker1)

	rpcAPI := []rpc.API{
		{
			Namespace: "tick",
			Version:   "1.0",
			Service:   service.NewTickService(),
			Public:    true,
		},
	}

	httpSrv := server.NewServer(rpcAPI, logger, server.HTTP)
	httpSrv.Start()
	defer httpSrv.Stop()

	wsSrv := server.NewServer(rpcAPI, logger, server.WS)
	wsSrv.Start()
	defer wsSrv.Stop()

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
			err := handlers.ProcessUpdateARC20(trxs)
			if err != nil {
				logger.Fatalf("process error, %s", err)
			}
			loader.DumpTickerInfoToDB(handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)
			loader.DumpBlockNumber()
			handlers.NotifyHistory()

		}
	}))

	err := w.RunTillExitFromBlock(blockNumber)
	logger.Errorln(err.Error())

}
