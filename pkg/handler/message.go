package handler

import (
	"halves/pkg/model"
	"net/http"
	"strconv"
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

// pkg/handler/message.go:
// func (h *MessageHandler) GetMessages(c *gin.Context) {
// 	userID := c.MustGet("userID").(string)
// 	from := c.Query("from")

// 	var messages []model.Message
// 	query := h.db.Where("receiver = ? AND created_at > ?", userID, from)
// 	if err := query.Order("created_at desc").Find(&messages).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"messages": messages})
//	}
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	fromStr := c.Query("from")

	// Parse Unix timestamp
	var fromTime time.Time
	if fromStr != "" {
		timestamp, err := strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp format"})
			return
		}
		fromTime = time.Unix(timestamp, 0).UTC()
	} else {
		// If no timestamp provided, return all messages
		fromTime = time.Time{}
	}

	var messages []model.Message
	query := h.db.Where("receiver = ? AND created_at > ?", userID, fromTime)
	if err := query.Order("created_at desc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}

	// Convert to response with Unix timestamps
	response := make([]gin.H, len(messages))
	for i, msg := range messages {
		response[i] = gin.H{
			"id":        msg.ID,
			"sender":    msg.Sender,
			"content":   msg.Content,
			"createdAt": msg.CreatedAt.Unix(),
			"delivered": msg.Delivered,
		}
	}

	c.JSON(http.StatusOK, gin.H{"messages": response})
}
