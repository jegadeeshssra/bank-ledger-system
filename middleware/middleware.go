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
	"time"

	"ledger-system/config"
	redisstore "ledger-system/redis"
	"ledger-system/utils"

	"github.com/google/uuid"
)

type loginRequestPayload struct {
	Username string `json:"username"`
}

var RedisClient = redisstore.NewRedisClient()

var defaultAllowedOrigins = map[string]struct{}{}

func parseAllowedOrigins() map[string]struct{} {
	configured := config.GetString("CORS_ALLOWED_ORIGINS", "")
	if configured == "" {
		return defaultAllowedOrigins
	}

	origins := make(map[string]struct{})
	for _, origin := range strings.Split(configured, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			origins[trimmed] = struct{}{}
		}
	}

	if len(origins) == 0 {
		return defaultAllowedOrigins
	}

	return origins
}

func isOriginAllowed(origin string, allowed map[string]struct{}) bool {
	_, ok := allowed[origin]
	return ok
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		//w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Idempotency-Key")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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

func normalizeRequestBody(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, body); err == nil {
		return buf.String()
	}

	return string(body)
}

type idempotencyResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rw *idempotencyResponseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *idempotencyResponseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func IdempotencyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempotencyKey := r.Header.Get("X-Idempotency-Key")
		if idempotencyKey == "" {
			http.Error(w, `{"error": "Idempotency-Key header required"}`, http.StatusBadRequest)
			return
		}

		if _, err := uuid.Parse(idempotencyKey); err != nil {
			http.Error(w, `{"error": "Invalid Idempotency-Key format. Must be a UUID."}`, http.StatusBadRequest)
			return
		}

		userID, err := extractUserIdentifier(r)
		if err != nil {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error": "Failed to read request body"}`, http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		normalizedBody := normalizeRequestBody(bodyBytes)
		redisKey := fmt.Sprintf("idempotent:%s:%s", userID, normalizedBody)

		cached, err := redisstore.GetStringValue(r.Context(), RedisClient, redisKey)
		if err != nil {
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if cached != "" {
			var cachedResponse struct {
				Status int    `json:"status"`
				Body   string `json:"body"`
			}
			if err := json.Unmarshal([]byte(cached), &cachedResponse); err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cachedResponse.Status)
				_, _ = w.Write([]byte(cachedResponse.Body))
				return
			}
		}

		recorder := &idempotencyResponseWriter{ResponseWriter: w}
		next(recorder, r)

		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}
		if recorder.status < 200 || recorder.status >= 300 {
			return
		}

		if recorder.body.Len() == 0 {
			return
		}

		cachedPayload, err := json.Marshal(map[string]any{
			"status": recorder.status,
			"body":   recorder.body.String(),
		})
		if err != nil {
			return
		}

		_, _ = redisstore.SetNXString(r.Context(), RedisClient, redisKey, string(cachedPayload), 24*time.Hour)
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
