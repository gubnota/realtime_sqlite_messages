package handler

import (
    "halves/pkg/model"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type SpaceHandler struct {
    db *gorm.DB
}

func NewSpaceHandler(db *gorm.DB) *SpaceHandler {
    return &SpaceHandler{db: db}
}

func (h *SpaceHandler) CreateSpace(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    var req struct {
        Name string `json:"name" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    space := model.Space{
        ID:        uuid.NewString(),
        Name:      req.Name,
        OwnerID:   userID,
        CreatedAt: time.Now().Unix(),
    }

    if err := h.db.Create(&space).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create space"})
        return
    }

    c.JSON(http.StatusCreated, space)
}