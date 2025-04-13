package auth

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware() gin.HandlerFunc {
	secretString := os.Getenv("JWT_SECRET")

	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		// if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		// }

		secret, err := base64.RawStdEncoding.DecodeString(secretString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server config error"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
					return
				}
			}

			if sub, ok := claims["sub"].(string); ok {
				c.Set("userID", sub)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
	}
}

func LastSeenUpdater() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		db := c.MustGet("db").(*sql.DB)
		deviceID := c.GetHeader("X-Device-ID")
		now := time.Now().Unix()

		_, _ = db.Exec("UPDATE users SET last_seen = ? WHERE id = ?", now, userID)

		if deviceID != "" {
			_, _ = db.Exec("UPDATE devices SET last_seen = ?, status = 'O' WHERE id = ?", now, deviceID)
		}

		c.Next()
	}
}
