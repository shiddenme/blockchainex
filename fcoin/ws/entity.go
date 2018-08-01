package ws

import (
// "time"
)

// {"cmd":"sub","args":["ticker.btcusdt","depth.L20.btcusdt","candle.H1.btcusdt"],"id":"1"}

type Topic struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
	ID   string   `json:"id"`
}

func NewTopic(args ...string) *Topic {
	return &Topic{
		Cmd:  "sub",
		Args: args,
		ID:   "1",
	}
}

type TopicResult struct {
	ID     string   `json:"id"`
	Type   string   `json:"type"`
	Topics []string `json:"topics"`
}

type FCoinWebSocketResult struct {
	Type string `json:"type"`
	Seq  int64  `json:"seq"`
	Date string `json:"date,omitempty"`
}

func (f *FCoinWebSocketResult) SetDate() {
	// f.Date = time.Unix(f.Seq, 0).Format("2006-01-02 13:04:05")
}

type Ticker struct {
	Ticker []float64 `json:"ticker"`
	FCoinWebSocketResult
}

type Depth struct {
	Bids []float64 `json:"bids"`
	Asks []float64 `json:"asks"`
	TS   int64     `json:"ts"`
	FCoinWebSocketResult
}

type Candle struct {
	Open     float64 `json:"open"`
	Close    float64 `json:"close"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	QuoteVol float64 `json:"quote_vol"`
	ID       int64   `json:"id"`
	Count    int64   `json:"count"`
	BaseVol  float64 `json:"base_vol"`
	FCoinWebSocketResult
}
