package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimit creates a rate limiting middleware.
// Uses Redis when available (cluster-safe), falls back to in-memory.
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	rdb := cache.Get()
	if rdb != nil {
		return redisRateLimit(maxRequests, window)
	}
	return memoryRateLimit(maxRequests, window)
}

// redisRateLimit uses Redis INCR + EXPIRE for distributed rate limiting
func redisRateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		rdb := cache.Get()
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis error: allow request through (fail-open)
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(maxRequests) {
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", http.StatusText(maxRequests))
		c.Header("X-RateLimit-Remaining", http.StatusText(int(int64(maxRequests)-count)))
		c.Next()
	}
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// memoryRateLimit is the fallback for single-instance deployments
func memoryRateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	visitors := make(map[string]*visitor)
	mu := sync.Mutex{}

	// Cleanup goroutine with stop mechanism
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				mu.Lock()
				for ip, v := range visitors {
					if time.Since(v.lastSeen) > window {
						delete(visitors, ip)
					}
				}
				mu.Unlock()
			}
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		v, exists := visitors[ip]
		if !exists || time.Since(v.lastSeen) > window {
			visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}
		v.count++
		v.lastSeen = time.Now()
		if v.count > maxRequests {
			mu.Unlock()
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
