package services

import (
	"sync"
	"time"

	"adam-french.co.uk/backend/models"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	clients  = make(map[*websocket.Conn]bool)
	messages = make([]models.Message, 0)
	mu       sync.Mutex
)

func HandleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	mu.Lock()
	clients[conn] = true

	// Send existing message history to new client
	for _, msg := range messages {
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

		incoming.CreatedAt = time.Now()

		// Store and broadcast
		mu.Lock()
		messages = append(messages, incoming)

		for client := range clients {
			if err := client.WriteJSON(incoming); err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}

	// Cleanup on disconnect
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
}
