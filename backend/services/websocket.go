package services

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"

	"github.com/gorilla/websocket"
)

const maxMessages = 50

var allowedDomain string

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		host := u.Hostname()
		return host == allowedDomain || host == "www."+allowedDomain || host == "localhost"
	},
}

var (
	clients      = make(map[*websocket.Conn]bool)
	mu           sync.Mutex
	wsDB         *gorm.DB
	nextAuthorID uint
)

const (
	rateLimitWindow  = time.Second
	rateLimitMaxMsgs = 10
)

func InitWebSocket(database *gorm.DB, domain string) {
	wsDB = database
	allowedDomain = domain
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

	msgCount := 0
	windowStart := time.Now()

	for {
		var incoming models.Message
		if err := conn.ReadJSON(&incoming); err != nil {
			break
		}

		now := time.Now()
		if now.Sub(windowStart) > rateLimitWindow {
			msgCount = 0
			windowStart = now
		}
		msgCount++
		if msgCount > rateLimitMaxMsgs {
			continue
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
