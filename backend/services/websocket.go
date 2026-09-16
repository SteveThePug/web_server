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
	// clients maps each open connection to whether it belongs to an admin.
	clients      = make(map[*websocket.Conn]bool)
	mu           sync.Mutex
	wsDB         *gorm.DB
	nextAuthorID uint
)

const (
	rateLimitWindow  = time.Second
	rateLimitMaxMsgs = 10
)

// wsIncoming is the envelope clients send over the socket. A plain chat
// message carries text/fileUrl/private; an admin delete carries action+id.
type wsIncoming struct {
	Action  string `json:"action,omitempty"`
	ID      uint   `json:"id,omitempty"`
	Text    string `json:"text"`
	FileURL string `json:"fileUrl,omitempty"`
	Private bool   `json:"private,omitempty"`
}

// wsDeleteEvent is broadcast to every client when a message is removed.
type wsDeleteEvent struct {
	Action string `json:"action"`
	ID     uint   `json:"id"`
}

func InitWebSocket(database *gorm.DB, domain string) {
	wsDB = database
	allowedDomain = domain
}

// HandleWebSocket serves one chat connection. isAdmin controls whether the
// connection may see private messages, send them, and delete messages.
func HandleWebSocket(conn *websocket.Conn, isAdmin bool) {
	defer conn.Close()

	mu.Lock()
	clients[conn] = isAdmin
	nextAuthorID++
	authorID := nextAuthorID

	var history []models.Message
	historyQuery := wsDB.Order("created_at ASC").Limit(maxMessages)
	if !isAdmin {
		historyQuery = historyQuery.Where("private = ?", false)
	}
	historyQuery.Find(&history)

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
		var incoming wsIncoming
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

		if incoming.Action == "delete" {
			if isAdmin && incoming.ID != 0 {
				deleteMessage(incoming.ID)
			}
			continue
		}

		msg := models.Message{
			Content:  incoming.Text,
			FileURL:  incoming.FileURL,
			AuthorID: authorID,
			// Only admins may mark a message private.
			Private: incoming.Private && isAdmin,
		}

		mu.Lock()
		wsDB.Create(&msg)
		wsDB.Where("id NOT IN (?)",
			wsDB.Model(&models.Message{}).Select("id").Order("created_at DESC").Limit(maxMessages),
		).Delete(&models.Message{})

		for client, clientAdmin := range clients {
			if msg.Private && !clientAdmin {
				continue
			}
			if err := client.WriteJSON(msg); err != nil {
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

// deleteMessage soft-deletes a message and tells every client to drop it.
// Clients that never received the message (non-admins for a private one)
// simply ignore the unknown id.
func deleteMessage(id uint) {
	mu.Lock()
	defer mu.Unlock()

	res := wsDB.Delete(&models.Message{}, id)
	if res.Error != nil || res.RowsAffected == 0 {
		return
	}

	event := wsDeleteEvent{Action: "delete", ID: id}
	for client := range clients {
		if err := client.WriteJSON(event); err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
