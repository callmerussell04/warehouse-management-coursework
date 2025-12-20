package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	customErrors "warehouse-management-system/internal/errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisRefreshSessionRepository struct {
	client *redis.Client
}

func NewRedisRefreshSessionRepository(client *redis.Client) *RedisRefreshSessionRepository {
	return &RedisRefreshSessionRepository{client: client}
}

func refreshTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func refreshTokenKey(hash string) string {
	return "refresh:" + hash
}

func refreshUserKey(userID uuid.UUID) string {
	return "refresh-user:" + userID.String()
}

func (r *RedisRefreshSessionRepository) Save(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	hash := refreshTokenHash(token)
	_, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, refreshTokenKey(hash), userID.String(), ttl)
		pipe.SAdd(ctx, refreshUserKey(userID), hash)
		pipe.Expire(ctx, refreshUserKey(userID), ttl)
		return nil
	})
	if err != nil {
		return customErrors.ErrInternal
	}
	return nil
}

var rotateRefreshScript = redis.NewScript(`
local user_id = redis.call("GET", KEYS[1])
if not user_id then
    return ""
end
redis.call("DEL", KEYS[1])
redis.call("SREM", "refresh-user:" .. user_id, ARGV[1])
redis.call("SET", KEYS[2], user_id, "PX", ARGV[3])
redis.call("SADD", "refresh-user:" .. user_id, ARGV[2])
redis.call("PEXPIRE", "refresh-user:" .. user_id, ARGV[3])
return user_id
`)

func (r *RedisRefreshSessionRepository) Rotate(ctx context.Context, oldToken, newToken string, ttl time.Duration) (uuid.UUID, error) {
	oldHash := refreshTokenHash(oldToken)
	newHash := refreshTokenHash(newToken)
	value, err := rotateRefreshScript.Run(ctx, r.client,
		[]string{refreshTokenKey(oldHash), refreshTokenKey(newHash)},
		oldHash, newHash, ttl.Milliseconds()).Text()
	if errors.Is(err, redis.Nil) || value == "" {
		return uuid.Nil, customErrors.ErrUnauthorized
	}
	if err != nil {
		return uuid.Nil, customErrors.ErrInternal
	}
	userID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, customErrors.ErrInternal
	}
	return userID, nil
}

func (r *RedisRefreshSessionRepository) Revoke(ctx context.Context, token string) error {
	hash := refreshTokenHash(token)
	value, err := r.client.GetDel(ctx, refreshTokenKey(hash)).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return customErrors.ErrInternal
	}
	userID, err := uuid.Parse(value)
	if err != nil {
		return customErrors.ErrInternal
	}
	if err := r.client.SRem(ctx, refreshUserKey(userID), hash).Err(); err != nil {
		return customErrors.ErrInternal
	}
	return nil
}

var revokeAllRefreshScript = redis.NewScript(`
local hashes = redis.call("SMEMBERS", KEYS[1])
for _, hash in ipairs(hashes) do
    redis.call("DEL", ARGV[1] .. hash)
end
redis.call("DEL", KEYS[1])
return #hashes
`)

func (r *RedisRefreshSessionRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	_, err := revokeAllRefreshScript.Run(ctx, r.client, []string{refreshUserKey(userID)}, "refresh:").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("%w: failed to revoke refresh sessions", customErrors.ErrInternal)
	}
	return nil
}
