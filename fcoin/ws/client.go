package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

var wsfCoin *WSFCoin

type WSFCoin struct {
	wsUrl   url.URL
	Ticker  *Ticker
	Depth   *Depth
	Candle  *Candle
	Updated int64 // 最后更新时间
}

func GetWSFCoin() *WSFCoin {
	return NewWSFCoin("api.fcoin.com", "/v2/ws")
}

func NewWSFCoin(host, path string) *WSFCoin {
	if wsfCoin == nil {
		wsfCoin = &WSFCoin{
			wsUrl:  url.URL{Scheme: "wss", Host: host, Path: path},
			Ticker: &Ticker{},
			Depth:  &Depth{},
			Candle: &Candle{},
		}
	}
	return wsfCoin
}

func Start() {
	time.Sleep(2 * time.Second)
	// "ticker.btcusdt", "depth.L20.btcusdt", "candle.H1.btcusdt"
	topic := &Topic{ID: "1", Cmd: "sub", Args: []string{"ticker.btcusdt", "depth.L20.btcusdt", "candle.H1.btcusdt"}}
	wsfcoin := NewWSFCoin("api.fcoin.com", "/v2/ws")
	wsfcoin.FCoinClient(topic)
}

func (f *WSFCoin) FCoinClient(topic *Topic) {
	// fcoinUrl := url.URL{Scheme: "wss", Host: "api.fcoin.com", Path: "/v2/ws"}
	interrupt := make(chan os.Signal)

	signal.Notify(interrupt, os.Interrupt)
	conn, _, err := websocket.DefaultDialer.Dial(f.wsUrl.String(), nil)
	if err != nil {
		log.Fatalln("Dial fail, ", err)
		return
	}
	defer conn.Close()
	done := make(chan struct{})

	// 接收消息
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Fatalln("read fail, ", err)
			}
			log.Println("recv: ", string(data))
			f.HandleRecv(data)
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			data, _ := json.Marshal(topic)
			log.Println("write message: ", string(data))
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Fatalln("Write fail, ", err)
				return
			}
		case <-interrupt:
			log.Println("close connection")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Fatalln("close write fail, ", err)
				return
			}
			select {
			case <-done:
			case <-time.After(1 * time.Second):
			}
		}
	}
}

func (f *WSFCoin) HandleRecv(data []byte) error {
	var (
		result FCoinWebSocketResult
	)
	err := json.Unmarshal(data, &result)
	if err != nil {
		log.Println("recv unmarshal fail, ", err)
		return err
	}
	if result.Type == "" {

		return fmt.Errorf("未知的返回")
	}
	if result.Type == "topics" { // 订阅成功

	} else {
		tp := strings.Split(result.Type, ".")
		var r interface{}
		// log.Println("type1: ", tp[0])
		switch tp[0] {
		case "candle":
			r = f.Candle
		case "ticker":
			r = f.Ticker
			// err = json.Unmarshal(data, &f.Ticker)
			// if err != nil {
			// 	log.Println("recv unmarshal fail, ", err)
			// 	return err
			// }
		case "depth":
			r = f.Depth
		}
		err = json.Unmarshal(data, &r)
		if err != nil {
			log.Println("recv unmarshal fail, ", err)
			return err
		}
		f.Updated = time.Now().Unix()
	}
	return nil
}
