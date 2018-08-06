package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func Client() {
	wsUrl := "ws://127.0.0.1:9003/ws/server"
	log.Println("wsUrl: ", wsUrl)
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatalln("dial fail: ", err)
		return
	}

	done := make(chan struct{})
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

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			// log.Println("write: ", t.String())
			err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Fatalln("write fail, ", err)
				return
			}
		case <-done:
			select {
			case <-time.After(time.Second):
			case <-done:
			}
		}
	}
}

func Client2() {
	wsUrl := "ws://127.0.0.1:9003/ws/server"
	log.Println("wsUrl: ", wsUrl)
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatalln("dial fail: ", err)
		return
	}

	done := make(chan struct{})
	defer conn.Close()

	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Fatalln("read fail, ", err)
				return
			}
			log.Println("client2 recv: ", string(data))
		}
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// log.Println("write: ", t.String())
			err := conn.WriteMessage(websocket.TextMessage, []byte("I'm Client2"))
			if err != nil {
				log.Fatalln("write fail, ", err)
				return
			}
		case <-done:
			select {
			case <-time.After(time.Second):
			case <-done:
			}
		}
	}
}
