package model

type TickMessage struct {
	Tick  string `json:"ticks"`
	Max   string `json:"max"`
	Limit string `json:"lim"`
}

type HistoryMessage struct {
	Tick   string `json:"ticks"`
	From   string `json:"fromAddress"`
	To     string `json:"toAddress"`
	Hash   string `json:"hash"`
	Amount string `json:"amount"`
	Time   uint64 `json:"time"`
	Status string `json:"status"`
	Number uint64 `json:"number"`
	Method string `json:"method"`
}

type PingMessage struct {
	Ping string `json:"ping"`
}

type PongMessage struct {
	Pong string `json:"pong"`
}
