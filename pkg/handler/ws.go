package handler

import (
	"halves/pkg/model"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate proper origins
	},
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
	deviceID := c.GetHeader("X-Device-ID")
	db := c.MustGet("db").(*gorm.DB)
	now := time.Now().Unix()

	if userID != tokenUserID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	client := &Client{
		Conn:   conn,
		UserID: userID,
	}
	// Update device status
	// db.Model(&model.Device{}).Where("id = ?", deviceID).Updates(map[string]interface{}{
	// 	"last_seen": time.Now().Unix(),
	// 	"status":    "online",
	// })
	// Create or update device
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"status": "O"}),
	}).Create(&model.Device{
		ID:        deviceID,
		UserID:    userID,
		LastSeen:  now,
		Status:    "O",
		UserAgent: c.GetHeader("User-Agent"),
	})

	//end of Update device status
	h.register <- client
	defer func() {
		// Update device status
		db.Model(&model.Device{}).
			Where("id = ?", deviceID).
			Update("status", "F")
		h.unregister <- userID
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
