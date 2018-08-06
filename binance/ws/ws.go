package ws

import (
	"blockchainex/configure"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

func NewBinanceWebsocket(ctx context.Context) *BinanceWebsocket {
	return &BinanceWebsocket{wsUrl: "wss://stream.binance.com:9443/ws/", ctx: ctx}
}

func (bw *BinanceWebsocket) AggTrade(symbol string) {
	path := fmt.Sprintf("%s@aggTrade", symbol)
	bw.Wss(path)
}

func (bw *BinanceWebsocket) Trade(symbol string) {
	path := fmt.Sprintf("%s@trade", symbol)
	bw.Wss(path)
}

func (bw *BinanceWebsocket) KLineCandleStick(symbol string) {
	klineList := []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d", "3d", "1w", "1M"}
	for _, kline := range klineList {
		path := fmt.Sprintf("%s@kline_%s", symbol, kline)
		bw.Wss(path)
	}
}

func (bw *BinanceWebsocket) MiniTicker(symbol string) {
	path := fmt.Sprintf("%s@miniTicker", symbol)
	bw.Wss(path)
}

func (bw *BinanceWebsocket) Ticker(symbol string) {
	path := fmt.Sprintf("%s@ticker", symbol)
	bw.Wss(path)
}

func (bw *BinanceWebsocket) Depth(symbol string) {
	depthList := []int{5, 10, 20}
	for _, depth := range depthList {
		path := fmt.Sprintf("%s@depth%d", symbol, depth)
		bw.Wss(path)
	}
}

func (bw *BinanceWebsocket) DiffDepth(symbol string) {
	path := fmt.Sprintf("%s@depth", symbol)
	bw.Wss(path)
}

func (bw *BinanceWebsocket) Wss(path string) {

	var (
		dial websocket.Dialer
	)
	if configure.IsWall {
		proxyUrl, _ := url.Parse(configure.WallProxyAddr)
		// websocket proxy
		dial = websocket.Dialer{
			Proxy: http.ProxyURL(proxyUrl),
		}
	} else {
		dial = websocket.Dialer{}
	}

	conn, _, err := dial.Dial(bw.wsUrl+path, nil)
	if err != nil {
		log.Fatalln("dial ", path, " fail, ", err)
		return
	}
	done := make(chan struct{})
	defer conn.Close()
	// read message
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Fatalln("read ", path, " fail, ", err)
			}
			log.Println(path, " recv: ", string(data))
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-done:
			select {
			case <-time.After(5 * time.Second):
			case <-done:
			}
		}
	}
}
