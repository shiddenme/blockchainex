package ws

import (
	"blockchainex/configure"
	"compress/gzip"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func NewHuobiWebsocket() *HuobiWebsocket {
	hw := &HuobiWebsocket{
		wsUrl: "wss://api.huobi.pro/ws",
	}
	hw.pingChan = make(chan int64, 2)
	hw.topicMessage = make(chan string, 10)
	return hw
}

func (hw *HuobiWebsocket) Topic() {
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
	conn, _, err := dial.Dial(hw.wsUrl, nil)
	if err != nil {
		log.Fatalln("dial fail, ", err)
		return
	}
	done := make(chan struct{})
	defer conn.Close()
	go func() {
		defer close(done)
		for {
			_, r, err := conn.NextReader()
			if err != nil {
				log.Fatalln("read fail, ", err)
				return
			}

			gzipReader, err := gzip.NewReader(r)
			if err != nil {
				log.Fatalln("gzip read fail, ", err)
				return
			}
			data, err := ioutil.ReadAll(gzipReader)
			if err != nil {
				log.Fatalln("read fail, ", err)
				return
			}

			log.Println("recv: ", string(data))
			ping := Ping{}
			err = json.Unmarshal(data, &ping)
			if err != nil {
				log.Fatalln("recv unmarshal fail, ", err)
				return
			}

			if ping.Ping > 0 {
				hw.Lock()
				hw.pingChan <- ping.Ping
				hw.Unlock()
			}
		}
	}()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case t := <-hw.pingChan:
			// response websocket server
			if t > 0 {
				pong := make(map[string]int64)
				pong["pong"] = t
				data, _ := json.Marshal(pong)
				err := conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Fatalln("write pong fail, ", err)
					return
				}
			}
		case msg := <-hw.topicMessage:
			// log.Println("write msg: ", msg)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Fatalln("write ", msg, " fail, ", err)
				return
			}
		case <-done:
			select {
			case <-time.After(5 * time.Second):
			case <-done:
			}
		}
	}
}
