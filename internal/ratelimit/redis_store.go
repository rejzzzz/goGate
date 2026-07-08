package ratelimit

// redis_store.go - Redis-backed rate limit storage
//
// Responsibilities:
// - Store token bucket state in Redis for distributed rate limiting
// - Execute Lua script atomically for token bucket operations
// - Use EVALSHA for performance (preload script at startup)
// - Fallback to EVAL if script not cached
// - Set TTL on rate limit keys to prevent unbounded memory growth
// - Build Redis key: ratelimit:{route}:{client_ip}
//
// Key Functions:
// - NewRedisStore(client *redis.Client) *RedisStore: Create Redis store
// - LoadScript() error: Preload Lua script and cache SHA digest
// - CheckRateLimit(route, clientIP string, rate float64, burst int) (allowed bool, remaining int, error): Execute rate limit check
//
// Lua Script:
// - KEYS[1]: ratelimit:{route}:{client_ip}
// - ARGV[1]: rate (tokens per second)
// - ARGV[2]: burst (max bucket size)
// - ARGV[3]: current timestamp (ms)
// - Returns: 1 (allowed) or 0 (denied)
//
// Inputs:
// - Route name and client IP
// - Rate limit configuration
// - Current timestamp
//
// Outputs:
// - Allowed/denied decision
// - Remaining tokens
// - Error if Redis operation fails

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const luaScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1]) or burst
local ts = tonumber(data[2]) or now

local elapsed = (now - ts) / 1000.0
tokens = math.min(burst, tokens + elapsed * rate)

if tokens >= 1 then
    tokens = tokens - 1
    redis.call("HMSET", key, "tokens", tokens, "ts", now)
    redis.call("EXPIRE", key, 60)
    return {1, math.floor(tokens)}
else
    redis.call("HMSET", key, "tokens", tokens, "ts", now)
    redis.call("EXPIRE", key, 60)
    return {0, math.floor(tokens)}
end
`

type RedisStore struct {
	client    *redis.Client
	scriptSHA string // Cached Lua script SHA digest
}

// NewRedisStore creates a new Redis-backed rate limit store
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

// LoadScript preloads the Lua script into Redis and caches the SHA
func (rs *RedisStore) LoadScript(ctx context.Context) error {
	sha, err := rs.client.ScriptLoad(ctx, luaScript).Result()
	if err != nil {
		return err
	}
	rs.scriptSHA = sha
	return nil
}

// CheckRateLimit checks if a request is allowed under rate limit
func (rs *RedisStore) CheckRateLimit(route, clientIP string, rate float64, burst int) (bool, int, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", route, clientIP)
	now := time.Now().UnixMilli()

	// Use EVALSHA if we have the SHA cached
	var res interface{}
	var err error

	if rs.scriptSHA != "" {
		res, err = rs.client.EvalSha(context.Background(), rs.scriptSHA, []string{key}, rate, burst, now).Result()
	}

	// Fallback to EVAL if script is not cached or was evicted
	if err != nil || rs.scriptSHA == "" {
		res, err = rs.client.Eval(context.Background(), luaScript, []string{key}, rate, burst, now).Result()
		if err != nil {
			return false, 0, err
		}
	}

	// Parse Lua script result: {allowed (1/0), remaining_tokens}
	vals, ok := res.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, fmt.Errorf("unexpected redis script result")
	}

	allowedInt, ok1 := vals[0].(int64)
	remainingInt, ok2 := vals[1].(int64)

	if !ok1 || !ok2 {
		return false, 0, fmt.Errorf("unexpected redis script result types")
	}

	return allowedInt == 1, int(remainingInt), nil
}
