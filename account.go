package main

import (
	"github.com/gorilla/websocket"
	"github.com/sencydai/qyh15c/encrypt"
	"github.com/sencydai/utils/timer"
)

type AccountData struct {
	conn       *websocket.Conn
	connStatus connectStatus
	pid        uint32
	isClose    bool

	accountName string
	accountId   int32
	actorId     float64

	encrypt *encrypt.Encrypt
	timers  []*timer.Timer
	data    map[string]interface{}
}
