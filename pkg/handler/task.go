package handler

import (
    "halves/pkg/model"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type TaskHandler struct {
    db  *gorm.DB
    hub *Hub
}

func NewTaskHandler(db *gorm.DB, hub *Hub) *TaskHandler {
    return &TaskHandler{db: db, hub: hub}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
    var req struct {
        SpaceID    string `json:"space_id" binding:"required"`
        RecipeID   string `json:"recipe_id" binding:"required"`
        Department string `json:"department"`
        AssignedTo string `json:"assigned_to"`
        Comment    string `json:"comment"`
        Quantity   int    `json:"quantity" binding:"required,min=1"`
        ETA        int    `json:"eta"` // minutes
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    task := model.Task{
        ID:         uuid.NewString(),
        SpaceID:    req.SpaceID,
        RecipeID:   req.RecipeID,
        Department: req.Department,
        AssignedTo: req.AssignedTo,
        Comment:    req.Comment,
        Quantity:   req.Quantity,
        ETA:        req.ETA,
        Status:     "new",
        ScheduledAt: time.Now(),
        CreatedAt: time.Now(),
    }

    if err := h.db.Create(&task).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
        return
    }

    // Notify via WebSocket
    h.hub.BroadcastToSpace(task.SpaceID, gin.H{
        "type": "task_created",
        "task": task,
    })

    c.JSON(http.StatusCreated, task)
}