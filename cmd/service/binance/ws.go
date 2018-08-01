package binance

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

func NewBinanceWebsocket(ctx context.Context) *BinanceWebsocket {
	return &BinanceWebsocket{ctx: ctx, wsUrl: "wss://stream.binance.com:9443/ws/"}
}

func (bw *BinanceWebsocket) Depth(symbol string) chan struct{} {
	wsUrl := bw.wsUrl + fmt.Sprintf("%s@depth", symbol)

	proxyUrl, _ := url.Parse("http://127.0.0.1:1080")

	// websocket proxy
	dial := websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	done := make(chan struct{})
	conn, _, err := dial.Dial(wsUrl, nil)

	if err != nil {
		log.Fatalln("dial fail: ", err)
		return done
	}

	go func() {
		defer close(done)
		defer conn.Close()
		for {
			select {
			case <-bw.ctx.Done():
			default:
				_, data, err := conn.ReadMessage()
				if err != nil {
					log.Fatalln("read message fail: ", err)
					return
				}
				log.Println("recv: ", string(data))
			}
		}
	}()
	go bw.exitHandler(conn, done)
	return done
}

func (bw *BinanceWebsocket) exitHandler(conn *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer conn.Close()
	for {
		select {
		case <-ticker.C:
			// err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			// if err != nil {
			// 	log.Fatalln("write fail, ", err)
			// 	return
			// }

		case <-bw.ctx.Done():
			select {
			case <-done:
			case <-time.After(5 * time.Second):
			}
			log.Println("connection closing.")
			return
		}
	}
}
