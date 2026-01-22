package handlers

import (
	"time"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (store *Store) ConnectWebSocket(ctx *gin.Context) {
	conn, err := services.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()
	for {
		conn.WriteMessage(websocket.TextMessage, []byte("Hello Websocket!"))
		time.Sleep(time.Second)
	}
}
