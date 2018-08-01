package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func NewOkexWebsocket() *OkexWebsocket {
	ow := &OkexWebsocket{wsUrl: "wss://real.okex.com:10440/websocket/okexapi"}
	ow.message = make(chan string, 100)
	ow.subscrible = make([]*Subscrible, 0)
	return ow
}

func (ow *OkexWebsocket) Ticker() {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	proxyUrl, _ := url.Parse("http://127.0.0.1:1080")
	dial := websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}

	conn, _, err := dial.Dial(ow.wsUrl, nil)
	if err != nil {
		log.Fatalln("dial fail, ", err)
		return
	}

	defer conn.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Fatalln("read fail, ", err)
				return
			}

			log.Println("recv: ", string(data))
		}
	}()
	ticker := time.NewTicker(time.Second * 30)
	// ow.message <- `{'event':'addChannel','channel':'ok_sub_futureusd_btc_kline_this_week_1min'}`
	// ow.AddMessage(`{'event':'addChannel','channel':'ok_sub_futureusd_btc_kline_this_week_day'}`)
	go ow.AddMessage()

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"ping"}`))
			if err != nil {
				log.Fatalln("write ping fail: ", err)
				return
			}
			// ow.message <- `{'event':'addChannel','channel':'ok_sub_futureusd_btc_kline_this_week_1min'}`
		case <-done:
			return
		case <-interrupt:
			log.Println("connection closing")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Fatalln("close fail: ", err)
				return
			}
			select {
			case <-done:
			case <-time.After(30 * time.Second):
			}
		case msg := <-ow.message:
			log.Println("write msg: ", msg)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Fatalln("write msg fail: ", msg)
				return
			}
		}

	}
}
