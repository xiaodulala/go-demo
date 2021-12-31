package home

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type MsgType uint8

const (
	Ping MsgType = iota + 1
	Pong
	Test
)

type Msg struct {
	MsgType MsgType `json:"msgType"`
	Data    string  `json:"data"`
}

type Controller struct {
	upgrade websocket.Upgrader
}

func NewHomeController() *Controller {
	return &Controller{
		upgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
