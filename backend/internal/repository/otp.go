package repository

import (
	"context"
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
	return fmt.Sprintf("otp:%s", email)
}

func (r *RedisOTPRepository) Save(ctx context.Context, email, code string, duration time.Duration) error {
	key := r.key(email)

	err := r.client.Set(ctx, key, code, duration).Err()
	if err != nil {
		r.logger.Error("repo: failed to save OTP to Redis", "error", err, "key", key)
		return customErrors.ErrInternal
	}
	return nil
}

func (r *RedisOTPRepository) Get(ctx context.Context, email string) (string, error) {
	key := r.key(email)

	val, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		r.logger.Info("repo: OTP not found in Redis (expired or incorrect)", "key", key)
		return "", customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repo: failed to retrieve OTP from Redis", "error", err, "key", key)
		return "", customErrors.ErrInternal
	}

	return val, nil
}

func (r *RedisOTPRepository) Delete(ctx context.Context, email string) error {
	key := r.key(email)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("repo: failed to delete OTP from Redis", "error", err, "key", key)
		return customErrors.ErrInternal
	}

	return nil
}
