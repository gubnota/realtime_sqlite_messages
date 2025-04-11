package middleware

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

func MaxConcurrentRequests(limit int32) gin.HandlerFunc {
	var current int32 = 0
	return func(c *gin.Context) {
		if atomic.AddInt32(&current, 1) > limit {
			atomic.AddInt32(&current, -1)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "server busy",
			})
			return
		}
		defer atomic.AddInt32(&current, -1)
		c.Next()
	}
}
