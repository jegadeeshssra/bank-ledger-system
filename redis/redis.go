// Package redisstore provides Redis helpers used by the application, including a Redis-backed
// token bucket rate limiter for login and registration requests.
package redisstore

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// LoginCapacity is the maximum number of login requests allowed within the token bucket.
	LoginCapacity = 3

	// LoginWindow is the base time window used for both the login throttling period and
	// token bucket expiration.
	LoginWindow = 5 * time.Minute

	// LoginRefillRate is the number of login tokens added per second to the bucket.
	// It is computed from LoginCapacity over the full LoginWindow.
	LoginRefillRate = 0.2

	// RegisterCapacity is the maximum number of registration requests allowed per IP within the token bucket.
	RegisterCapacity = 1

	// RegisterWindow is the time window for registration rate limiting (1 hour).
	RegisterWindow = 1 * time.Hour

	// RegisterRefillRate is the number of registration tokens added per second to the bucket.
	// It is computed from RegisterCapacity over the full RegisterWindow.
	RegisterRefillRate = 0.001

	// Authenticated request rate limit values by HTTP method.
	AuthWindow = 1 * time.Minute

	GetCapacity      = 60
	GetRefillRate    = float64(GetCapacity) / 60.0
	PostCapacity     = 30
	PostRefillRate   = float64(PostCapacity) / 60.0
	PutCapacity      = 20
	PutRefillRate    = float64(PutCapacity) / 60.0
	DeleteCapacity   = 10
	DeleteRefillRate = float64(DeleteCapacity) / 60.0

	// Endpoint-specific capacities for authenticated routes.
	GetAccountsCapacity    = 40
	GetAccountCapacity     = 30
	GetTransactionCapacity = 20
	GetEntriesCapacity     = 25
	PutReconcileCapacity   = 15
	PostAccountsCapacity   = 15
	PostDepositCapacity    = 20
	PostWithdrawCapacity   = 20
	PostTransfersCapacity  = 10
	DeleteAccountCapacity  = 5
)

// RateLimitConfig holds the configuration for a rate limiter endpoint.
type RateLimitConfig struct {
	// Capacity is the maximum number of tokens in the bucket.
	Capacity float64
	// Window is the time window over which tokens refill.
	Window time.Duration
	// RefillRate is the number of tokens added per second.
	RefillRate float64
	// KeyPrefix is the Redis key prefix (e.g., "rate_limit", "register_limit").
	KeyPrefix string
}

// LoginConfig returns the rate limit config for login requests.
func LoginConfig() RateLimitConfig {
	return RateLimitConfig{
		Capacity:   LoginCapacity,
		Window:     LoginWindow,
		RefillRate: LoginRefillRate,
		KeyPrefix:  "rate_limit",
	}
}

// RegisterConfig returns the rate limit config for registration requests.
func RegisterConfig() RateLimitConfig {
	return RateLimitConfig{
		Capacity:   RegisterCapacity,
		Window:     RegisterWindow,
		RefillRate: RegisterRefillRate,
		KeyPrefix:  "register_limit",
	}
}

// AuthConfig returns the rate limit config for authenticated API requests based on HTTP method and endpoint.
func AuthConfig(method, endpoint string) RateLimitConfig {
	method = strings.ToUpper(method)
	endpoint = strings.ToLower(endpoint)

	switch method {
	case "GET":
		switch endpoint {
		case "getaccounts":
			return RateLimitConfig{Capacity: GetAccountsCapacity, Window: AuthWindow, RefillRate: float64(GetAccountsCapacity) / 60.0, KeyPrefix: "auth_get_accounts_limit"}
		case "getaccount":
			return RateLimitConfig{Capacity: GetAccountCapacity, Window: AuthWindow, RefillRate: float64(GetAccountCapacity) / 60.0, KeyPrefix: "auth_get_account_limit"}
		case "gettransaction":
			return RateLimitConfig{Capacity: GetTransactionCapacity, Window: AuthWindow, RefillRate: float64(GetTransactionCapacity) / 60.0, KeyPrefix: "auth_get_transaction_limit"}
		case "getentries":
			return RateLimitConfig{Capacity: GetEntriesCapacity, Window: AuthWindow, RefillRate: float64(GetEntriesCapacity) / 60.0, KeyPrefix: "auth_get_entries_limit"}
		default:
			return RateLimitConfig{Capacity: GetCapacity, Window: AuthWindow, RefillRate: GetRefillRate, KeyPrefix: "auth_get_limit"}
		}
	case "POST":
		switch endpoint {
		case "createaccount":
			return RateLimitConfig{Capacity: PostAccountsCapacity, Window: AuthWindow, RefillRate: float64(PostAccountsCapacity) / 60.0, KeyPrefix: "auth_post_accounts_limit"}
		case "deposit":
			return RateLimitConfig{Capacity: PostDepositCapacity, Window: AuthWindow, RefillRate: float64(PostDepositCapacity) / 60.0, KeyPrefix: "auth_post_deposit_limit"}
		case "withdraw":
			return RateLimitConfig{Capacity: PostWithdrawCapacity, Window: AuthWindow, RefillRate: float64(PostWithdrawCapacity) / 60.0, KeyPrefix: "auth_post_withdraw_limit"}
		case "transfer":
			return RateLimitConfig{Capacity: PostTransfersCapacity, Window: AuthWindow, RefillRate: float64(PostTransfersCapacity) / 60.0, KeyPrefix: "auth_post_transfers_limit"}
		default:
			return RateLimitConfig{Capacity: PostCapacity, Window: AuthWindow, RefillRate: PostRefillRate, KeyPrefix: "auth_post_limit"}
		}
	case "PUT":
		switch endpoint {
		case "reconcile":
			return RateLimitConfig{Capacity: PutReconcileCapacity, Window: AuthWindow, RefillRate: float64(PutReconcileCapacity) / 60.0, KeyPrefix: "auth_put_reconcile_limit"}
		default:
			return RateLimitConfig{Capacity: PutCapacity, Window: AuthWindow, RefillRate: PutRefillRate, KeyPrefix: "auth_put_limit"}
		}
	case "DELETE":
		switch endpoint {
		case "deleteaccount":
			return RateLimitConfig{Capacity: DeleteAccountCapacity, Window: AuthWindow, RefillRate: float64(DeleteAccountCapacity) / 60.0, KeyPrefix: "auth_delete_account_limit"}
		default:
			return RateLimitConfig{Capacity: DeleteCapacity, Window: AuthWindow, RefillRate: DeleteRefillRate, KeyPrefix: "auth_delete_limit"}
		}
	default:
		return RateLimitConfig{Capacity: GetCapacity, Window: AuthWindow, RefillRate: GetRefillRate, KeyPrefix: "auth_default_limit"}
	}
}

// NewRedisClient constructs a Redis client configured for local development.
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "RedisPassword",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

// AllowRequest checks whether the given identifier is allowed to make a request based on the provided rate limit config.
//
// This function implements a Redis-backed token bucket:
//   - Each identifier starts with config.Capacity tokens.
//   - Tokens refill over time at config.RefillRate per second.
//   - A request consumes one token if available.
//   - If no tokens remain, the request is denied.
//
// The Redis hash stores the remaining token count and the last refill timestamp.
// The config parameter allows dynamic handling of different endpoints (login, register, etc.) with different limits.
func AllowRequest(ctx context.Context, rdb *redis.Client, config RateLimitConfig, identifier string) (bool, error) {
	key := config.KeyPrefix + ":" + identifier
	now := time.Now().Unix()

	data, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Start with a full bucket by default.
	tokens := config.Capacity
	lastRefill := now

	if len(data) != 0 {
		parsedTokens, err := strconv.ParseFloat(data["tokens"], 64)
		if err == nil {
			tokens = parsedTokens
		}

		parsedLastRefill, err := strconv.ParseInt(data["last_refill"], 10, 64)
		if err == nil {
			lastRefill = parsedLastRefill
		}

		// Guard against clocks or stale data reporting a future refill time.
		if lastRefill > now {
			lastRefill = now
		}

		elapsed := float64(now - lastRefill)
		tokens += elapsed * config.RefillRate
		if tokens > config.Capacity {
			tokens = config.Capacity
		}
		lastRefill = now
	}

	// Deny if there are no tokens available.
	if tokens < 1 {
		return false, nil
	}

	// Consume one token for the current request.
	tokens--

	err = rdb.HSet(ctx, key,
		"tokens", strconv.FormatFloat(tokens, 'f', -1, 64),
		"last_refill", strconv.FormatInt(lastRefill, 10),
	).Err()
	if err != nil {
		return false, err
	}

	// Keep the rate limit record alive for a little longer than the window so that
	// identifiers that stop making requests are eventually removed from Redis.
	rdb.Expire(ctx, key, config.Window+time.Minute)
	return true, nil
}
