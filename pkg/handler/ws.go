package handler

import (
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	Conn   *websocket.Conn
	UserID string
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan string
	mutex      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan string),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.UserID] = client
			h.mutex.Unlock()
		case userID := <-h.unregister:
			h.mutex.Lock()
			if client, ok := h.clients[userID]; ok {
				client.Conn.Close()
				delete(h.clients, userID)
			}
			h.mutex.Unlock()
		}
	}
}

func (h *Hub) WebSocketHandler(c *gin.Context) {
	userID := c.Param("uuid")
	tokenUserID := c.MustGet("userID").(string)
	if userID != tokenUserID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		return
	}
	deviceID := c.GetHeader("X-Device-ID")
	db := c.MustGet("db").(*sql.DB)
	now := time.Now().Unix()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "websocket upgrade failed"})
		return
	}
	client := &Client{Conn: conn, UserID: userID}

	// Create/update device
	_, err = db.Exec(`
	INSERT INTO devices (id, user_id, last_seen, status, user_agent)
	VALUES (?, ?, ?, 'O', ?)
	ON CONFLICT(id) DO UPDATE SET last_seen = excluded.last_seen, status = 'O'
	`, deviceID, userID, now, c.GetHeader("User-Agent"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "device status update failed"})
		return
	}

	h.register <- client
	defer func() {
		db.Exec("UPDATE devices SET status = 'F' WHERE id = ?", deviceID)
		h.unregister <- userID
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
