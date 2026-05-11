package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"ledger-system/utils"

	"github.com/google/uuid"
)

type loginAttempt struct {
	count int
	first time.Time
}

type ipRateLimiter struct {
	mu          sync.Mutex
	requests    map[string]*loginAttempt
	maxAttempts int
	window      time.Duration
}

func newIPRateLimiter(maxAttempts int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		requests:    make(map[string]*loginAttempt),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

// Limits the request rate based on the src ip for a fixed windows
func (l *ipRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	attempt, ok := l.requests[ip]
	if !ok || now.Sub(attempt.first) > l.window {
		l.requests[ip] = &loginAttempt{count: 1, first: now}
		return true
	}

	if attempt.count >= l.maxAttempts {
		return false
	}

	attempt.count++
	return true
}

// func (l *ipRateLimiter) Reset(ip string) {
// 	l.mu.Lock()
// 	defer l.mu.Unlock()
// 	delete(l.requests, ip)
// }

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return ip
}

var LoginRateLimiter = newIPRateLimiter(5, 5*time.Minute)

// JWTMiddleware validates JWT tokens and adds user ID to request context
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Invalid authorization header format. Use 'Bearer <token>'"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(w, `{"error": "Token is required"}`, http.StatusUnauthorized)
			return
		}

		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid token: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// RateLimitMiddleware limits the number of requests per IP for login attempts
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		if !LoginRateLimiter.Allow(clientIP) {
			http.Error(w, "Too many login attempts. Please try again in 5 minutes.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	return userID, ok
}
