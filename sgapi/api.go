package sgapi

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/withmandala/go-log"

	"os"
)

var logger = log.New(os.Stdout)

func Listen(w *WsConn, cb func(resp string)) error {
	defer w.conn.Close()
	for {
		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			logger.Warn(err)
			return err
		} else {
			cb(string(msg))
		}
	}
}

func Send(w *WsConn, message string) error {
	err := w.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		w.conn.Close()
		return err
	}
	return nil
}

func New(i Instance, u uuid.UUID) (*WsConn, error) {
	endpoint := fmt.Sprintf("%s/%s", i.Endpoint, u)
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}
	return &WsConn{
		conn: conn,
		Id:   u,
	}, nil
}
