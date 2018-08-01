package example

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func ClientExample(msg string) {

	time.Sleep(2 * time.Second)

	wsUrl := url.URL{Scheme: "ws", Host: "127.0.0.1:9001", Path: "/example/ws"}
	wsUrl = url.URL{Scheme: "wss", Host: "api.fcoin.com", Path: "/v2/ws"}
	wsUrl = url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/BNBBTC"}
	fmt.Println("websocket url: ", wsUrl.String())

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)

	if err != nil {
		log.Println("dial fail, ", err)
		return
	}
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	defer conn.Close()
	done := make(chan struct{})
	// 接收消息
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read Faild, ", err)
				return
			}

			log.Printf("client recv: %s", string(data))
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	// 发送消息
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			log.Println("client send message: ", msg)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("Write fail, ", err)
				return
			}

		case <-interrupt:
			log.Println("interrupt")

			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close: ", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
		}

	}
}

func FCoinClient(msg string) {
	time.Sleep(1 * time.Second)
	fcUrl := url.URL{Scheme: "wss", Host: "api.fcoin.com", Path: "/v2/ws"}
	log.Println("URL: ", fcUrl.String())
	conn, _, err := websocket.DefaultDialer.Dial(fcUrl.String(), nil)
	if err != nil {
		log.Println("Dial fail, ", err)
		return
	}

	defer conn.Close()
	done := make(chan struct{})
	topic := make(chan string)

	// 接收消息
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Println("read fail, ", err)
				return
			}
			log.Println("recv: ", string(data))
		}
	}()
	// go func() { topic <- msg }()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	// 发起订阅
	for {
		select {
		case <-ticker.C:
			log.Println("client write: ", msg)
			err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("write fail, ", err)
				return
			}
		case <-done:
			return
		case m := <-topic:
			log.Println("topic: ", m)
			err = conn.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Println("write fail, ", err)
				return
			}
		}
	}
}

func BinanceClient(ctx context.Context) {
	time.Sleep(2 * time.Second)
	wsUrl := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@depth"}
	// log.Println(fmt.Printf("wss://stream.binance.com:9443/ws/%s@depth", "btcusdt"))
	fmt.Println("websocket url: ", wsUrl.String())

	// websocket proxy
	uProxy, _ := url.Parse("http://127.0.0.1:1080")
	dialer := websocket.Dialer{
		Proxy: http.ProxyURL(uProxy),
	}
	conn, _, err := dialer.Dial(wsUrl.String(), nil)

	// conn, _, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)
	if err != nil {
		log.Fatalln("dial fail, ", err)
		return
	}

	done := make(chan struct{})
	go func() {
		defer conn.Close()
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				log.Println("closing reader")
				return
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

	ticker := time.NewTicker(time.Minute * 2)
	defer ticker.Stop()
	defer conn.Close()
	for {
		select {
		case t := <-ticker.C:
			log.Println("send message: ", t.String())
			err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Fatalln("write fatal:", err)
				return
			}
		case <-ctx.Done():
			select {

			case <-done:
			case <-time.After(time.Second * 10):
			}

			log.Println("connection closing")
			return
		}
	}
}
