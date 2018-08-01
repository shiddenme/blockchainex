package fcoin

import (
	"blockchainex/cmd/service/entity"
	"blockchainex/fcoin/ws"
	// "encoding/json"
	"github.com/labstack/echo"
)

func GetTicker(ctx echo.Context) error {
	fcoinws := ws.GetWSFCoin()
	// data, _ := json.Marshal(fcoinws.Ticker)
	var response entity.Response
	// fcoinws.Ticker.SetDate()
	response.Success(fcoinws.Ticker)
	return response.Write(ctx)
}

func GetCandle(ctx echo.Context) error {
	fcoinws := ws.GetWSFCoin()
	var response entity.Response
	response.Success(fcoinws.Candle)
	return response.Write(ctx)
}

func GetDepth(ctx echo.Context) error {
	fcoinws := ws.GetWSFCoin()
	var response entity.Response
	response.Success(fcoinws.Depth)
	return response.Write(ctx)
}
