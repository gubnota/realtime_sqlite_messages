package main

import (
	"halves/pkg/auth"
	"halves/pkg/handler"
	"halves/pkg/model"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	go func() {
		log.Println("Starting pprof on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// Load .env from the same directory as the binary
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	exeDir := filepath.Dir(exePath)

	err = godotenv.Load(filepath.Join(exeDir, ".env"))
	if err != nil {
		log.Println("No .env file found in binary directory (optional).")
	}
	// Configure SQLite path
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "halves.db" // Default database file
	}

	// Create database directory if needed
	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0755); err != nil {
		log.Fatal("Failed to create database directory:", err)
	}
	log.Println("Database path:", sqlitePath)
	// Initialize DB
	db, err := gorm.Open(sqlite.Open(os.Getenv("SQLITE_PATH")), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// Add automigrate after DB connection
	err = db.AutoMigrate(
		&model.User{},
		&model.Message{},
		&model.Device{},
		&model.Game{},
		&model.Result{},
	)

	// In main.go, replace the device reset code with:
	if err := db.Exec("UPDATE devices SET status = 'F'").Error; err != nil {
		log.Println("Failed to reset device statuses:", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite requires single connection

	// Create services
	authService := auth.NewAuthService(db)
	wsHub := handler.NewHub()
	messageHandler := handler.NewMessageHandler(db, wsHub)
	userHandler := handler.NewUserHandler(db)
	gameHandler := handler.NewGameHandler(db, wsHub)
	resultHandler := handler.NewReslutHandler(db)

	go wsHub.Run()

	// Create router
	r := gin.New()
	r.Use(gin.Recovery()) // ✅ Panic recovery middleware
	// r.Use(middleware.MaxConcurrentRequests(1)) // ✅ Reject excess load

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
	lastSeenMiddleware := auth.LastSeenUpdater()
	r.POST("/send", authMiddleware, lastSeenMiddleware, messageHandler.SendMessage)
	r.GET("/ws/:uuid", authMiddleware, lastSeenMiddleware, wsHub.WebSocketHandler)
	// r.GET("/tws/:uuid", wsHub.WebSocketHandler) // Without auth middleware
	r.GET("/messages", authMiddleware, lastSeenMiddleware, messageHandler.GetMessages)
	// Add to routes
	r.POST("/reset-password", authService.RequestPasswordReset)
	r.POST("/reset-password/confirm", authService.ResetPassword)
	r.POST("/update-score", authMiddleware, userHandler.UpdateScore)
	r.GET("/health", func(c *gin.Context) {
		if err := db.Exec("SELECT 1").Error; err != nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.Status(http.StatusOK)
	})
	r.POST("/game/invite", authMiddleware, gameHandler.CreateGame)
	r.POST("/game/vote", authMiddleware, gameHandler.HandleVote)
	r.GET("/games/active", authMiddleware, gameHandler.GetActiveGames)
	r.GET("/result", resultHandler.GetResult)
	r.DELETE("/users", userHandler.DeleteUsers)
	// Add periodic cleanup task (after route setup)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			threshold := time.Now().Add(-1 * time.Hour).Unix()
			db.Model(&model.Device{}).
				Where("last_seen < ? AND status = 'O'", threshold).
				Update("status", "F")
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
