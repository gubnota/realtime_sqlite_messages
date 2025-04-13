package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"halves/pkg/auth"
	dbi "halves/pkg/db"
	"halves/pkg/handler"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	// Load .env from the same directory as the binary
	// exePath, err := os.Executable()
	// if err != nil {
	// 	log.Fatal("Failed to get executable path:", err)
	// }
	// exeDir := filepath.Dir(exePath)

	// err = godotenv.Load(filepath.Join(exeDir, ".env"))
	// if err != nil {
	// 	log.Println("No .env file found in binary directory (optional).")
	// }
	// load .env from current dir
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found in current directory (optional).")
	}
	// Configure SQLite path
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "data/sq.db" // Default database file
	}

	// Create database directory if needed
	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0755); err != nil {
		log.Fatal("Failed to create database directory:", err)
	}
	log.Println("Database path:", sqlitePath)
	// Create tables if they don't exist
	dbi.CreateTablesSQL(sqlitePath)

	// Initialize DB using modernc.org/sqlite
	db, err = sql.Open("sqlite", sqlitePath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Optional PRAGMA tuning
	// db.Exec(`PRAGMA journal_mode = WAL`)
	// db.Exec(`PRAGMA foreign_keys = ON`)

	// Reset all device statuses
	if _, err := db.Exec("UPDATE devices SET status = 'F'"); err != nil {
		log.Println("Failed to reset device statuses:", err)
	}

	// Create services
	// authService := auth.NewAuthService(db)
	wsHub := handler.NewHub()
	messageHandler := handler.NewMessageHandler(db, wsHub)
	userHandler := handler.NewUserHandler(db)
	gameHandler := handler.NewGameHandler(db, wsHub)
	resultHandler := handler.NewReslutHandler(db)

	go wsHub.Run()

	// Create router
	r := gin.New()
	r.Use(gin.Recovery())
	// r.Use(middleware.MaxConcurrentRequests(1))

	// Add database to context middleware
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Auth routes
	r.POST("/register", auth.RegisterUserHandler(db))
	r.POST("/login", auth.LoginUserHandler(db))
	r.POST("/reset-password", auth.RequestPasswordResetHandler(db))
	// Message routes
	authMiddleware := auth.JWTMiddleware()
	lastSeenMiddleware := auth.LastSeenUpdater()
	r.POST("/send", authMiddleware, lastSeenMiddleware, messageHandler.SendMessage)
	r.GET("/ws/:uuid", authMiddleware, lastSeenMiddleware, wsHub.WebSocketHandler)
	r.GET("/messages", authMiddleware, lastSeenMiddleware, messageHandler.GetMessages)
	r.POST("/update-score", authMiddleware, userHandler.UpdateScore)
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.Status(http.StatusOK)
	})
	r.POST("/game/invite", authMiddleware, gameHandler.CreateGame)
	r.POST("/game/vote", authMiddleware, gameHandler.HandleVote)
	r.GET("/games/active", authMiddleware, gameHandler.GetActiveGames)
	r.GET("/result", resultHandler.GetResult)

	// Add periodic cleanup task
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			threshold := time.Now().Add(-1 * time.Hour).Unix()
			_, err := db.Exec("UPDATE devices SET status = 'F' WHERE last_seen < ? AND status = 'O'", threshold)
			if err != nil {
				log.Println("Cleanup failed:", err)
			}
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
