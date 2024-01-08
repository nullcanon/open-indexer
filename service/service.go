package service

import (
	"context"
	"open-indexer/handlers"
	"open-indexer/model"
	"open-indexer/server"

	"github.com/ethereum/go-ethereum/rpc"
)

type RateParams struct {
	Symbol   string
	Currency string
}

type TickService struct{}

func NewTickService() *TickService {
	return &TickService{}
}

// Http
func (s *TickService) Ticks(symbol string, currency string) (float64, error) {
	return 2, nil
}

func (s *TickService) Balance(address string, tick string) (string, error) {
	userBalances, ok := handlers.UserBalances[address]
	if !ok {
		return "0", nil
	}

	balance, ok := userBalances[tick]
	if !ok {
		return "0", nil
	}
	return balance.String(), nil
}

// WebSocket
func (s *TickService) Newtick(ctx context.Context) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	for tick, tokens := range handlers.Tokens {
		var tickMsg model.TickMessage
		tickMsg.Limit = tokens.Limit.String()
		tickMsg.Max = tokens.Max.String()
		tickMsg.Tick = tick
		notifier.Notify(rpcSub.ID, &tickMsg)
	}

	go func() {
		tickCh := make(server.DataChannel, 1)
		server.SubHandler.Subscribe(server.NewTick, tickCh, rpcSub.ID)
		for {
			select {
			case <-rpcSub.Err(): // client send an unsubscribe request
				server.SubHandler.Unsubscribe(server.NewTick, tickCh, rpcSub.ID)
				return
			case <-notifier.Closed(): // connection dropped
				server.SubHandler.Unsubscribe(server.NewTick, tickCh, rpcSub.ID)
				return
			case message := <-tickCh:
				notifier.Notify(rpcSub.ID, message)
			}
		}
	}()
	return rpcSub, nil
}

func (s *TickService) Ping(ctx context.Context) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()
	go func() {
		tickCh := make(server.DataChannel, 1)
		server.SubHandler.Subscribe(server.Ping, tickCh, rpcSub.ID)
		for {
			select {
			case <-rpcSub.Err(): // client send an unsubscribe request
				server.SubHandler.Unsubscribe(server.Ping, tickCh, rpcSub.ID)
				return
			case <-notifier.Closed(): // connection dropped
				server.SubHandler.Unsubscribe(server.Ping, tickCh, rpcSub.ID)
				return
			case message := <-tickCh:
				notifier.Notify(rpcSub.ID, message)
			}
		}
	}()
	return rpcSub, nil
}

func (s *TickService) History(ctx context.Context) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		historyCh := make(server.DataChannel, 1)
		server.SubHandler.Subscribe(server.History, historyCh, rpcSub.ID)
		for {
			select {
			case <-rpcSub.Err(): // client send an unsubscribe request
				server.SubHandler.Unsubscribe(server.History, historyCh, rpcSub.ID)
				return
			case <-notifier.Closed(): // connection dropped
				server.SubHandler.Unsubscribe(server.History, historyCh, rpcSub.ID)
				return
			case message := <-historyCh:
				notifier.Notify(rpcSub.ID, message)
			}
		}
	}()
	return rpcSub, nil
}
