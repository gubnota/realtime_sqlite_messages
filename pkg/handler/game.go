package handler

import (
	"database/sql"
	"halves/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	db  *sql.DB
	hub *Hub
}

type GameDTO struct {
	ID       int64     `json:"id"`
	Sender   string    `json:"sender"`
	Receiver string    `json:"receiver"`
	Created  time.Time `json:"created"`
}

type GameResponse struct {
	ID       int64     `json:"id"`
	Sender   string    `json:"sender"`
	Receiver string    `json:"receiver"`
	Created  time.Time `json:"created"`
	Status   string    `json:"status"`
}

type ReslutHandler struct {
	db *sql.DB
}

func NewGameHandler(db *sql.DB, hub *Hub) *GameHandler {
	return &GameHandler{db: db, hub: hub}
}

func NewReslutHandler(db *sql.DB) *ReslutHandler {
	return &ReslutHandler{db: db}
}

func (h *GameHandler) CreateGame(c *gin.Context) {
	var req struct {
		Receiver string `json:"receiver" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sender := c.MustGet("userID").(string)
	created := time.Now()

	result, err := h.db.Exec(`INSERT INTO games (sender, receiver, created, status, svote, rvote) VALUES (?, ?, ?, 'open', 0, 0)`, sender, req.Receiver, created)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game"})
		return
	}
	gameID, _ := result.LastInsertId()

	game := GameDTO{
		ID:       gameID,
		Sender:   sender,
		Receiver: req.Receiver,
		Created:  created,
	}

	h.sendGameNotification(req.Receiver, gin.H{"type": "game_invite", "game": game})
	go h.gameTimeoutWorker(gameID)

	c.JSON(http.StatusCreated, game)
}

func (h *GameHandler) HandleVote(c *gin.Context) {
	var req struct {
		GameID int64 `json:"game_id"`
		Vote   int   `json:"vote" binding:"required,oneof=-1 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("userID").(string)

	var game model.Game
	err := h.db.QueryRow(`SELECT id, sender, receiver, svote, rvote, status, created FROM games WHERE id = ?`, req.GameID).
		Scan(&game.ID, &game.Sender, &game.Receiver, &game.Svote, &game.Rvote, &game.Status, &game.Created)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	if game.Status != "open" {
		c.JSON(http.StatusForbidden, gin.H{"error": "game already closed"})
		return
	}

	field := ""
	switch {
	case userID == game.Sender && game.Svote == 0:
		field = "svote"
	case userID == game.Receiver && game.Rvote == 0:
		field = "rvote"
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "already voted or invalid"})
		return
	}

	_, err = h.db.Exec("UPDATE games SET "+field+" = ? WHERE id = ?", req.Vote, req.GameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vote failed"})
		return
	}

	// re-fetch game votes
	err = h.db.QueryRow("SELECT svote, rvote FROM games WHERE id = ?", req.GameID).Scan(&game.Svote, &game.Rvote)
	if err == nil && game.Svote != 0 && game.Rvote != 0 {
		h.db.Exec("UPDATE games SET status = 'closed' WHERE id = ?", req.GameID)
		h.calculateScores(&game)
		h.sendGameNotification(game.Sender, gin.H{"type": "game_result", "game": game})
		h.sendGameNotification(game.Receiver, gin.H{"type": "game_result", "game": game})
	}

	c.JSON(http.StatusOK, game)
}

func (h *GameHandler) calculateScores(game *model.Game) {
	senderScore, receiverScore := 0, 0
	switch {
	case game.Svote == 1 && game.Rvote == -1:
		receiverScore = 5
	case game.Svote == -1 && game.Rvote == 1:
		senderScore = 5
	case game.Svote == 1 && game.Rvote == 1:
		senderScore, receiverScore = 3, 3
	case game.Svote == -1 && game.Rvote == -1:
		senderScore, receiverScore = 1, 1
	}

	now := time.Now()
	h.db.Exec(`
	INSERT INTO results (user_id, score, last_updated) VALUES (?, ?, ?), (?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET score = score + excluded.score, last_updated = excluded.last_updated`,
		game.Sender, senderScore, now,
		game.Receiver, receiverScore, now)
}

func (h *GameHandler) GetActiveGames(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	rows, err := h.db.Query("SELECT id, sender, receiver, created, status FROM games WHERE (sender = ? OR receiver = ?) AND status = 'open'", userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	var games []GameResponse
	for rows.Next() {
		var g GameResponse
		if err := rows.Scan(&g.ID, &g.Sender, &g.Receiver, &g.Created, &g.Status); err == nil {
			games = append(games, g)
		}
	}
	c.JSON(http.StatusOK, games)
}

func (h *GameHandler) gameTimeoutWorker(gameID int64) {
	time.AfterFunc(2*time.Hour, func() {
		var status string
		err := h.db.QueryRow("SELECT status FROM games WHERE id = ?", gameID).Scan(&status)
		if err != nil || status != "open" {
			return
		}
		h.db.Exec("UPDATE games SET status = 'closed', rvote = 1 WHERE id = ?", gameID)
	})
}

func (h *GameHandler) sendGameNotification(userID string, data any) {
	h.hub.mutex.RLock()
	defer h.hub.mutex.RUnlock()
	if client, ok := h.hub.clients[userID]; ok {
		client.Conn.WriteJSON(data)
	}
}

func (h *ReslutHandler) GetResult(c *gin.Context) {
	rows, err := h.db.Query("SELECT user_id, score, last_updated FROM results ORDER BY score DESC LIMIT 100")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
		return
	}
	defer rows.Close()

	var result []model.Result
	for rows.Next() {
		var r model.Result
		if err := rows.Scan(&r.UserID, &r.Score, &r.LastUpdated); err == nil {
			result = append(result, r)
		}
	}
	c.JSON(http.StatusOK, result)
}
