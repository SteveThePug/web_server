package handlers

import (
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

func (store *Store) ConnectWebSocket(ctx *gin.Context) {
	conn, err := services.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	services.HandleWebSocket(conn)
}
