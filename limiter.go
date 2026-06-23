package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (rl *RateLimiter)Allow(client string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.reqs[client]++
	if rl.reqs[client] > rl.Max_reqs {
		return false
	}
	return true
}

func (rl *RateLimiter)Middleware() gin.HandlerFunc{
	return func(c *gin.Context) {
		if !rl.Allow(c.ClientIP()) {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}

func (rl *RateLimiter)Cleanup() {
	ticker := time.NewTicker(rl.Reset_time)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		rl.reqs = make(map[string]uint)
		rl.mu.Unlock()
		log.Print("rate limit: reset")
	}
}
