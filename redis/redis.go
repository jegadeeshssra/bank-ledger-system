// Package redisstore provides Redis helpers used by the application, including a Redis-backed
// token bucket rate limiter for login and registration requests.
package redisstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ledger-system/config"

	"github.com/redis/go-redis/v9"
)

const (
	defaultLoginCapacity = 3
	defaultLoginWindow   = 5 * time.Minute

	defaultRegisterCapacity = 1
	defaultRegisterWindow   = 1 * time.Hour

	// Authenticated request rate limit values by HTTP method.
	defaultAuthWindow = 1 * time.Minute

	defaultGetCapacity    = 60
	defaultPostCapacity   = 30
	defaultPutCapacity    = 20
	defaultDeleteCapacity = 10

	defaultGetAccountsCapacity    = 40
	defaultGetAccountCapacity     = 30
	defaultGetTransactionCapacity = 20
	defaultGetEntriesCapacity     = 25
	defaultPutReconcileCapacity   = 15
	defaultPostAccountsCapacity   = 15
	defaultPostDepositCapacity    = 20
	defaultPostWithdrawCapacity   = 20
	defaultPostTransfersCapacity  = 10
	defaultDeleteAccountCapacity  = 5
)

var (
	LoginCapacity   = float64(config.GetInt("LOGIN_CAPACITY", defaultLoginCapacity))
	LoginWindow     = config.GetDuration("LOGIN_WINDOW", defaultLoginWindow)
	LoginRefillRate = float64(LoginCapacity) / LoginWindow.Seconds()

	RegisterCapacity   = float64(config.GetInt("REGISTER_CAPACITY", defaultRegisterCapacity))
	RegisterWindow     = config.GetDuration("REGISTER_WINDOW", defaultRegisterWindow)
	RegisterRefillRate = float64(RegisterCapacity) / RegisterWindow.Seconds()

	AuthWindow = config.GetDuration("AUTH_WINDOW", defaultAuthWindow)

	GetCapacity      = float64(config.GetInt("AUTH_GET_CAPACITY", defaultGetCapacity))
	GetRefillRate    = float64(GetCapacity) / AuthWindow.Seconds()
	PostCapacity     = float64(config.GetInt("AUTH_POST_CAPACITY", defaultPostCapacity))
	PostRefillRate   = float64(PostCapacity) / AuthWindow.Seconds()
	PutCapacity      = float64(config.GetInt("AUTH_PUT_CAPACITY", defaultPutCapacity))
	PutRefillRate    = float64(PutCapacity) / AuthWindow.Seconds()
	DeleteCapacity   = float64(config.GetInt("AUTH_DELETE_CAPACITY", defaultDeleteCapacity))
	DeleteRefillRate = float64(DeleteCapacity) / AuthWindow.Seconds()

	// Endpoint-specific capacities for authenticated routes.
	GetAccountsCapacity    = float64(config.GetInt("AUTH_GET_ACCOUNTS_CAPACITY", defaultGetAccountsCapacity))
	GetAccountCapacity     = float64(config.GetInt("AUTH_GET_ACCOUNT_CAPACITY", defaultGetAccountCapacity))
	GetTransactionCapacity = float64(config.GetInt("AUTH_GET_TRANSACTION_CAPACITY", defaultGetTransactionCapacity))
	GetEntriesCapacity     = float64(config.GetInt("AUTH_GET_ENTRIES_CAPACITY", defaultGetEntriesCapacity))
	PutReconcileCapacity   = float64(config.GetInt("AUTH_PUT_RECONCILE_CAPACITY", defaultPutReconcileCapacity))
	PostAccountsCapacity   = float64(config.GetInt("AUTH_POST_ACCOUNTS_CAPACITY", defaultPostAccountsCapacity))
	PostDepositCapacity    = float64(config.GetInt("AUTH_POST_DEPOSIT_CAPACITY", defaultPostDepositCapacity))
	PostWithdrawCapacity   = float64(config.GetInt("AUTH_POST_WITHDRAW_CAPACITY", defaultPostWithdrawCapacity))
	PostTransfersCapacity  = float64(config.GetInt("AUTH_POST_TRANSFERS_CAPACITY", defaultPostTransfersCapacity))
	DeleteAccountCapacity  = float64(config.GetInt("AUTH_DELETE_ACCOUNT_CAPACITY", defaultDeleteAccountCapacity))
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

var allowRequestScript = redis.NewScript(`
-- KEYS[1] is the Redis key used to track this identifier's token bucket.
-- ARGV[1] is the bucket capacity 
-- ARGV[2] is refill rate per second
-- ARGV[3] is the current UNIX timestamp
-- ARGV[4] is key expiry in seconds.

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refillRate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local expiry = tonumber(ARGV[4])

-- Read the current bucket state from Redis.
-- data[1] = tokens, data[2] = last_refill
local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = capacity
local lastRefill = now

-- If we already have stored tokens, parse them.
if data[1] and data[1] ~= false and data[1] ~= "" then
	tokens = tonumber(data[1]) or capacity
end

-- If we already have a previous refill timestamp, parse it.
if data[2] and data[2] ~= false and data[2] ~= "" then
	lastRefill = tonumber(data[2]) or now
end

-- Prevent using a future timestamp if the stored value is ahead of now.
if lastRefill > now then
	lastRefill = now
end

-- Refill tokens based on elapsed seconds.
local elapsed = now - lastRefill
tokens = tokens + elapsed * refillRate

-- Cap tokens at the maximum bucket capacity.
if tokens > capacity then
	tokens = capacity
end

-- Update the last refill time to now.
lastRefill = now

-- If no tokens remain, deny the request.
if tokens < 1 then
	return 0
end

-- Consume one token for this request.
tokens = tokens - 1

-- Persist the updated bucket state atomically.
redis.call("HSET", key, "tokens", tostring(tokens), "last_refill", tostring(lastRefill))
redis.call("EXPIRE", key, expiry)
return 1
`)

// NewRedisClient constructs a Redis client configured for local development.
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         config.GetString("REDIS_ADDR", "localhost:6379"),
		Password:     config.GetString("REDIS_PASSWORD", "Password"),
		DB:           config.GetInt("REDIS_DB", 0),
		PoolSize:     config.GetInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: config.GetInt("REDIS_MIN_IDLE_CONNS", 5),
		DialTimeout:  config.GetDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  config.GetDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: config.GetDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
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
	expiry := int(config.Window.Seconds() + 60)

	result, err := allowRequestScript.Run(ctx, rdb, []string{key}, config.Capacity, config.RefillRate, now, expiry).Result()
	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected Redis script result type: %T", result)
	}

	return allowed == 1, nil
}
