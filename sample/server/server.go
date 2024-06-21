package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/randyardiansyah25/wsbase-handler"
)

func main() {
	hub := wsbase.NewHub()
	hub.SetOnCloseHandlerFunc(func(id string) {
		fmt.Println("Ini custom saat menerima info close dari client. client yang terputus : ", id)
	})

	hub.SetOnPongHandlerFunc(func(id string, nextPongWait time.Time) {
		fmt.Println("Kalo ini, custom untuk penanganan pong dari client. Client yang mengirim : ", id, ". Pong selanjutnya harus diterima pada : ", nextPongWait.Format("15:04:05"))
	})
	go hub.Run()

	r := gin.Default()

	//accept new client websocket connection
	r.GET("/connect/:id", func(c *gin.Context) {
		hub.RegisterClient(c.Param("id"), c.Writer, c.Request)
	})

	r.POST("/broadcast", func(ctx *gin.Context) {
		msg := wsbase.Message{}
		msg.Type = wsbase.TypeBroadcast
		msg.Action = "action_broadcast"
		msg.Title = ctx.Request.FormValue("title")
		msg.Body = ctx.Request.FormValue("message")

		hub.PushMessage(msg)
		ctx.String(200, "okey")
	})

	r.POST("/publish/to", func(ctx *gin.Context) {
		requestPayload := struct {
			Title   string   `form:"title"`
			Message string   `form:"message"`
			To      []string `form:"to"`
		}{}

		ctx.ShouldBind(&requestPayload)

		msg := wsbase.Message{}
		msg.Type = wsbase.TypePrivate
		msg.Action = "action_private_message"
		msg.Title = requestPayload.Title
		msg.Body = requestPayload.Message
		msg.To = requestPayload.To
		fmt.Println(requestPayload.To)

		hub.PushMessage(msg)
		ctx.String(200, "okey")
	})

	r.Run(":8881")
}
