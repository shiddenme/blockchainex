package ws

import (
	"blockchainex/configure"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var (
	// btc, ltc, eth, etc, bch,eos,xrp,btg
	BTC = "btc"
	LTC = "ltc"
	ETH = "eth"
	ETC = "etc"
	BCH = "bch"
	EOS = "eos"
	XRP = "xrp"
	BTG = "btg"

	CoinList = []string{BTC, LTC, ETH, ETC, EOS, XRP, BTC, BCH}

	ThisWeek = "this_week"
	NextWeek = "next_week"
	Quarter  = "quarter"

	CycleList = []string{ThisWeek, NextWeek, Quarter}

	DateList = []string{"1min", "3min", "5min", "15min", "30min", "1hour", "2hour", "4hour", "6hour", "12hour", "day", "3day", "week"}

	DepthList = []string{"5", "10", "20"}
)

type OkexWebsocket struct {
	sync.RWMutex
	wsUrl      string
	ctx        context.Context
	message    chan string
	subscrible []*Subscrible
}

type Subscrible struct {
	Channel    string            `json:"channel,omitempty"`
	Event      string            `json:"event"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

func (ow *OkexWebsocket) AddMessage() {
	// ow.message <- message
	for _, sub := range ow.subscrible {
		data, err := json.Marshal(sub)
		if err != nil {
			continue
		}
		// log.Println("add message: ", string(data))
		ow.message <- string(data)
	}
}

// SubFutureusdTicker 订阅行情数据
func (ow *OkexWebsocket) SubFutureusdTicker() {
	// fmt.Println(CoinList)
	// fmt.Println(CycleList)
	for _, coin := range CoinList {
		for _, cycle := range CycleList {
			// log.Printf("ok_sub_futureusd_%s_ticker_%s", coin, cycle)
			ow.Lock()
			ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("ok_sub_futureusd_%s_ticker_%s", coin, cycle)})
			ow.Unlock()
		}
	}
}

// SubFutureusdKLine 订阅K线数据
func (ow *OkexWebsocket) SubFutureusdKLine() {
	for _, coin := range CoinList {
		for _, cycle := range CycleList {
			for _, date := range DateList {
				ow.Lock()
				ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("ok_sub_futureusd_%s_kline_%s_%s", coin, cycle, date)})
				ow.Unlock()
			}
		}
	}
}

// SubFutureusdDepth 订阅合约市场深度(200增量数据返回)
func (ow *OkexWebsocket) SubFutureusdDepth() {
	for _, coin := range CoinList {
		for _, cycle := range CycleList {
			ow.Lock()
			ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("ok_sub_futureusd_%s_depth_%s", coin, cycle)})
			ow.Unlock()
			for _, depth := range DepthList {
				ow.Lock()
				ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("ok_sub_futureusd_%s_depth_%s_%s", coin, cycle, depth)})
				ow.Unlock()
			}
		}
	}
}

// SubFutureusdTrade 订阅合约交易信息
func (ow *OkexWebsocket) SubFutureusdTrade() {
	for _, coin := range CoinList {
		for _, cycle := range CycleList {
			ow.Lock()
			ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("ok_sub_futureusd_%s_trade_%s", coin, cycle)})
			ow.Unlock()
		}
	}
}

// SubFutureusdIndex 订阅合约指数
func (ow *OkexWebsocket) SubFutureusdIndex() {
	for _, coin := range CoinList {
		ow.Lock()
		ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("ok_sub_futureusd_%s_index", coin)})
		ow.Unlock()
	}
}

// ForecastPrice 合约预估交割价格
func (ow *OkexWebsocket) ForecastPrice() {
	for _, coin := range CoinList {
		ow.Lock()
		ow.subscrible = append(ow.subscrible, &Subscrible{Event: "addChannel", Channel: fmt.Sprintf("%s_forecast_price", coin)})
		ow.Unlock()
	}
}

func (ow *OkexWebsocket) Login() {
	var (
		subLogin Subscrible
	)
	params := url.Values{}
	params.Set("api_key", configure.OkApiKey)
	// params.Set("event", "login")
	subLogin.Parameters = make(map[string]string, 0)
	subLogin.Parameters["api_key"] = configure.OkApiKey
	subLogin.Parameters["sign"] = okcoinSign(params, configure.OkSecretKey)
	// subLogin.Event = "ok_sub_futureusd_trades"
	subLogin.Event = "login"
	ow.Lock()
	ow.subscrible = append(ow.subscrible, &subLogin)
	// ow.subscrible = append(ow.subscrible, &Subscrible{Channel: "addChannel", Event: "ok_sub_futureusd_trades"})
	ow.Unlock()
}

// okcoinSign 参数签名
func okcoinSign(params url.Values, secretKey string) string {
	ctx := md5.New()
	signStr := params.Encode() + "&secret_key=" + secretKey
	fmt.Println("signStr: ", signStr)
	ctx.Write([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(ctx.Sum(nil)))
}
