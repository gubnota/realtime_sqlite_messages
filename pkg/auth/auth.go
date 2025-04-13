package auth

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var existing string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existing)
		if err != sql.ErrNoRows {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}

		hashed, err := req.Password, nil //hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
			return
		}
		id := generateUUID()
		_, err = db.Exec(`INSERT INTO users (id, email, password, created_at, last_seen, score) VALUES (?, ?, ?, ?, ?, ?)`,
			id, req.Email, hashed, time.Now().Unix(), 0, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "email": req.Email})
	}
}

func LoginUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id, hashed string
		err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", req.Email).Scan(&id, &hashed)
		if err != nil || !checkPassword(req.Password, hashed) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		secret, err := base64.RawStdEncoding.DecodeString(os.Getenv("JWT_SECRET"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "jwt secret invalid"})
			return
		}

		exp := time.Now().Add(time.Hour * 24).Unix()
		token := generateJWT(id, exp, secret)

		c.JSON(http.StatusOK, gin.H{
			"token":      token,
			"expires_in": exp,
			"uuid":       id,
		})
	}
}

func RequestPasswordResetHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"status": "reset link sent if email exists"})
			return
		}

		payload := map[string]interface{}{
			"email":      req.Email,
			"user_id":    id,
			"expires_at": time.Now().Add(10 * time.Minute).Unix(),
		}
		body, _ := json.Marshal(payload)
		resp, err := http.Post(os.Getenv("EMAIL_WEBHOOK"), "application/json",
			bytes.NewBuffer(body))
		if err != nil || resp.StatusCode != 200 {
			log.Printf("Webhook error: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{"status": "reset link sent if email exists"})
	}
}
