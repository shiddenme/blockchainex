package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"time"
)

func DepthWebsocket1(ctx context.Context) (chan *DepthEvent, chan struct{}, error) {
	url := "wss://stream.binance.com:9443/ws/adausdt"
	url = "wss://api.fcoin.com/v2/ws"
	// url = "wss://real.okex.com:10441/websocket"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}
	fmt.Println("Connection success")
	done := make(chan struct{})

	dech := make(chan *DepthEvent)

	go func() {
		defer conn.Close()

		defer close(done)

		for {
			// err = conn.WriteMessage(websocket.TextMessage, []byte("symbol.btcusdt"))
			// fmt.Println("WriteMessage")
			// if err != nil {
			// 	fmt.Println("wsWrite: ", err)
			// 	continue
			// }
			select {
			case <-ctx.Done():
				fmt.Println("closing reader")
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					fmt.Println("wsRead: ", err)
					return
				}

				fmt.Println("Read Message: ", string(message))

				rawDepth := struct {
					Type          string          `json:"e"`
					Time          float64         `json:"E"`
					Symbol        string          `json:"s"`
					UpdateID      int             `json:"u"`
					BidDepthDelta [][]interface{} `json:"b"`
					AskDepthDelta [][]interface{} `json:"a"`
				}{}

				if err := json.Unmarshal(message, &rawDepth); err != nil {
					fmt.Println("Depth response unmarshal failed, ", err)
					continue
				}
				t := time.Unix(0, int64(rawDepth.Time)*int64(time.Millisecond))

				de := &DepthEvent{
					WSEvent: WSEvent{
						Type:   rawDepth.Type,
						Time:   t,
						Symbol: rawDepth.Symbol,
					},
					UpdateID: rawDepth.UpdateID,
				}

				for _, b := range rawDepth.BidDepthDelta {
					p, err := strconv.ParseFloat(fmt.Sprint(b[0]), 64)
					if err != nil {
						fmt.Println("wsUnmarshal ", err)
						return
					}

					q, err := strconv.ParseFloat(fmt.Sprint(b[1]), 64)
					if err != nil {
						fmt.Println("wsUnmarshal ", err)
						return
					}

					de.Bids = append(de.Bids, &Order{
						Price:    p,
						Quantity: q,
					})
				}
				dech <- de
			}
		}
	}()

	go exitHandler1(conn, done, ctx)
	return dech, done, nil
}

func exitHandler1(conn *websocket.Conn, done chan struct{}, ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer conn.Close()

	for {
		select {
		case t := <-ticker.C:
			fmt.Println("wsWrite: ", t.String())
			// err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			err := conn.WriteMessage(websocket.TextMessage, []byte("ticker.btcusdt"))
			// hq := "{'event':'addChannel','channel':'ok_sub_spot_X_ticker'}"
			// err := conn.WriteMessage(websocket.TextMessage, []byte(hq))
			if err != nil {
				fmt.Println("wsWrite: ", err)
				return
			}
			// err = conn.WriteMessage(websocket.TextMessage, []byte("symbol.btcusdt"))
			// if err != nil {
			// 	fmt.Println("wsWrite: ", err)
			// 	return
			// }

		case <-ctx.Done():
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			fmt.Println("closing connection")
			return
		}

	}
}
