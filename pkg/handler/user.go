package handler

import (
	"halves/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) UpdateScore(c *gin.Context) {
	var req struct {
		Score int `json:"score" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(string)

	result := h.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("score", req.Score)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"new_score": req.Score})
}

func (h *UserHandler) DeleteUsers(c *gin.Context) {
	// result := h.db.Exec("DELETE FROM users WHERE created_at < ?", time.Now().Add(-100*time.Minute))
	// if result.Error != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete users"})
	// 	return
	// }
	result := h.db.Model(&model.User{}).
		Where("created_at < ?", time.Now().Add(-1*time.Minute)).
		Delete(&model.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete users"})
		return
	}
	h.db.Exec("DELETE FROM devices WHERE 1")
	h.db.Exec("DELETE FROM messages WHERE 1")

	c.JSON(http.StatusOK, gin.H{"message": "users deleted"})

}
