package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	customErrors "warehouse-management-system/internal/errors"

	"github.com/redis/go-redis/v9"
)

type RedisOTPRepository struct {
	client *redis.Client
	logger *slog.Logger
}

func NewRedisOTPRepository(client *redis.Client, logger *slog.Logger) *RedisOTPRepository {
	return &RedisOTPRepository{client: client, logger: logger}
}

func (r *RedisOTPRepository) key(email string) string {
	sum := sha256.Sum256([]byte(email))
	return fmt.Sprintf("otp:%s", hex.EncodeToString(sum[:]))
}

func (r *RedisOTPRepository) Save(ctx context.Context, email, code string, duration time.Duration) error {
	key := r.key(email)

	_, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, key, code, duration)
		pipe.Del(ctx, key+":attempts")
		return nil
	})
	if err != nil {
		r.logger.Error("repo: failed to save OTP to Redis", "error", err)
		return customErrors.ErrInternal
	}
	return nil
}

var verifyOTPScript = redis.NewScript(`
local value = redis.call("GET", KEYS[1])
if not value then
    return -1
end
if value == ARGV[1] then
    redis.call("DEL", KEYS[1], KEYS[2])
    return 1
end
local attempts = redis.call("INCR", KEYS[2])
if attempts == 1 then
    redis.call("PEXPIRE", KEYS[2], redis.call("PTTL", KEYS[1]))
end
if attempts >= tonumber(ARGV[2]) then
    redis.call("DEL", KEYS[1], KEYS[2])
end
return 0
`)

func (r *RedisOTPRepository) Verify(ctx context.Context, email, code string, maxAttempts int) (bool, error) {
	key := r.key(email)
	result, err := verifyOTPScript.Run(ctx, r.client, []string{key, key + ":attempts"}, code, maxAttempts).Int()
	if err != nil {
		r.logger.Error("repo: failed to verify OTP in Redis", "error", err)
		return false, customErrors.ErrInternal
	}
	if result < 0 {
		return false, customErrors.ErrNotFound
	}
	return result == 1, nil
}
