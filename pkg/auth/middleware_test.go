package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-1234567890-abcd-1234-567890")
	defer os.Unsetenv("JWT_SECRET")

	t.Run("Missing Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		middleware := JWTMiddleware()
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "missing token")
	})

	t.Run("Valid Token", func(t *testing.T) {
		// Generate valid token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		middleware := JWTMiddleware()
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, "user123", c.GetString("userID"))
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		// Generate token with different secret
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("wrong-secret"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		middleware := JWTMiddleware()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid token")
	})
}
