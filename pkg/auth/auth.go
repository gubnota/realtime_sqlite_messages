package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"halves/pkg/model"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing model.User
	if result := s.db.Where("email = ?", req.Email).First(&existing); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	// hashedPassword, err := hashPassword(req.Password)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
	// 	return
	// }

	user := model.User{
		ID:        generateUUID(),
		Email:     req.Email,
		Password:  req.Password, //hashedPassword,
		CreatedAt: time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if result := s.db.WithContext(ctx).Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

func (s *AuthService) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if result := s.db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !checkPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Decode base64-encoded secret
	secret, err := base64.RawStdEncoding.DecodeString(os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid jwt secret configuration"})
		return
	}

	// Create token with additional claims
	exp := time.Now().Add(time.Hour * 24).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID, // Subject
		"exp": exp,     // Expiration
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_in": exp,
		"uuid":       user.ID,
	})

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	// 	"sub": user.ID,
	// 	"exp": time.Now().Add(time.Hour * 24).Unix(),
	// })

	// tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// JWT Middleware and utility functions remain the same
func (s *AuthService) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if result := s.db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		// Don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"status": "reset link sent if email exists"})
		return
	}

	// Call Python microservice
	webhookURL := os.Getenv("EMAIL_WEBHOOK")
	payload := map[string]interface{}{
		"email":      user.Email,
		"user_id":    user.ID,
		"expires_at": time.Now().Add(10 * time.Minute).Unix(),
	}
	jsonPayload, err := json.Marshal(payload)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Failed to call email webhook: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "reset link sent if email exists"})
}
