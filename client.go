package wsbase

import "github.com/gorilla/websocket"

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
	Id   string
}
