package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	customErrors "warehouse-management-system/internal/errors"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

var rateLimitScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return {count, redis.call("PTTL", KEYS[1])}
`)

func (r *RedisRateLimiter) Allow(ctx context.Context, identity string, limit int64, window time.Duration) (bool, time.Duration, error) {
	sum := sha256.Sum256([]byte(identity))
	key := "rate-limit:" + hex.EncodeToString(sum[:])

	result, err := rateLimitScript.Run(ctx, r.client, []string{key}, window.Milliseconds()).Int64Slice()
	if err != nil || len(result) != 2 {
		return false, 0, customErrors.ErrInternal
	}
	retryAfter := time.Duration(result[1]) * time.Millisecond
	if retryAfter < 0 {
		retryAfter = window
	}
	return result[0] <= limit, retryAfter, nil
}
