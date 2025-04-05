package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID                string `gorm:"primaryKey"`
	Email             string `gorm:"unique"`
	EncryptedPassword string
	CreatedAt         time.Time
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

var db *gorm.DB
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func main() {
	// Initialize database
	var err error
	db, err = gorm.Open(sqlite.Open(os.Getenv("SQLITE_PATH")), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate schema
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Database migration failed:", err)
	}

	// Set up router
	r := gin.Default()

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(r.Run(":" + port))
}

func registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existing User
	if result := db.Where("email = ?", req.Email).First(&existing); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user
	user := User{
		ID:                generateUUID(),
		Email:             req.Email,
		EncryptedPassword: hashedPassword,
		CreatedAt:         time.Now(),
	}

	if result := db.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user User
	if result := db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	if !checkPassword(req.Password, user.EncryptedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%s|%s", base64.RawStdEncoding.EncodeToString(hash),
		base64.RawStdEncoding.EncodeToString(salt)), nil
}

func checkPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "|")
	if len(parts) != 2 {
		return false
	}

	hash, _ := base64.RawStdEncoding.DecodeString(parts[0])
	salt, _ := base64.RawStdEncoding.DecodeString(parts[1])

	newHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, newHash) == 1
}

func generateUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
