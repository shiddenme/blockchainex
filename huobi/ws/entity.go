package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var (
	PeriodList = []string{"1min", "5min", "15min", "30min", "60min", "4hour", "1day", "1mon", "1week", "1year"}
	TypeList   = []string{"step0", "step1", "step2", "step3", "step4", "step5", "percent10"}
	SymbolList = []string{"btcusdt", "ethbtc", "etcbtc", "bchbtc", "ltcbtc"}
)

type Ping struct {
	Ping int64 `json:"ping"`
}

type HuobiWebsocket struct {
	sync.RWMutex
	wsUrl        string
	pingChan     chan int64
	topicMessage chan string
	Subscrible   []*Subscrible
}

func (hw *HuobiWebsocket) AddMsg() {
	// hw.Lock()
	// hw.topicMessage <- "{market.tickers"
	// hw.Unlock()
	// hw.Subscrible = append(hw.Subscrible, &Subscrible{Req: "market.tickers", ID: "id10"})
	// for _, sub := range hw.Subscrible {
	// 	hw.Lock()
	// 	hw.topicMessage <- sub.Marshal()
	// 	hw.Unlock()
	// 	time.Sleep(1 * time.Second)
	// }
	i := 0
	second := 1
	go func() {
		for {
			if len(hw.Subscrible) > 0 {
				sub := hw.Subscrible[i%len(hw.Subscrible)]
				hw.Lock()
				hw.topicMessage <- sub.Marshal()
				hw.Unlock()
				i++
				second = 1
			} else {
				second = 10
			}
			time.Sleep(time.Duration(second) * time.Second)
		}
	}()
}

type Subscrible struct {
	Req   string `json:"req,omitempty"`
	ID    string `json:"id,omitempty"`
	Unsub string `json:"unsub,omitempty"`
}

func (s *Subscrible) Marshal() string {
	data, _ := json.Marshal(s)
	return string(data)
}

func (hw *HuobiWebsocket) AddMarketKlineMsg() {
	for _, symbol := range SymbolList {
		for _, period := range PeriodList {
			hw.Subscrible = append(hw.Subscrible, &Subscrible{Req: fmt.Sprintf("market.%s.kline.%s", symbol, period), ID: ""})
		}
	}
}

func (hw *HuobiWebsocket) AddMarketDepthMsg() {
	for _, symbol := range SymbolList {
		for _, t := range TypeList {
			hw.Subscrible = append(hw.Subscrible, &Subscrible{Req: fmt.Sprintf("market.%s.depth.%s", symbol, t), ID: ""})
		}
	}
}

func (hw *HuobiWebsocket) AddMarketTradeMsg() {
	for _, symbol := range SymbolList {
		hw.Subscrible = append(hw.Subscrible, &Subscrible{Req: fmt.Sprintf("market.%s.trade.detail", symbol), ID: ""})
	}
}

func (hw *HuobiWebsocket) AddMarketDetailMsg() {
	for _, symbol := range SymbolList {
		hw.Subscrible = append(hw.Subscrible, &Subscrible{Req: fmt.Sprintf("market.%s.detail", symbol), ID: ""})
	}
}
