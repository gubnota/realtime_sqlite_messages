package main

import (
	"halves/pkg/auth"
	"halves/pkg/handler"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initialize DB
	db, err := gorm.Open(sqlite.Open(os.Getenv("SQLITE_PATH")), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create services
	authService := auth.NewAuthService(db)
	wsHub := handler.NewHub()
	messageHandler := handler.NewMessageHandler(db, wsHub)
	go wsHub.Run()

	// Create router
	r := gin.Default()

	// Add database to context middleware
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Auth routes
	r.POST("/register", authService.Register)
	r.POST("/login", authService.Login)

	// Message routes
	authMiddleware := auth.JWTMiddleware()
	r.POST("/send", authMiddleware, messageHandler.SendMessage)
	r.GET("/ws/:uuid", authMiddleware, wsHub.WebSocketHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
