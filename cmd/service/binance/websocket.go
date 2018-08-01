package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"strconv"
	"sync"
	"time"
)

var gLocker sync.Mutex

var gCondition *sync.Cond

func DepthWebsocket(ctx context.Context) (chan *DepthEvent, chan struct{}, error) {
	url := "wss://stream.binance.com:9443/ws/btcusdt"
	fmt.Println("url: ", url)
	conn, err := websocket.Dial(url, "", "http://stream.binance.com:9443")
	if err != nil {
		fmt.Println(err)
		// return
	}

	done := make(chan struct{})
	dech := make(chan *DepthEvent)
	gLocker.Lock()

	gCondition = sync.NewCond(&gLocker)

	_, err = conn.Write([]byte("hello"))

	go func(conn *websocket.Conn) {
		gLocker.Lock()
		defer gLocker.Unlock()
		defer conn.Close()

		for {
			msg := make([]byte, 128)
			select {
			case <-ctx.Done():
				fmt.Println("closing reader")
				return
			default:
				_, err := conn.Read(msg)
				if err != nil {
					fmt.Println("WsRead", err)
					return
				}

				rawDepth := struct {
					Type          string          `json:"e"`
					Time          float64         `json:"E"`
					Symbol        string          `json:"s"`
					UpdateID      int             `json:"u"`
					BidDepthDelta [][]interface{} `json:"b"`
					AskDepthDelta [][]interface{} `json:"a"`
				}{}

				if err := json.Unmarshal(msg, &rawDepth); err != nil {
					fmt.Println("Depth response unmarshal failed, ", err)
					return
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
	}(conn)
	go exitHandler(conn, done, ctx)
	return dech, done, nil
}

func exitHandler(c *websocket.Conn, done chan struct{}, ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer c.Close()
	for {
		select {
		case t := <-ticker.C:
			_, err := c.Write([]byte(t.String()))
			if err != nil {
				fmt.Println("wsWrite, ", err)
				return
			}
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
