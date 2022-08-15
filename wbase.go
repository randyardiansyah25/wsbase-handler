package wsbase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/randyardiansyah25/wsbase-handler/common/slices"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type OnReadMessageFunc func(msg string)

type Hub interface {
	Run()
	RegisterClient(id string, w http.ResponseWriter, r *http.Request) error
	SetOnReadMessageFunc(OnReadMessageFunc)
	PushMessage(Message)
}

func NewHub() Hub {
	return &hubimpl{
		Clients:           make(map[*Client]bool),
		PushClientMessage: make(chan Message),
		Register:          make(chan *Client),
		Unregister:        make(chan *Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		onReadMsg: nil,
	}
}

type hubimpl struct {
	Clients           map[*Client]bool
	PushClientMessage chan Message
	Register          chan *Client
	Unregister        chan *Client
	upgrader          websocket.Upgrader
	onReadMsg         OnReadMessageFunc
}

func (h *hubimpl) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			delete(h.Clients, client)
			close(client.Send)
		case message := <-h.PushClientMessage:
			{
				for client := range h.Clients {
					if message.SenderId == client.Id {
						continue
					}
					b, er := json.Marshal(message)
					if er != nil {
						fmt.Println("[wsbase] Error while marshal message : ", er.Error())
						continue
					}
					if message.Type == TypeBroadcast {
						if message.SenderId == client.Id {
							continue
						}

						client.Send <- b
					} else {
						if slices.Contains(message.To, client.Id) {
							client.Send <- b
						}
					}
					//entities.PrintLog("Broadcast to ", client.Id)
				}
			}
		}
	}
}

func (h *hubimpl) RegisterClient(id string, w http.ResponseWriter, r *http.Request) (er error) {
	conn, er := h.upgrader.Upgrade(w, r, nil)
	if er != nil {
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte),
		Id:   id,
	}

	h.Register <- client

	go readPump(h, client)
	go writePump(h, client)
	return
}

func (h *hubimpl) SetOnReadMessageFunc(handler OnReadMessageFunc) {
	h.onReadMsg = handler
}
func (h *hubimpl) PushMessage(msg Message) {
	h.PushClientMessage <- msg
}

func readPump(h *hubimpl, c *Client) {
	defer func() {
		h.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(appData string) error {
		nextTime := time.Now().Add(pongWait)
		fmt.Println("Get pong from [", c.Id, "], renew pong wait to ", nextTime.Format("15:04:05"))
		c.Conn.SetReadDeadline(nextTime)
		return nil
	})

	for {
		_, message, er := c.Conn.ReadMessage()
		if er != nil {
			if websocket.IsUnexpectedCloseError(er, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("Getting unexpected closing client :", er.Error())
				break
			} else if websocket.IsCloseError(er, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("Getting closing client :", er.Error())
				break
			}
		}
		smessage := Message{}
		fmt.Println("string msg", string(message))
		_ = json.Unmarshal(message, &smessage)
		fmt.Println("getting message :", &message)
		if h.onReadMsg != nil {
			h.onReadMsg(string(message))
		}
	}
}

func writePump(h *hubimpl, c *Client) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			{
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					//The hub closed the channel
					c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				if er := c.Conn.WriteMessage(websocket.TextMessage, msg); er != nil {
					return
				}
			}
		case <-ticker.C:
			{
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				if er := c.Conn.WriteMessage(websocket.PingMessage, nil); er != nil {
					//if we got any error, close the unused this socket.
					return
				}
			}
		}
	}

}
