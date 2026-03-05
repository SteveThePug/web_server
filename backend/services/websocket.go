package services

import (
	"net/http"
	"sync"
	"time"

	"adam-french.co.uk/backend/models"

	"github.com/gorilla/websocket"
)

const maxMessages = 50

type MessageRing struct {
	buf  []models.Message
	next int  // index to write next
	full bool // have we wrapped at least once
}

func NewMessageRing(size int) *MessageRing {
	return &MessageRing{
		buf: make([]models.Message, size),
	}
}

func (r *MessageRing) Add(m models.Message) {
	r.buf[r.next] = m
	r.next = (r.next + 1) % len(r.buf)
	if r.next == 0 {
		r.full = true
	}
}

// Messages returns messages in chronological order (oldest -> newest).
func (r *MessageRing) Messages() []models.Message {
	if !r.full {
		// buffer not wrapped yet; only use [0:next)
		return r.buf[:r.next]
	}

	// wrapped: start at next, then to end, then from 0 to next
	out := make([]models.Message, len(r.buf))
	copy(out, r.buf[r.next:])
	copy(out[len(r.buf)-r.next:], r.buf[:r.next])
	return out
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients  = make(map[*websocket.Conn]bool)
	messages = NewMessageRing(maxMessages)
	mu       sync.Mutex
)

func HandleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	mu.Lock()
	clients[conn] = true

	history := messages.Messages()

	// Send existing message history to new client
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

		incoming.CreatedAt = time.Now()

		// Store and broadcast
		mu.Lock()
		messages.Add(incoming)

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
