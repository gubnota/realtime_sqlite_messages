// pkg/handler/ws.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
			h.clients[client.UserID] = client
		case userID := <-h.unregister:
			delete(h.clients, userID)
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

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	client := &Client{
		Conn:   conn,
		UserID: userID,
	}

	h.register <- client
	defer func() {
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
