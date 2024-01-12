package main

import (
	"container/list"
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
	"open-indexer/server"
	"open-indexer/service"
	"open-indexer/structs"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	eth_rpc "github.com/ethereum/go-ethereum/rpc"
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

	handlers.RocketQueue = list.New()

	config.InitConfig()
	api := config.Global.Api
	var logger = handlers.GetLogger()
	handlers.Ethrpc = rpc.NewEthRPC(api, logger)

	db.Setup()
	blockscan := db.BlockScan{}
	nblockNumber, number := blockscan.GetNumber()
	blockNumber := uint64(nblockNumber) + 1
	handlers.InscriptionNumber = uint64(number)
	w := indexer.NewHttpBasedEthWatcher(context.Background(), api, logger)

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

	ticker2 := time.NewTicker(1 * time.Second)
	defer ticker2.Stop()
	go func(t *time.Ticker) {
		duration := 60 * time.Second
		time.Sleep(duration)
		for {
			<-t.C
			loader.LoadAndPushRocketMsg()
		}
	}(ticker2)

	// go func(t *time.Ticker) {
	// 	for {
	// 		<-t.C
	// 		handlers.DBLock.Lock()
	// 		loader.DumpTickerInfoToDB(handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)
	// 		loader.DumpBlockNumber()
	// 		handlers.DBLock.Unlock()
	// 	}
	// }(ticker1)

	rpcAPI := []eth_rpc.API{
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
	var err error
	handlers.ReockP, err = rocketmq.NewProducer(producer.WithNameServer([]string{"43.139.3.138:9876"}))
	if err != nil {
		logger.Infof("process NewProducer, %s", err)
	}

	if err = handlers.ReockP.Start(); err != nil {
		logger.Infof("process p.Start() error, %s", err)
	}

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
			handlers.PushRockMQ()
			loader.DumpTickerInfoToDB(handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)
			loader.DumpBlockNumber()
			handlers.NotifyHistory()

		}
	}))

	err = w.RunTillExitFromBlock(blockNumber)
	logger.Errorln(err.Error())

}
