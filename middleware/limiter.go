package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// var limiter = rate.NewLimiter(1, 5) // 1 req/sec with burst of 5
var visitors = make(map[string]*rate.Limiter)

func getLimiter(ip string) *rate.Limiter {
	if l, exists := visitors[ip]; exists {
		return l
	}
	limiter := rate.NewLimiter(1, 5)
	visitors[ip] = limiter
	return limiter
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !getLimiter(c.RemoteIP()).Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
