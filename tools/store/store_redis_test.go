package store

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisStore(t *testing.T) {
	// Create Redis client for testing
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use test database
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping Redis store tests")
		return
	}

	// Clean up test database
	rdb.FlushDB(ctx)

	// Create store
	store := NewRedisStore[string, int](rdb, "test")

	t.Run("Length", func(t *testing.T) {
		if store.Length() != 0 {
			t.Errorf("Expected length 0, got %d", store.Length())
		}
	})

	t.Run("Set and Get", func(t *testing.T) {
		store.Set("key1", 42)
		if value := store.Get("key1"); value != 42 {
			t.Errorf("Expected 42, got %d", value)
		}
	})

	t.Run("Has", func(t *testing.T) {
		if !store.Has("key1") {
			t.Error("Expected key1 to exist")
		}
		if store.Has("nonexistent") {
			t.Error("Expected nonexistent key to not exist")
		}
	})

	t.Run("GetOk", func(t *testing.T) {
		value, ok := store.GetOk("key1")
		if !ok || value != 42 {
			t.Errorf("Expected (42, true), got (%d, %v)", value, ok)
		}

		value, ok = store.GetOk("nonexistent")
		if ok {
			t.Errorf("Expected (0, false), got (%d, %v)", value, ok)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		store.Remove("key1")
		if store.Has("key1") {
			t.Error("Expected key1 to be removed")
		}
	})

	t.Run("SetFunc", func(t *testing.T) {
		store.SetFunc("counter", func(old int) int {
			return old + 1
		})
		if value := store.Get("counter"); value != 1 {
			t.Errorf("Expected 1, got %d", value)
		}

		store.SetFunc("counter", func(old int) int {
			return old + 1
		})
		if value := store.Get("counter"); value != 2 {
			t.Errorf("Expected 2, got %d", value)
		}
	})

	t.Run("GetOrSet", func(t *testing.T) {
		// Test existing key
		value := store.GetOrSet("counter", func() int { return 999 })
		if value != 2 {
			t.Errorf("Expected 2, got %d", value)
		}

		// Test new key
		value = store.GetOrSet("newkey", func() int { return 123 })
		if value != 123 {
			t.Errorf("Expected 123, got %d", value)
		}
	})

	t.Run("SetIfLessThanLimit", func(t *testing.T) {
		// Clear store first
		store.RemoveAll()

		// Test within limit
		success := store.SetIfLessThanLimit("limit1", 1, 10)
		if !success {
			t.Error("Expected SetIfLessThanLimit to succeed")
		}

		// Test at limit
		success = store.SetIfLessThanLimit("limit2", 2, 2)
		if !success {
			t.Error("Expected SetIfLessThanLimit to succeed")
		}

		// Test over limit
		success = store.SetIfLessThanLimit("limit3", 3, 2)
		if success {
			t.Error("Expected SetIfLessThanLimit to fail")
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		all := store.GetAll()
		expectedKeys := []string{"limit1", "limit2"}
		if len(all) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(all))
		}

		for _, key := range expectedKeys {
			if _, exists := all[key]; !exists {
				t.Errorf("Expected key %s to exist in GetAll", key)
			}
		}
	})

	t.Run("Values", func(t *testing.T) {
		values := store.Values()
		if len(values) != 2 {
			t.Errorf("Expected 2 values, got %d", len(values))
		}
	})

	t.Run("Reset", func(t *testing.T) {
		newData := map[string]int{
			"reset1": 100,
			"reset2": 200,
		}
		store.Reset(newData)

		all := store.GetAll()
		if len(all) != 2 {
			t.Errorf("Expected 2 keys after reset, got %d", len(all))
		}

		if value := store.Get("reset1"); value != 100 {
			t.Errorf("Expected 100, got %d", value)
		}
	})

	t.Run("RemoveAll", func(t *testing.T) {
		store.RemoveAll()
		if store.Length() != 0 {
			t.Errorf("Expected length 0 after RemoveAll, got %d", store.Length())
		}
	})

	t.Run("SetWithExpiration", func(t *testing.T) {
		store.SetWithExpiration("expire", 999, 100*time.Millisecond)
		if value := store.Get("expire"); value != 999 {
			t.Errorf("Expected 999, got %d", value)
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)
		if store.Has("expire") {
			t.Error("Expected key to be expired")
		}
	})

	t.Run("GetWithTTL", func(t *testing.T) {
		store.SetWithExpiration("ttl", 888, 1*time.Second)
		value, ttl, ok := store.GetWithTTL("ttl")
		if !ok {
			t.Error("Expected key to exist")
		}
		if value != 888 {
			t.Errorf("Expected 888, got %d", value)
		}
		if ttl <= 0 {
			t.Error("Expected positive TTL")
		}
	})

	// Clean up
	rdb.FlushDB(ctx)
}
