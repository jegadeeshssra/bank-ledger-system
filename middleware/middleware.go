package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	redisstore "ledger-system/redis"
	"ledger-system/utils"

	"github.com/google/uuid"
)

type loginRequestPayload struct {
	Username string `json:"username"`
}

var RedisClient = redisstore.NewRedisClient()

func extractLoginUsername(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var payload loginRequestPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return "", nil
	}

	return payload.Username, nil
}

// JWTMiddleware validates JWT tokens and adds user ID to request context
func extractClientIP(r *http.Request) (string, error) {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip, nil
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if r.RemoteAddr != "" {
			return r.RemoteAddr, nil
		}
		return "", err
	}
	return host, nil
}

func extractUserIdentifier(r *http.Request) (string, error) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		return "", fmt.Errorf("user ID missing in context")
	}
	return userID.String(), nil
}

func extractLoginIdentifier(r *http.Request) (string, error) {
	username, err := extractLoginUsername(r)
	if err != nil {
		return "", err
	}
	return username, nil
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	return userID, ok
}

// LoginRateLimitMiddleware enforces a per-username token bucket on the login endpoint.
func LoginRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return rateLimitMiddleware(redisstore.LoginConfig(), extractLoginIdentifier, "Too many login attempts for this user. Please try again in 5 minutes.")(next)
}

// RegisterRateLimitMiddleware enforces a per-IP token bucket on the register endpoint.
func RegisterRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return rateLimitMiddleware(redisstore.RegisterConfig(), extractClientIP, "Too many registration attempts from this IP. Please try again later.")(next)
}

func AuthRateLimitMiddleware(method, endpoint string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			identifier, err := extractUserIdentifier(r)
			if err != nil {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			config := redisstore.AuthConfig(method, endpoint)
			allowed, err := redisstore.AllowRequest(r.Context(), RedisClient, config, identifier)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, fmt.Sprintf(`{"error": "Rate limit exceeded for %s %s. Try again later."}`, strings.ToUpper(method), endpoint), http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

func rateLimitMiddleware(config redisstore.RateLimitConfig, extractIdentifier func(*http.Request) (string, error), errorMessage string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			identifier, err := extractIdentifier(r)
			if err != nil || identifier == "" {
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				next(w, r)
				return
			}

			allowed, err := redisstore.AllowRequest(r.Context(), RedisClient, config, identifier)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, errorMessage, http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

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

		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
