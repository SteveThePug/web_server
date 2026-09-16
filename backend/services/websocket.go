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

// This file is the chat hub behind GET /ws. It is deliberately a package-level
// singleton (one process, one chat room) rather than an injectable type, so the
// connection set, the database handle and the allowed origin all live in
// package vars initialised once by InitWebSocket.

// maxMessages is both the size of the history replayed to a new client and the
// number of rows kept in the database: every insert trims older messages.
const maxMessages = 50

// allowedDomain is the site's domain, set by InitWebSocket and read by
// Upgrader.CheckOrigin. It is a package var because the Upgrader is one too.
var allowedDomain string

// Upgrader turns the HTTP request into a WebSocket.
//
// CheckOrigin is the only cross-origin defence on this endpoint: the browser
// sends cookies with a WebSocket handshake regardless of origin and there is
// no preflight, so without this check any site could open a socket as a
// logged-in visitor. Default gorilla behaviour (same-host only) is too strict
// here because the site is reached as both example.com and www.example.com,
// and "localhost" is allowed for the Vite dev server.
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
	clients = make(map[*websocket.Conn]bool)
	// mu guards clients AND serialises the write-then-broadcast sequence, so
	// that two concurrent senders cannot interleave and deliver messages to
	// different clients in different orders.
	mu   sync.Mutex
	wsDB *gorm.DB
	// nextAuthorID is a per-process counter handed out on connect. It is NOT
	// a user id and has no meaning across restarts: the chat is pseudonymous,
	// and the front end only uses it to colour/group consecutive messages
	// from the same connection.
	nextAuthorID uint
)

// Per-connection flood limit, separate from services.RateLimiter because it
// counts frames on one socket rather than requests from one IP.
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

// wsHistoryEvent is sent once on connect with the recent messages the client
// may see. Sending it as a single batch lets the client replace its list
// atomically, so a reconnect does not blank the chat and reload every image.
type wsHistoryEvent struct {
	Action   string           `json:"action"`
	Messages []models.Message `json:"messages"`
}

// InitWebSocket wires the package-level hub to the database and the origin
// allow-list. Call it once, before serving.
func InitWebSocket(database *gorm.DB, domain string) {
	wsDB = database
	allowedDomain = domain
}

// HandleWebSocket serves one chat connection. isAdmin controls whether the
// connection may see private messages, send them, and delete messages.
func HandleWebSocket(conn *websocket.Conn, isAdmin bool) {
	defer conn.Close()

	// Registration, history read and history send all happen under one lock.
	// That is the point: it guarantees the client is already in `clients`
	// before its snapshot is taken, so a message broadcast concurrently is
	// either in the history or delivered afterwards — never dropped in the
	// gap, and never delivered before the history that would overwrite it.
	mu.Lock()
	clients[conn] = isAdmin
	nextAuthorID++
	authorID := nextAuthorID

	history := make([]models.Message, 0, maxMessages)
	historyQuery := wsDB.Order("created_at ASC").Limit(maxMessages)
	if !isAdmin {
		historyQuery = historyQuery.Where("private = ?", false)
	}
	historyQuery.Find(&history)

	if err := conn.WriteJSON(wsHistoryEvent{Action: "history", Messages: history}); err != nil {
		delete(clients, conn)
		mu.Unlock()
		return
	}
	mu.Unlock()

	// Fixed-window flood limit. Over-limit frames are dropped silently with
	// no error to the client and without resetting the window.
	msgCount := 0
	windowStart := time.Now()

	for {
		var incoming wsIncoming
		// Any read error — clean close, network drop or malformed JSON —
		// ends the connection; there is no attempt to resynchronise.
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
		// Keep only the newest maxMessages rows. This is a soft delete
		// (models.Message has gorm.DeletedAt), so the rows stay in the table
		// with deleted_at set and GORM filters them out of later reads; the
		// table therefore grows forever even though the chat does not.
		// The subquery is itself soft-delete filtered, so already-trimmed
		// rows are not re-deleted.
		wsDB.Where("id NOT IN (?)",
			wsDB.Model(&models.Message{}).Select("id").Order("created_at DESC").Limit(maxMessages),
		).Delete(&models.Message{})

		// Deleting from a map while ranging over it is explicitly allowed in
		// Go, so pruning dead clients inline here is safe.
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
