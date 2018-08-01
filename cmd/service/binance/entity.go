package binance

import (
	"context"
	"time"
)

type Order struct {
	Price    float64
	Quantity float64
}

type OrderBook struct {
	LastUpdateID int `json:"lastUpdateId"`
	Bids         []*Order
	Asks         []*Order
}

type WSEvent struct {
	Type   string
	Time   time.Time
	Symbol string
}

type DepthEvent struct {
	WSEvent
	UpdateID int
	OrderBook
}

type BinanceWebsocket struct {
	wsUrl string
	ctx   context.Context
}
