package redis

import (
	"context"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	instance *redis.Client
	once     sync.Once
	mu       sync.RWMutex
)

// GetClient returns a singleton Redis client instance
// It will initialize the client on first call and reuse it for subsequent calls
func GetClient() *redis.Client {
	once.Do(func() {
		instance = initRedisClient()
	})
	return instance
}

// ResetClient closes the current client and resets the singleton
// Useful for testing or when configuration changes
func ResetClient() {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		instance.Close()
		instance = nil
	}

	// Reset once to allow reinitialization
	once = sync.Once{}
}

// initRedisClient initializes a new Redis client from environment variables
func initRedisClient() *redis.Client {
	redisURL := os.Getenv("PB_REDIS_URL")
	if redisURL == "" {
		return nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil
	}

	return client
}

// IsAvailable checks if Redis client is available and connected
func IsAvailable() bool {
	client := GetClient()
	if client == nil {
		return false
	}

	ctx := context.Background()
	return client.Ping(ctx).Err() == nil
}

// GetClientWithFallback returns the Redis client or nil if unavailable
// This is useful when you want to handle the nil case explicitly
func GetClientWithFallback() (*redis.Client, bool) {
	client := GetClient()
	if client == nil {
		return nil, false
	}

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, false
	}

	return client, true
}
