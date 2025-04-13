package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
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

	_, err := h.db.Exec("UPDATE users SET score = ? WHERE id = ?", req.Score, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"new_score": req.Score})
}
