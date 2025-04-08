package handler

import (
	"halves/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Add at the top
type GameDTO struct {
	ID       uint      `json:"id"`
	Sender   string    `json:"sender"`
	Receiver string    `json:"receiver"`
	Created  time.Time `json:"created"`
}
type GameResponse struct {
	ID       uint      `json:"id"`
	Sender   string    `json:"sender"`
	Receiver string    `json:"receiver"`
	Created  time.Time `json:"created_at"`
	Status   string    `json:"status"`
}
type LeaderboardHandler struct {
	db *gorm.DB
}

type GameHandler struct {
	db  *gorm.DB
	hub *Hub
}

func NewLeaderboardHandler(db *gorm.DB) *LeaderboardHandler {
	return &LeaderboardHandler{db: db}
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	var entries []model.Leaderboard

	result := h.db.Order("score DESC").Limit(100).Find(&entries)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func NewGameHandler(db *gorm.DB, hub *Hub) *GameHandler {
	return &GameHandler{db: db, hub: hub}
}

func (h *GameHandler) CreateGame(c *gin.Context) {
	var req struct {
		Receiver string `json:"receiver" binding:"required,uuid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game := model.Game{
		Sender:   c.MustGet("userID").(string),
		Receiver: req.Receiver,
		Created:  time.Now(),
	}

	if result := h.db.Create(&game); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game"})
		return
	}

	// Send limited game data
	h.sendGameNotification(game.Receiver, gin.H{
		"type": "game_invite",
		"game": GameDTO{
			ID:       game.ID,
			Sender:   game.Sender,
			Receiver: game.Receiver,
			Created:  game.Created,
		},
	})

	// Start game timeout
	go h.gameTimeoutWorker(game.ID)

	c.JSON(http.StatusCreated, GameDTO{
		ID:       game.ID,
		Sender:   game.Sender,
		Receiver: game.Receiver,
		Created:  game.Created,
	})
}

func (h *GameHandler) gameTimeoutWorker(gameID uint) {
	time.AfterFunc(2*time.Hour, func() {
		h.db.Transaction(func(tx *gorm.DB) error {
			var game model.Game
			if err := tx.First(&game, gameID).Error; err != nil {
				return err
			}

			if game.Status == "open" {
				// Set default votes if not voted
				if game.Svote == 0 && game.Rvote == 0 {
					game.Svote = -1
					game.Rvote = -1
				}

				game.Status = "closed"
				if err := tx.Save(&game).Error; err != nil {
					return err
				}

				h.calculateScores(tx, &game)

				h.sendGameNotification(game.Sender, gin.H{
					"type": "game_timeout",
					"game": GameResponse{
						ID:       game.ID,
						Sender:   game.Sender,
						Receiver: game.Receiver,
						Created:  game.Created,
						Status:   game.Status,
					},
				})

				h.sendGameNotification(game.Receiver, gin.H{
					"type": "game_timeout",
					"game": GameResponse{
						ID:       game.ID,
						Sender:   game.Sender,
						Receiver: game.Receiver,
						Created:  game.Created,
						Status:   game.Status,
					},
				})
			}
			return nil
		})
	})
}

func (h *GameHandler) HandleVote(c *gin.Context) {
	var req struct {
		GameID uint `json:"game_id" binding:"required"`
		Vote   int  `json:"vote" binding:"required,oneof=-1 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(string)

	var game model.Game
	if err := h.db.First(&game, req.GameID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	tx := h.db.Begin()
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		tx.Rollback()
	// 	}
	// }()
	// if err := tx.Commit().Error; err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
	// 	return
	// }

	updateField := ""
	switch {
	case game.Sender == userID && game.Svote == 0:
		updateField = "svote"
	case game.Receiver == userID && game.Rvote == 0:
		updateField = "rvote"
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid vote operation"})
		return
	}

	if err := tx.Model(&game).Update(updateField, req.Vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vote failed"})
		return
	}

	// Check if both voted
	if game.Svote != 0 && game.Rvote != 0 {
		h.calculateScores(tx, &game)
		if err := tx.Model(&game).Update("status", "closed").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close game"})
			return
		}

		// Notify both players
		h.sendGameNotification(game.Sender, gin.H{
			"type": "game_result",
			"game": game,
		})
		h.sendGameNotification(game.Receiver, gin.H{
			"type": "game_result",
			"game": game,
		})
	}

	tx.Commit()
	c.JSON(http.StatusOK, game)
}

// GetActiveGames returns current user's active games
func (h *GameHandler) GetActiveGames(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var games []model.Game
	result := h.db.Where("(sender = ? OR receiver = ?) AND status = 'open'", userID, userID).Find(&games)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch games"})
		return
	}

	response := make([]GameResponse, len(games))
	for i, game := range games {
		response[i] = GameResponse{
			ID:       game.ID,
			Sender:   game.Sender,
			Receiver: game.Receiver,
			Created:  game.Created,
			Status:   game.Status,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *GameHandler) calculateScores(tx *gorm.DB, game *model.Game) {
	var senderScore, receiverScore int

	switch {
	case game.Svote == 1 && game.Rvote == -1:
		senderScore, receiverScore = 0, 5
	case game.Svote == -1 && game.Rvote == 1:
		senderScore, receiverScore = 5, 0
	case game.Svote == 1 && game.Rvote == 1:
		senderScore, receiverScore = 3, 3
	case game.Svote == -1 && game.Rvote == -1:
		senderScore, receiverScore = 1, 1
	}

	tx.Exec(`
		INSERT INTO leaderboards (user_id, score, last_updated)
		VALUES (?, ?, ?), (?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			score = leaderboards.score + EXCLUDED.score,
			last_updated = EXCLUDED.last_updated
	`,
		game.Sender, senderScore, time.Now(),
		game.Receiver, receiverScore, time.Now())
}

func (h *GameHandler) sendGameNotification(userID string, data interface{}) {
	h.hub.mutex.RLock()
	defer h.hub.mutex.RUnlock()

	if client, ok := h.hub.clients[userID]; ok {
		client.Conn.WriteJSON(data)
	}
}
