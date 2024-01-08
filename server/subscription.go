package server

import (
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
)

type DataChannel chan interface{}
type DataChannelMap map[rpc.ID]DataChannel

type SubEventType int32

const (
	NewTick SubEventType = 0
	History SubEventType = 1
	Ping    SubEventType = 2
)

type Subscriber struct {
	subscribers map[SubEventType]DataChannelMap
	mutexSub    sync.Mutex
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		mutexSub:    sync.Mutex{},
		subscribers: make(map[SubEventType]DataChannelMap),
	}
}

func (d *Subscriber) Subscribe(topic SubEventType, ch DataChannel, id rpc.ID) {
	d.mutexSub.Lock()
	defer d.mutexSub.Unlock()

	if _, exist := d.subscribers[topic]; exist {
		d.subscribers[topic][id] = ch
	} else {
		c := make(DataChannelMap)
		c[id] = ch
		d.subscribers[topic] = c
	}
}

func (d *Subscriber) Publish(topic SubEventType, data interface{}) {
	d.mutexSub.Lock()
	defer d.mutexSub.Unlock()

	if chans, found := d.subscribers[topic]; found {
		go func(data interface{}, dataChannelMap DataChannelMap) {
			for _, ch := range dataChannelMap {
				ch <- data
			}
		}(data, chans)
	}
}

func (d *Subscriber) Unsubscribe(topic SubEventType, ch DataChannel, id rpc.ID) {
	d.mutexSub.Lock()
	defer d.mutexSub.Unlock()

	close(ch)
	delete(d.subscribers[topic], id)
}
