package wsbase

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/randyardiansyah25/wsbase-handler/common/slices"
	"github.com/randyardiansyah25/wsbase-handler/common/winnet"

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
	SetLogHandler(LogHandler)
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
		onReadMsg:  nil,
		loghandler: PrintDefault,
	}
}

type hubimpl struct {
	Clients           map[*Client]bool
	PushClientMessage chan Message
	Register          chan *Client
	Unregister        chan *Client
	upgrader          websocket.Upgrader
	onReadMsg         OnReadMessageFunc
	loghandler        LogHandler
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
	go writePump(client)
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
		h.printlog(LOG, "Get pong from [", c.Id, "], renew pong wait to ", nextTime.Format("15:04:05"))
		c.Conn.SetReadDeadline(nextTime)
		return nil
	})

	for {
		_, message, er := c.Conn.ReadMessage()
		if er != nil {
			isCloseError := false
			if websocket.IsUnexpectedCloseError(er, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.printlog(ERR, "Getting unexpected closing client :", er.Error())
				isCloseError = true
			} else if websocket.IsCloseError(er, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.printlog(ERR, "Getting closing client [", c.Id, "] :", er.Error())
				isCloseError = true
			} else if opErr, ok := er.(*net.OpError); ok {
				if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
					if errno, ok := sysErr.Err.(syscall.Errno); ok {
						if errno == syscall.ECONNABORTED || errno == winnet.WSAECONNABORTED {
							h.printlog(ERR, "closing client [", c.Id, "] : aborted:", er.Error())
							isCloseError = true
						} else if errno == syscall.ECONNRESET || errno == winnet.WSAECONNRESET {
							h.printlog(ERR, "closing client [", c.Id, "] : reset :", er.Error())
							isCloseError = true
						}
					}
				}
			}

			if !isCloseError {
				h.printlog(ERR, "Getting unknown closing client [", c.Id, "] : ", er.Error())
			}

			break
		}

		smessage := Message{}
		h.printlog(LOG, "RECV,", c.Id, ",", string(message))
		_ = json.Unmarshal(message, &smessage)
		if h.onReadMsg != nil {
			h.onReadMsg(string(message))
		}
	}
}

func (h hubimpl) printlog(logtype int, val ...interface{}) {
	h.loghandler(logtype, fmt.Sprint(val...))
}

func writePump(c *Client) {
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

func (h *hubimpl) SetLogHandler(handler LogHandler) {
	h.loghandler = handler
}
