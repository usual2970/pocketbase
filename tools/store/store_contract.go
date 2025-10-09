package store

import (
	"os"

	redisutil "github.com/pocketbase/pocketbase/tools/redis"
)

type Storer[K comparable, T any] interface {
	// Length returns the current number of elements in the store.
	Length() int

	// RemoveAll removes all the existing store entries.
	RemoveAll()

	// Remove removes a single entry from the store.
	// Remove does nothing if key doesn't exist in the store.
	Remove(key K)

	// Has checks if element with the specified key exist or not.
	Has(key K) bool

	// Get returns a single element value from the store.
	// If key is not set, the zero T value is returned.
	Get(key K) T

	// GetOk is similar to Get but returns also a boolean indicating whether the key exists or not.
	GetOk(key K) (T, bool)

	// GetAll returns a shallow copy of the current store data.
	GetAll() map[K]T

	// Values returns a slice with all of the current store values.
	Values() []T

	// Set sets (or overwrite if already exists) a new value for key.
	Set(key K, value T)

	// SetFunc sets (or overwrite if already exists) a new value resolved
	// from the function callback for the provided key.
	// The function callback receives as argument the old store element value (if exists).
	// If there is no old store element, the argument will be the T zero value.
	SetFunc(key K, fn func(old T) T)

	// GetOrSet retrieves a single existing value for the provided key
	// or stores a new one if it doesn't exist.
	GetOrSet(key K, setFunc func() T) T

	// SetIfLessThanLimit sets (or overwrite if already exist) a new value for key.
	// This method is similar to Set() but **it will skip adding new elements**
	// to the store if the store length has reached the specified limit.
	// false is returned if maxAllowedElements limit is reached.
	SetIfLessThanLimit(key K, value T, maxAllowedElements int) bool

	// Reset clears the store and replaces the store data with a
	// shallow copy of the provided newData.
	Reset(newData map[K]T)
}

type Config struct {
	RedisKeyPrefix string
}
type Option func(*Config)

func WithRedisKeyPrefix(prefix string) Option {
	return func(c *Config) {
		c.RedisKeyPrefix = prefix
	}
}

// New creates a new Store[T] instance with a shallow copy of the provided data (if any).
func NewGeneral[K comparable, T any](data map[K]T, opts ...Option) Storer[K, T] {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}
	storeType := os.Getenv("PB_STORE_TYPE")
	switch storeType {
	case "redis":

		return NewRedisStore[K, T](redisutil.GetClient(), config.RedisKeyPrefix)
	default:
		return New[K, T](data)
	}
}
