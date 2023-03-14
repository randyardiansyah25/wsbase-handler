package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	strutils "github.com/randyardiansyah25/libpkg/util/str"
	"github.com/randyardiansyah25/wsbase-handler"
)

var RequestList = make(map[string](chan Message))

type Message struct {
	Type     int         `json:"type"`
	SenderId string      `json:"sender_id"`
	To       []string    `json:"to"`
	Action   string      `json:"action"`
	Title    string      `json:"title"`
	Body     interface{} `json:"body"`
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	strutils.PrepareNumberGenerator()

	hub := wsbase.NewHub()
	go hub.Run()

	r := gin.Default()

	//accept new client websocket connection
	r.GET("/connect/:id", func(c *gin.Context) {
		hub.RegisterClient(c.Param("id"), c.Writer, c.Request)
	})

	r.POST("/ack-req", func(ctx *gin.Context) {
		msg := Message{}
		ctx.ShouldBindJSON(&msg)
		fmt.Println(msg)
		requestId := msg.Action

		ch, ok := RequestList[requestId]
		if ok {
			ch <- msg
		}

		ctx.String(200, "success")
	})

	r.POST("/request/to/:id", func(ctx *gin.Context) {
		reqId, _ := strutils.GenerateNumber(6)

		payload := struct {
			CustomerId      string  `form:"customer_id"`
			CustomerName    string  `form:"customer_name"`
			CustomerBalance float64 `form:"customer_balance"`
		}{}

		ctx.ShouldBind(&payload)

		msg := wsbase.Message{}
		msg.Type = wsbase.TypePrivate
		msg.Action = reqId
		msg.Title = "Ack Mechanism"
		msg.Body = payload
		msg.To = []string{ctx.Param("id")}

		deadline := time.Now().Add(30 * time.Second)
		c, cancelC := context.WithDeadline(context.Background(), deadline)
		defer cancelC()

		resCh := make(chan Message)
		defer close(resCh)

		RequestList[reqId] = resCh
		defer delete(RequestList, reqId)

		hub.PushMessage(msg)

		select {
		case respMsg := <-resCh:
			fmt.Println(respMsg)
			b, _ := json.Marshal(respMsg)
			ctx.String(200, string(b))
		case <-c.Done():
			fmt.Println("Timeout exceeded..")
			ctx.String(500, "Timeout exceeded..")
			return
		}
	})

	r.Run(":8881")
}
