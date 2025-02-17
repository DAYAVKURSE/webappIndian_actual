package redis

import (
	"BlessedApi/pkg/logger"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisService represents the Redis service
type RedisService struct {
	client *redis.Client // Keep the field unexported
}

// Client returns the Redis client
func (r *RedisService) Client() *redis.Client {
	return r.client
}

// NewRedisService creates a new instance of the Redis service
func NewRedisService(redisAddr string, redisPassword string) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("%v", err)
	}

	logger.Info("Connected to Redis")

	return &RedisService{
		client: client,
	}
}

// SetKey sets a key-value pair in Redis
func (r *RedisService) SetKey(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return logger.WrapError(err, "")
	}
	return nil
}

// GetKey retrieves the value of a key from Redis
func (r *RedisService) GetKey(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", logger.WrapError(err, "")
	}
	return val, nil
}

// DeleteKey removes a key from Redis
func (r *RedisService) DeleteKey(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return logger.WrapError(err, "")
	}
	return nil
}
