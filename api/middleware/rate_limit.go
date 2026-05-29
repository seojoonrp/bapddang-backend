// api/middleware/rate_limit.go
// IP 기반 인메모리 rate limiter

package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/apperr"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientLimiter
	rate     rate.Limit
	burst    int
	idleTime time.Duration
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	rl := &ipRateLimiter{
		clients:  make(map[string]*clientLimiter),
		rate:     r,
		burst:    burst,
		idleTime: 10 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *ipRateLimiter) get(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cl, ok := rl.clients[ip]
	if !ok {
		cl = &clientLimiter{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.clients[ip] = cl
	}
	cl.lastSeen = time.Now()
	return cl.limiter
}

func (rl *ipRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, cl := range rl.clients {
			if time.Since(cl.lastSeen) > rl.idleTime {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware that limits to r requests per second with allowed burst, keyed by client IP
func RateLimitByIP(r rate.Limit, burst int) gin.HandlerFunc {
	rl := newIPRateLimiter(r, burst)
	return func(c *gin.Context) {
		if !rl.get(c.ClientIP()).Allow() {
			c.Error(apperr.TooManyRequests("too many requests", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}
