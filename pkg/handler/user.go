package handler

import (
	"halves/pkg/model"
	"net/http"

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
