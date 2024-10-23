// backend/internal/api/middlewares/ratelimit.go

package middlewares

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// RateLimiter defines the structure for rate limiting
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.Mutex
	r        rate.Limit
	b        int
	logger   *logrus.Logger
}

// Visitor holds the rate limiter and last seen time for a client
type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter initializes a new RateLimiter
func NewRateLimiter(r rate.Limit, b int, logger *logrus.Logger) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		r:        r,
		b:        b,
		logger:   logger,
	}

	go rl.cleanupVisitors()
	return rl
}

// getVisitor retrieves or creates a rate limiter for a given IP
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.r, rl.b)
		rl.visitors[ip] = &Visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes visitors that haven't been seen for over 3 minutes
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware enforces rate limiting based on client IP
func RateLimitMiddleware(rl *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			limiter := rl.getVisitor(ip)

			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the client's IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first (if behind a proxy)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, the first one is the client's IP
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}
