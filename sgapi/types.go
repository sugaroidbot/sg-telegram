package sgapi

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Instance struct {
	Endpoint string
}

type WsConn struct {
	Id   uuid.UUID
	conn *websocket.Conn
}
