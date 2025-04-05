package handler

import (
	"halves/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db    *gorm.DB
	wsHub *Hub
}

func NewMessageHandler(db *gorm.DB, wsHub *Hub) *MessageHandler {
	return &MessageHandler{
		db:    db,
		wsHub: wsHub,
	}
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

	message := model.Message{
		Sender:    senderID,
		Receiver:  req.Receiver,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if result := h.db.Create(&message); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	// Notify receiver via WebSocket
	if client, ok := h.wsHub.clients[req.Receiver]; ok {
		err := client.Conn.WriteJSON(gin.H{
			"type": "message",
			"data": gin.H{
				"id":        message.ID,
				"sender":    message.Sender,
				"content":   message.Content,
				"createdAt": message.CreatedAt,
			},
		})

		if err == nil {
			h.db.Model(&message).Update("delivered", true)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        message.ID,
		"createdAt": message.CreatedAt,
		"delivered": message.Delivered,
	})
}
