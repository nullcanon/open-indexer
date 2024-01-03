package model

type TickMessage struct {
	Tick  string `json:"tick"`
	Max   string `json:"max"`
	Limit string `json:"lim"`
}

type HistoryMessage struct {
	Tick   string `json:"tick"`
	From   string `json:"from"`
	To     string `json:"to"`
	Hash   string `json:"hash"`
	Amount string `json:"amount"`
	Time   uint64 `json:"time"`
	Status string `json:"status"`
	Number uint64 `json:"number"`
}

type PingMessage struct {
	Ping string `json:"ping"`
}

type PongMessage struct {
	Pong string `json:"pong"`
}
