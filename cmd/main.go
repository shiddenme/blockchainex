package main

import (
	binancews "blockchainex/binance/ws"
	_ "blockchainex/cmd/service/binance"
	"blockchainex/cmd/service/example"
	"blockchainex/cmd/service/fcoin"
	_ "blockchainex/cmd/service/okex"
	"blockchainex/configure"
	"blockchainex/fcoin/ws"
	okexws "blockchainex/okex/ws"
	"context"
	_ "context"
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"time"
)

func init() {
	flag.IntVar(&configure.Port, "port", 9001, "service listen port")
	flag.StringVar(&configure.OkApiKey, "ok-api-key", "", "okex api key")
	flag.StringVar(&configure.OkSecretKey, "ok-secret-key", "", "okex api key")
	flag.BoolVar(&configure.IsWall, "is-wall", false, "is wall")
	flag.StringVar(&configure.WallProxyAddr, "wall-proxy-addr", "http://127.0.0.1:1080", "wall proxy address")
	flag.Parse()

	// binance.DepthWebsocket(context.Background())
	// go binance.DepthWebsocket1(context.Background())
	// go example.ClientExample("ticker.btcusdt")
	// go example.ClientExample(`{"cmd":"sub","args":["ticker.btcusdt","depth.L20.btcusdt","candle.H1.btcusdt"],"id":"1"}`)
	// go example.FCoinClient(`{"cmd":"sub","args":["ticker.btcusdt","depth.L20.btcusdt","candle.H1.btcusdt"],"id":"1"}`)
	// go example.ClientExample(`{"cmd":"sub","args":["depth.L100.btcusdt"],"id":"1"}`)
	// go example.ClientExample("client2")
	// go ws.Start()

	// go example.BinanceClient(context.Background())

	go binancewebsocket()

	// go okws()
}

func binancewebsocket() {
	// bw := binance.NewBinanceWebsocket(context.Background())
	// bw.Depth("btcusdt")
	ctx, cancelCtx := context.WithCancel(context.Background())
	bws := binancews.NewBinanceWebsocket(ctx)
	// bws.AggTrade("btcusdt")
	bws.Trade("btcusdt")
	time.Sleep(30 * time.Second)
	cancelCtx()
}

func okws() {
	// okex.Ticker()
	ow := okexws.NewOkexWebsocket()
	// fmt.Println("okws")
	ow.SubFutureusdTicker()
	ow.SubFutureusdKLine()
	ow.SubFutureusdDepth()
	ow.SubFutureusdTrade()
	ow.SubFutureusdIndex()
	ow.ForecastPrice()
	ow.Login()
	ow.Ticker()
	// time.Sleep(30 * time.Second)
	// ow.AddMessage("{'event':'addChannel','channel':'ok_sub_futureusd_btc_ticker_this_week'}")
	// ow.AddMessage("{'event':'addChannel','channel':'ok_sub_futureusd_btc_kline_this_week_1min'}")
}

func fcoinws() {
	go ws.Start()

	go example.BinanceClient(context.Background())
}

func main() {
	e := echo.New()

	// e.Use(middleware.Logger())

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] method=${method} path=${path} remote=${remote_ip} status=${status}\n",
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, `
			blockchain exchange
		`)
	})

	okRouter := e.Group("/ok")
	{
		okRouter.POST("/", func(c echo.Context) error {
			return c.String(http.StatusOK, `
				OKEX
			`)
		})
	}

	exampleRouter := e.Group("/example")
	{
		exampleRouter.GET("/", example.Home)
		exampleRouter.GET("/echo", example.Echo)
		exampleRouter.GET("/ws", example.WsHandler)
	}

	fcoinRouter := e.Group("/fcoin")
	{
		fcoinRouter.GET("/ticker", fcoin.GetTicker)
		fcoinRouter.GET("/candle", fcoin.GetCandle)
		fcoinRouter.GET("/depth", fcoin.GetDepth)
	}

	// e.Logger.Fatal(e.StartTLS(":"+fmt.Sprint(configure.Port), "cert.pem", "key.pem"))
	e.Logger.Fatal(e.Start(":" + fmt.Sprint(configure.Port)))
}
