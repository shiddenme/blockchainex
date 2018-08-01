package entity

import (
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
)

type Response struct {
	Ret  int         `json:"ret"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (r *Response) Fail(ret int, msg string) {
	r.Ret = ret
	r.Msg = msg
}

func (r *Response) Success(data interface{}) {
	r.Ret = 0
	if r.Msg == "" {
		r.Msg = "success"
	}
	r.Data = data
}

func (r *Response) Write(ctx echo.Context) error {
	data, _ := json.Marshal(r)
	return ctx.JSONBlob(http.StatusOK, data)
}
