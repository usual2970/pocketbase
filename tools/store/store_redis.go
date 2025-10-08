package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore[K comparable, T any] struct {
	redis     *redis.Client
	ctx       context.Context
	keyPrefix string
}

func NewRedisStore[K comparable, T any](redis *redis.Client, keyPrefix string) *RedisStore[K, T] {
	return &RedisStore[K, T]{
		redis:     redis,
		ctx:       context.Background(),
		keyPrefix: keyPrefix,
	}
}

// getRedisKey returns the full Redis key with prefix
func (s *RedisStore[K, T]) getRedisKey(key K) string {
	return fmt.Sprintf("%s:%v", s.keyPrefix, key)
}

// Length returns the current number of elements in the store.
func (s *RedisStore[K, T]) Length() int {
	keys, err := s.redis.Keys(s.ctx, s.keyPrefix+":*").Result()
	if err != nil {
		return 0
	}
	return len(keys)
}

// RemoveAll removes all the existing store entries.
func (s *RedisStore[K, T]) RemoveAll() {
	keys, err := s.redis.Keys(s.ctx, s.keyPrefix+":*").Result()
	if err != nil {
		return
	}
	if len(keys) > 0 {
		s.redis.Del(s.ctx, keys...)
	}
}

// Remove removes a single entry from the store.
// Remove does nothing if key doesn't exist in the store.
func (s *RedisStore[K, T]) Remove(key K) {
	s.redis.Del(s.ctx, s.getRedisKey(key))
}

// Has checks if element with the specified key exist or not.
func (s *RedisStore[K, T]) Has(key K) bool {
	exists, _ := s.redis.Exists(s.ctx, s.getRedisKey(key)).Result()
	return exists > 0
}

// Get returns a single element value from the store.
// If key is not set, the zero T value is returned.
func (s *RedisStore[K, T]) Get(key K) T {
	val, err := s.redis.Get(s.ctx, s.getRedisKey(key)).Result()
	if err != nil {
		var zero T
		return zero
	}

	var result T
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		var zero T
		return zero
	}
	return result
}

// GetOk is similar to Get but returns also a boolean indicating whether the key exists or not.
func (s *RedisStore[K, T]) GetOk(key K) (T, bool) {
	val, err := s.redis.Get(s.ctx, s.getRedisKey(key)).Result()
	if err != nil {
		var zero T
		return zero, false
	}

	var result T
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		var zero T
		return zero, false
	}
	return result, true
}

// GetAll returns a shallow copy of the current store data.
func (s *RedisStore[K, T]) GetAll() map[K]T {
	keys, err := s.redis.Keys(s.ctx, s.keyPrefix+":*").Result()
	if err != nil {
		return make(map[K]T)
	}

	result := make(map[K]T, len(keys))
	for _, redisKey := range keys {
		// Extract the original key from Redis key
		originalKey := redisKey[len(s.keyPrefix)+1:] // Remove prefix and colon

		val, err := s.redis.Get(s.ctx, redisKey).Result()
		if err != nil {
			continue
		}

		var value T
		if err := json.Unmarshal([]byte(val), &value); err != nil {
			continue
		}

		// Convert string back to original key type
		// This is a simplified approach - in practice you might need more sophisticated key conversion
		var key K
		switch any(key).(type) {
		case string:
			key = any(originalKey).(K)
		case int:
			if intVal, err := strconv.Atoi(originalKey); err == nil {
				key = any(intVal).(K)
			}
		default:
			// For other types, you might need custom conversion logic
			continue
		}

		result[key] = value
	}

	return result
}

// Values returns a slice with all of the current store values.
func (s *RedisStore[K, T]) Values() []T {
	keys, err := s.redis.Keys(s.ctx, s.keyPrefix+":*").Result()
	if err != nil {
		return []T{}
	}

	values := make([]T, 0, len(keys))
	for _, redisKey := range keys {
		val, err := s.redis.Get(s.ctx, redisKey).Result()
		if err != nil {
			continue
		}

		var value T
		if err := json.Unmarshal([]byte(val), &value); err != nil {
			continue
		}

		values = append(values, value)
	}

	return values
}

// Set sets (or overwrite if already exists) a new value for key.
func (s *RedisStore[K, T]) Set(key K, value T) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	s.redis.Set(s.ctx, s.getRedisKey(key), data, 0) // 0 means no expiration
}

// SetFunc sets (or overwrite if already exists) a new value resolved
// from the function callback for the provided key.
// The function callback receives as argument the old store element value (if exists).
// If there is no old store element, the argument will be the T zero value.
func (s *RedisStore[K, T]) SetFunc(key K, fn func(old T) T) {
	oldValue := s.Get(key)
	newValue := fn(oldValue)
	s.Set(key, newValue)
}

// GetOrSet retrieves a single existing value for the provided key
// or stores a new one if it doesn't exist.
func (s *RedisStore[K, T]) GetOrSet(key K, setFunc func() T) T {
	// Try to get existing value
	if value, ok := s.GetOk(key); ok {
		return value
	}

	// Set new value
	newValue := setFunc()
	s.Set(key, newValue)
	return newValue
}

// SetIfLessThanLimit sets (or overwrite if already exist) a new value for key.
// This method is similar to Set() but **it will skip adding new elements**
// to the store if the store length has reached the specified limit.
// false is returned if maxAllowedElements limit is reached.
func (s *RedisStore[K, T]) SetIfLessThanLimit(key K, value T, maxAllowedElements int) bool {
	// Check if key already exists
	if s.Has(key) {
		s.Set(key, value)
		return true
	}

	// Check current length
	currentLength := s.Length()
	if currentLength >= maxAllowedElements {
		return false
	}

	s.Set(key, value)
	return true
}

// Reset clears the store and replaces the store data with a
// shallow copy of the provided newData.
func (s *RedisStore[K, T]) Reset(newData map[K]T) {
	// Clear existing data
	s.RemoveAll()

	// Set new data
	for key, value := range newData {
		s.Set(key, value)
	}
}

// SetWithExpiration sets a key-value pair with expiration
func (s *RedisStore[K, T]) SetWithExpiration(key K, value T, expiration time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	s.redis.Set(s.ctx, s.getRedisKey(key), data, expiration)
}

// GetWithTTL returns the value and remaining TTL for a key
func (s *RedisStore[K, T]) GetWithTTL(key K) (T, time.Duration, bool) {
	val, err := s.redis.Get(s.ctx, s.getRedisKey(key)).Result()
	if err != nil {
		var zero T
		return zero, 0, false
	}

	ttl, err := s.redis.TTL(s.ctx, s.getRedisKey(key)).Result()
	if err != nil {
		var zero T
		return zero, 0, false
	}

	var result T
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		var zero T
		return zero, 0, false
	}

	return result, ttl, true
}
