package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"open-indexer/handlers"
	"open-indexer/model"
	"open-indexer/server"

	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"github.com/ubiq/go-ubiq/common/hexutil"
	crypto1 "github.com/ubiq/go-ubiq/crypto"
	"github.com/ubiq/go-ubiq/crypto/secp256k1"
	"golang.org/x/crypto/sha3"
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
func (s *TickService) Sign(creater string, buyer string, amount string, price string, fee string, exTime string, id string) (string, error) {

	creater = common.HexToAddress(creater).Hex()
	buyer = common.HexToAddress(buyer).Hex()
	// key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	// if err != nil {
	// 	panic(err)
	// }
	//0x2BB32bBE01840DEb3d2512C74c4fdA7c1992993D
	// input pirvate key
	pkeyb, err := hex.DecodeString("ea0415ba66565243136a0e698cf194d1bc720bd97b026de2d81185bc511348f6")
	if err != nil {
		// s.logger.Fatalln(err)
		return "", err
	}

	// message := "TEST"
	// // Turn the message into a 32-byte hash
	// hash := solsha3.SoliditySHA3(solsha3.String(message))

	// types := []string{"address", "bytes1", "uint8[]", "bytes32", "uint256", "address[]", "uint32"}
	// values := []interface{}{
	//     "0x935F7770265D0797B621c49A5215849c333Cc3ce",
	//     "0xa",
	//     []uint8{128, 255},
	//     "0x4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45",
	//     "100000000000000000",
	//     []string{
	//         "0x13D94859b23AF5F610aEfC2Ae5254D4D7E3F191a",
	//         "0x473029549e9d898142a169d7234c59068EDcBB33",
	//     },
	//     123456789,
	// }
	// ot, err := strconv.Atoi(orderType)
	// if err != nil {
	// 	return "", err
	// }

	types := []string{"address", "address", "uint256", "uint256", "uint256", "uint256", "bytes32"}
	values := []interface{}{
		creater,
		buyer,
		amount,
		price,
		fee,
		exTime,
		id,
	}

	hash := solsha3.SoliditySHA3(types, values)

	types = []string{"string", "bytes32"}
	values = []interface{}{
		"\x19Ethereum Signed Message:\n32",
		hash,
	}
	// Prefix and then hash to mimic behavior of eth_sign
	prefixed := solsha3.SoliditySHA3(types, values)
	// sig, err := secp256k1.Sign(prefixed, math.PaddedBigBytes(key.D, 32))
	sig, err := secp256k1.Sign(prefixed, pkeyb)
	if err != nil {
		// panic(err)
		return "", err
	}
	// s.logger.Printf("Process sign : %s", hex.EncodeToString(sig))
	return "0x" + hex.EncodeToString(sig), nil
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

func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

func TextAndHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func (s *TickService) RecoverPubkey(msg string, sig string) (string, error) {
	// s.logger.Printf("Process RecoverPubkey : %s, %s", msg, sig)

	hexdata, _ := hexutil.Decode(msg)

	hash, _ := TextAndHash(hexdata)

	// raw_hash := crypto1.Keccak256(prefixed)

	if len(sig) != 132 && !has0xPrefix(sig) {
		// s.logger.Fatalln("sig length error : ", sig)
		return "", errors.New("sig length error, must length  130 and 0x prefix")
	}

	hexsig, err := hex.DecodeString(sig[2:])
	if err != nil {
		// s.logger.Fatalln(err)
		return "", err
	}
	hexsig[64] -= 27
	// pksource, err := secp256k1.RecoverPubkey(signHash(raw_hash), hexsig);
	// if err != nil {
	// 	s.logger.Printf("Process RecoverPubkey error : %s, %s", msg, sig)
	// 	return "", err;
	// }
	// address := crypto.PubkeyToAddress(pksource).Hex()
	// pubkey, _ := crypto1.SigToPub(signHash(raw_hash), hexsig)
	pubkey, _ := crypto1.SigToPub(hash, hexsig)

	address := crypto1.PubkeyToAddress(*pubkey)
	return address.String(), nil
}
