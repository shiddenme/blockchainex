package okex

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func Ticker() {
	wsUrl := "wss://real.okex.com:10440/websocket/okexapi"
	log.Println(wsUrl)
	interrupt := make(chan os.Signal)

	signal.Notify(interrupt, os.Interrupt)

	proxyUrl, _ := url.Parse("http://127.0.0.1:1080")
	msg := "{'event':'addChannel','channel':'ok_sub_futureusd_btc_ticker_this_week'}"
	// message := make(chan string)
	// message <- "{'event':'addChannel','channel':'ok_sub_futureusd_X_ticker_Y'}"

	// websocket proxy
	dial := websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	done := make(chan struct{})
	conn, _, err := dial.Dial(wsUrl, nil)

	if err != nil {
		log.Fatalln("dial fail, ", err)
		return
	}

	defer conn.Close()

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

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		// case msg := <-message:
		// 	log.Println("write message: ", msg)
		// 	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		// 	if err != nil {
		// 		log.Fatalln("write fail, ", err)
		// 		return
		// 	}
		case <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"ping"}`))
			if err != nil {
				log.Fatalln("write fail, ", err)
				return
			}
			log.Println("write message: ", msg)
			err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Fatalln("write fail, ", err)
				return
			}
		case <-done:
			return
		case <-interrupt:
			log.Println("connection interrupt")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Fatalln("interrupt fail, ", err)
				return
			}
			select {
			case <-done:
			case <-time.After(1 * time.Second):
			}
		}
	}
}
