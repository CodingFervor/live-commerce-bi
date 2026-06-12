package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	""+MOD+"/internal/config"
)

var client *redis.Client

func Init(cfg *config.RedisConfig) error {
	client = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	log.Println("[Redis] Connected")
	return nil
}

func Get() *redis.Client {
	return client
}

func Close() error {
	if client != nil {
		if err := client.Close(); err != nil {
			return err
		}
		log.Println("[Redis] Connection closed")
	}
	return nil
}

// Helper functions for common cache operations

func SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return client.Set(ctx, key, value, ttl).Err()
}

func GetJSON(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func IncrBy(ctx context.Context, key string, val int64) (int64, error) {
	return client.IncrBy(ctx, key, val).Result()
}

func HSet(ctx context.Context, key string, values ...interface{}) error {
	return client.HSet(ctx, key, values...).Err()
}

func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return client.HGetAll(ctx, key).Result()
}

func ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return client.ZAdd(ctx, key, members...).Err()
}

func ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return client.ZRangeByScore(ctx, key, opt).Result()
}

func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, key string) (bool, error) {
	n, err := client.Exists(ctx, key).Result()
	return n > 0, err
}

func Expire(ctx context.Context, key string, ttl time.Duration) error {
	return client.Expire(ctx, key, ttl).Err()
}

func TTL(ctx context.Context, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}

// Publish for real-time updates via Pub/Sub
func Publish(ctx context.Context, channel string, message interface{}) error {
	return client.Publish(ctx, channel, message).Err()
}

// Subscribe for real-time updates
func Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return client.Subscribe(ctx, channels...)
}
