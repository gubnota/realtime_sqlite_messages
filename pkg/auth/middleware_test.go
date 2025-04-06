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
	os.Setenv("JWT_SECRET", "TlFuT3JUMWNXano4N2pVN0FmU3BuamRUdFNTTzAzMndBQzRmN1BBemtlbz0K")
	defer os.Unsetenv("JWT_SECRET")

	t.Run("login token check", func(t *testing.T) {
		// Create token with additional claims
		exp := time.Now().Add(time.Hour * 24).Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "2c71da62-a057-4c24-beac-114b8e5d0dff", // Subject
			"exp": exp,                                    // Expiration
		})

		tokenString, err := token.SignedString(os.Getenv("JWT_SECRET"))
		if err != nil {
			return
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		middleware := JWTMiddleware()
		middleware(c)
		print(tokenString)
		assert.Equal(t, http.StatusOK, w.Code)

	})

	// t.Run("Valid Token", func(t *testing.T) {
	// 	w := httptest.NewRecorder()
	// 	c, _ := gin.CreateTestContext(w)
	// 	c.Request = &http.Request{
	// 		Header: make(http.Header),
	// 	}

	// 	// Generate a real JWT token
	// 	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	// 	// 	"sub": "1234567890",
	// 	// 	"exp": time.Now().Add(time.Hour * 24).Unix(),
	// 	// })
	// 	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQwMTYzNDIsInN1YiI6IjJjNzFkYTYyLWEwNTctNGMyNC1iZWFjLTExNGI4ZTVkMGRmZiJ9.8lFF_wHWYEtog69X2LA3MASDg9hISZqCfx-t3x7M_Xk" //token.SignedString([]byte(os.Getenv("JWT_SECRET"))) , err
	// 	// assert.NoError(t, err)

	// 	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// 	middleware := JWTMiddleware()
	// 	middleware(c)

	// 	assert.Equal(t, http.StatusOK, w.Code)
	// 	assert.Equal(t, "2c71da62-a057-4c24-beac-114b8e5d0dff", c.MustGet("userID"))
	// })

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
