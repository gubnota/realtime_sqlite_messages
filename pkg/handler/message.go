package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	db    *sql.DB
	wsHub *Hub
}

func NewMessageHandler(db *sql.DB, wsHub *Hub) *MessageHandler {
	return &MessageHandler{db: db, wsHub: wsHub}
}

type MessageRequest struct {
	Receiver string `json:"receiver" binding:"required,uuid4"`
	Content  string `json:"content" binding:"required,max=500"`
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	senderID := c.MustGet("userID").(string)
	now := time.Now().Unix()

	res, err := h.db.Exec(`
		INSERT INTO messages (sender, receiver, content, created_at, delivered)
		VALUES (?, ?, ?, ?, false)
	`, senderID, req.Receiver, req.Content, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message"})
		return
	}
	messageID, _ := res.LastInsertId()

	go func() {
		pushURL := os.Getenv("PUSH_WEBHOOK")
		payload := map[string]interface{}{
			"receiver": req.Receiver,
			"sender":   senderID,
			"message":  req.Content,
		}
		jsonPayload, _ := json.Marshal(payload)
		http.Post(pushURL, "application/json", bytes.NewBuffer(jsonPayload))
	}()

	// Notify over WS
	if client, ok := h.wsHub.clients[req.Receiver]; ok {
		msg := gin.H{
			"type": "message",
			"data": gin.H{
				"id":        messageID,
				"sender":    senderID,
				"content":   req.Content,
				"createdAt": now,
			},
		}
		if err := client.Conn.WriteJSON(msg); err == nil {
			h.db.Exec("UPDATE messages SET delivered = true WHERE id = ?", messageID)
			h.db.Exec("UPDATE messages SET delivered = true WHERE sender = ? AND receiver = ? AND created_at <= ?",
				senderID, req.Receiver, now)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        messageID,
		"createdAt": now,
		"delivered": false,
	})
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	fromStr := c.Query("from")

	from := int64(0)
	if fromStr != "" {
		t, err := strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
			return
		}
		from = t
	}

	rows, err := h.db.Query(`
		SELECT id, sender, content, created_at, delivered
		FROM messages WHERE receiver = ? AND created_at > ?
		ORDER BY created_at DESC
	`, userID, from)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}
	defer rows.Close()

	var messages []gin.H
	for rows.Next() {
		var id int64
		var sender, content string
		var createdAt int64
		var delivered bool
		if err := rows.Scan(&id, &sender, &content, &createdAt, &delivered); err == nil {
			messages = append(messages, gin.H{
				"id":        id,
				"sender":    sender,
				"content":   content,
				"createdAt": createdAt,
				"delivered": delivered,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
