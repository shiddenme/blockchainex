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
	proxyUrl, _ := url.Parse("http://127.0.0.1:1080")
	dial := websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	log.Println(bw.wsUrl + fmt.Sprintf("%s@aggTrade", symbol))
	conn, _, err := dial.Dial(bw.wsUrl+fmt.Sprintf("%s@aggTrade", symbol), nil)
	if err != nil {
		log.Fatalln("dial fail, ", err)
		return
	}
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer conn.Close()
		for {
			select {
			case <-bw.ctx.Done():
			default:
				_, data, err := conn.ReadMessage()
				if err != nil {
					log.Fatalln("read fail, ", err)
					return
				}
				log.Println("recv: ", string(data))
			}
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-bw.ctx.Done():
			select {
			case <-time.After(10 * time.Second):
			case <-done:
			}
			return
		}
	}

}

func (bw *BinanceWebsocket) Trade(symbol string) {
	path := fmt.Sprintf("%s@trade", symbol)
	bw.Wss(path)
}

func (bw *BinanceWebsocket) Wss(path string) {

	var (
		dial websocket.Dialer
	)
	if configure.IsWall {
		proxyUrl, _ := url.Parse(configure.WallProxyAddr)
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
