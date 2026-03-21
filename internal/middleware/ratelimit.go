package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	ipLimiters = make(map[string]*ipEntry)
	ipMu       sync.Mutex
)

func init() {
	go cleanupIPLimiters()
}

// cleanupIPLimiters removes entries that have been idle for more than 5 minutes.
func cleanupIPLimiters() {
	for {
		time.Sleep(time.Minute)
		ipMu.Lock()
		for ip, e := range ipLimiters {
			if time.Since(e.lastSeen) > 5*time.Minute {
				delete(ipLimiters, ip)
			}
		}
		ipMu.Unlock()
	}
}

func getLimiterForIP(ip string) *rate.Limiter {
	ipMu.Lock()
	defer ipMu.Unlock()

	e, exists := ipLimiters[ip]
	if !exists {
		// 60 requests per minute: 1 token/second, burst of 60.
		e = &ipEntry{limiter: rate.NewLimiter(rate.Every(time.Minute/60), 60)}
		ipLimiters[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

// RateLimit returns a Gin middleware that enforces 60 req/min per IP.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !getLimiterForIP(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
