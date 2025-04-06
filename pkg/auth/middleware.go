// pkg/auth/middleware.go
package auth

import (
	"encoding/base64"
	"fmt"
	"halves/pkg/model"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func JWTMiddleware() gin.HandlerFunc {
	secretString := os.Getenv("JWT_SECRET")
	return func(c *gin.Context) {
		// 1. Get Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// 2. Extract Bearer token
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		// 3. Decode JWT secret
		secret, err := base64.RawStdEncoding.DecodeString(secretString)
		if err != nil || len(secret) < 32 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Server configuration error",
			})
			return
		}

		// 4. Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		// 5. Handle parsing errors
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// 6. Validate claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Check expiration
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "Token expired",
					})
					return
				}
			}

			// Set user ID in context
			if sub, ok := claims["sub"].(string); ok {
				c.Set("userID", sub)
				c.Next()
				return
			}
		}

		// 7. Fallback error
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token claims",
		})
	}
}
func LastSeenUpdater() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID, exists := c.Get("userID"); exists {
			db := c.MustGet("db").(*gorm.DB)
			now := time.Now().Unix()

			// Update user's last_seen
			db.Model(&model.User{}).
				Where("id = ?", userID).
				Update("last_seen", now)

			// Update device status if available
			if deviceID := c.GetHeader("X-Device-ID"); deviceID != "" {
				db.Model(&model.Device{}).
					Where("id = ?", deviceID).
					Updates(map[string]interface{}{
						"last_seen": now,
						"status":    "online",
					})
			}
		}
		c.Next()
	}
}
