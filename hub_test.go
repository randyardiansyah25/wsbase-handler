package wsbase_test

import (
	"testing"

	"github.com/randyardiansyah25/wsbase-handler"

	"github.com/gin-gonic/gin"
)

func TestServer(t *testing.T) {
	hub := wsbase.NewHub()
	go hub.Run()

	r := gin.Default()
	r.GET("/connect/:id", func(c *gin.Context) {
		hub.RegisterClient(c.Param("id"), c.Writer, c.Request)
		c.String(200, "done")
	})

	r.POST("/publish", func(ctx *gin.Context) {
		msg := wsbase.Message{}
		ctx.BindJSON(&msg)
		hub.PushMessage(msg)
		ctx.String(200, "okey")
	})

	r.Run(":8881")

}
