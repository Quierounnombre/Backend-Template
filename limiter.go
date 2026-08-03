package main

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (rl *RateLimiter)Allow(client string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[client]
	if !ok {
		b = &Bucket{tokens: rl.max}
		rl.buckets[client] = b
	}
	if b.tokens == 0 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter)Middleware() gin.HandlerFunc{
	return func(c *gin.Context) {
		//Only would exist on local testing endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/internal/") {
			c.Next()
			return
		}
		if !rl.Allow(c.ClientIP()) {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}

func (rl *RateLimiter)Refill() {
	ticker := time.NewTicker(rl.reset_time)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for key, b := range rl.buckets {
			if b.tokens > rl.max {
				delete(rl.buckets, key)
				continue
			}
			b.tokens += rl.ref_amount
		}
		rl.mu.Unlock()
		slog.Info("rate limit: reset")
	}
}
