package wsbase

import (
	"encoding/json"
	"fmt"
	"log"

	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient interface {
	SetMessageHandler(MessageHandler)
	SetLogHandler(LogHandler)
	SetReconnectPeriod(time.Duration)
	Start() error
	// Close() error
}

func logHandlerImpl(logType int, val string) {
	log.Println(val)
}

type MessageHandler func(Message)
type LogHandler func(logType int, val string)

func NewWSClient(addr string, path string, secure bool) WSClient {
	inst := &wsclientimpl{
		addr:             addr,
		path:             path,
		secure:           secure,
		reconnectPeriod:  10 * time.Second,
		interrupt:        make(chan os.Signal, 1),
		disconnectSignal: make(chan bool),
		loghandler:       logHandlerImpl,
	}

	return inst
}

type wsclientimpl struct {
	addr            string
	path            string
	secure          bool
	conn            *websocket.Conn
	handler         MessageHandler
	loghandler      LogHandler
	reconnectPeriod time.Duration
	interrupted     bool
	interrupt       chan os.Signal
	// mt               sync.RWMutex
	disconnectSignal chan bool
}

func (c *wsclientimpl) SetMessageHandler(handler MessageHandler) {
	c.handler = handler
}

func (c *wsclientimpl) SetLogHandler(handler LogHandler) {
	c.loghandler = handler
}

func (c *wsclientimpl) SetReconnectPeriod(period time.Duration) {
	c.reconnectPeriod = period
}

func (c *wsclientimpl) Start() (er error) {
	signal.Notify(c.interrupt, os.Interrupt)

	er = c.connect()
	if er != nil {
		return er
	}

	c.reconnectObserver()
	return
}

func (c wsclientimpl) printlog(logtype int, val ...interface{}) {
	c.loghandler(logtype, fmt.Sprint(val...))
}

func (c *wsclientimpl) connect() (er error) {

	scheme := "ws"
	if c.secure {
		scheme = "wss"
	}

	wsserverUrl := url.URL{Scheme: scheme, Host: c.addr, Path: c.path}
	c.printlog(LOG, "connecting to ", wsserverUrl.String())
	c.conn, _, er = websocket.DefaultDialer.Dial(wsserverUrl.String(), nil)

	if er != nil {
		return
	}

	go c.messageHandler()

	return nil
}

func (c *wsclientimpl) messageHandler() {
	done := make(chan error)
	c.interrupted = false
	defer func() {
		c.printlog(INFO, "closing message handler")
		close(done)
	}()

	go func() {
		defer func() {
			c.printlog(INFO, "closing read message..")
		}()
		for {
			msgType, message, erx := c.conn.ReadMessage()
			if erx != nil {
				c.printlog(WARN, "disconnected from server :", erx)
				done <- erx
				return
			}
			if msgType == websocket.TextMessage && c.handler != nil {
				msg := Message{}
				json.Unmarshal(message, &msg)
				c.handler(msg)
			}
		}
	}()

	for {
		select {
		case <-done:
			if !c.interrupted {
				c.disconnectSignal <- true
			}
			return
		case <-c.interrupt:
			{
				c.interrupted = true
				c.close()
				return
			}
		}
	}
}

func (c *wsclientimpl) reconnectObserver() (er error) {
	defer func() {
		close(c.disconnectSignal)
	}()

	for {
		select {
		case <-c.disconnectSignal:
			{
				do := true
				for do {
					c.printlog(INFO, "reconnect after "+c.reconnectPeriod.String())
					<-time.After(c.reconnectPeriod)
					c.printlog(INFO, "reconnecting to ws server...")
					er = c.connect()
					if er != nil {
						c.printlog(ERR, "Unable to connect : ", er.Error())
						// c.disconnectSignal <- true
					} else {
						do = false
					}
				}
			}
		case <-c.interrupt:
			return
		}
	}
}

func (c *wsclientimpl) close() (er error) {
	er = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	<-time.After(time.Second)
	return er
}
