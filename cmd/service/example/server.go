package example

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"net/http"
	"sync"
	"time"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsMessage struct {
	MessageType int
	data        []byte
}

type wsConnection struct {
	conn      *websocket.Conn
	rChan     chan *wsMessage
	wChan     chan *wsMessage
	mutex     sync.Mutex
	isClose   bool
	closeChan chan byte
}

func (ws *wsConnection) wsReadLoop() {
	for {
		mt, data, err := ws.conn.ReadMessage()
		if err != nil {
			goto error
		}

		req := &wsMessage{
			MessageType: mt,
			data:        data,
		}

		select {
		case ws.rChan <- req:
		case <-ws.closeChan:
			goto closed
		}
	}
error:
	ws.wsClose()
closed:
}

func (ws *wsConnection) wsWriteLoop() {
	for {
		select {
		case msg := <-ws.rChan:
			if err := ws.conn.WriteMessage(msg.MessageType, msg.data); err != nil {
				goto error
			}
		case <-ws.closeChan:
			goto closed
		}
	}
error:
	ws.wsClose()
closed:
}

func (ws *wsConnection) procLoop() {
	go func() {
		for {
			time.Sleep(time.Second * 2)
			if err := ws.conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
				fmt.Println("hello faild")
				ws.wsClose()
				break
			}
		}
	}()

	for {
		msg, err := ws.wsRead()
		if err != nil {
			fmt.Println("read fail, ", err)
			break
		}

		fmt.Println("server read: ", string(msg.data))

		err = ws.wsWrite(msg.MessageType, msg.data)
		if err != nil {
			fmt.Println("write fail, ", err)
			break
		}
	}
}

func (ws *wsConnection) wsRead() (*wsMessage, error) {
	select {
	case msg := <-ws.rChan:
		return msg, nil
	case <-ws.closeChan:

	}
	return nil, errors.New("websocket close")
}

func (ws *wsConnection) wsWrite(msgType int, data []byte) error {
	select {
	case ws.wChan <- &wsMessage{data: data, MessageType: msgType}:
	case <-ws.closeChan:
		return errors.New("websocket close")
	}
	return nil
}

func (ws *wsConnection) wsClose() {
	ws.conn.Close()
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !ws.isClose {
		ws.isClose = true
		close(ws.closeChan)
	}
}

func WsHandler(ctx echo.Context) error {
	wsSocket, err := wsUpgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}

	connection := &wsConnection{
		conn:      wsSocket,
		wChan:     make(chan *wsMessage, 1000),
		rChan:     make(chan *wsMessage, 1000),
		closeChan: make(chan byte),
		isClose:   false,
	}

	go connection.procLoop()
	go connection.wsReadLoop()
	go connection.wsWriteLoop()
	return nil
}
