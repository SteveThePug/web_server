package services

import (
	"net/http"
	"sync"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"

	"github.com/gorilla/websocket"
)

const maxMessages = 50

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients      = make(map[*websocket.Conn]bool)
	mu           sync.Mutex
	wsDB         *gorm.DB
	nextAuthorID uint
)

func InitWebSocket(database *gorm.DB) {
	wsDB = database
}

func HandleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	nextAuthorID++
	authorID := nextAuthorID

	var history []models.Message
	wsDB.Order("created_at ASC").Limit(maxMessages).Find(&history)

	for _, msg := range history {
		if err := conn.WriteJSON(msg); err != nil {
			mu.Unlock()
			return
		}
	}
	mu.Unlock()

	for {
		var incoming models.Message
		if err := conn.ReadJSON(&incoming); err != nil {
			break
		}

		incoming.AuthorID = authorID

		mu.Lock()
		wsDB.Create(&incoming)
		wsDB.Where("id NOT IN (?)",
			wsDB.Model(&models.Message{}).Select("id").Order("created_at DESC").Limit(maxMessages),
		).Delete(&models.Message{})

		for client := range clients {
			if err := client.WriteJSON(incoming); err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}

	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
}
