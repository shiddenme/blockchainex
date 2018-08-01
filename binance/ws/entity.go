package ws

import (
	"context"
)

type BinanceWebsocket struct {
	wsUrl string
	ctx   context.Context
}
