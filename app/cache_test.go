package main

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	cache := &Cache{cache: make(map[string]CacheItem)}

	// Test basic set/get without expiration
	cache.Set("key1", "value1", map[string]interface{}{})
	if v := cache.Get("key1"); v != "value1" {
		t.Errorf("expected value1, got %v", v)
	}
}

func TestCacheSetGetWithEX(t *testing.T) {
	cache := &Cache{cache: make(map[string]CacheItem)}

	// Set with EX=2 (2 seconds)
	cache.Set("key2", "value2", map[string]interface{}{"EX": 2})

	// Immediately get, should return value
	if v := cache.Get("key2"); v != "value2" {
		t.Errorf("expected value2, got %v", v)
	}

	// Wait 3 seconds
	time.Sleep(3 * time.Second)

	// Get again, should return nil (expired and deleted)
	if v := cache.Get("key2"); v != nil {
		t.Errorf("expected nil (expired), got %v", v)
	}
}

func TestCacheSetGetWithPX(t *testing.T) {
	cache := &Cache{cache: make(map[string]CacheItem)}

	// Set with PX=1000 (1000 milliseconds = 1 second)
	cache.Set("key3", "value3", map[string]interface{}{"PX": 1000})

	// Immediately get, should return value
	if v := cache.Get("key3"); v != "value3" {
		t.Errorf("expected value3, got %v", v)
	}

	// Wait 2 seconds
	time.Sleep(2 * time.Second)

	// Get again, should return nil (expired and deleted)
	if v := cache.Get("key3"); v != nil {
		t.Errorf("expected nil (expired), got %v", v)
	}
}

func TestCacheSetGetPXOverridesEX(t *testing.T) {
	cache := &Cache{cache: make(map[string]CacheItem)}

	// Set with both EX and PX, PX should take precedence (Redis behavior)
	cache.Set("key4", "value4", map[string]interface{}{"EX": 10, "PX": 1000})

	// Wait 2 seconds (PX=1000ms = 1s, so expired)
	time.Sleep(2 * time.Second)

	// Should be expired
	if v := cache.Get("key4"); v != nil {
		t.Errorf("expected nil (expired), got %v", v)
	}
}

func TestHandleSetWithPXIntegration(t *testing.T) {
	// Clear cache
	GetInstance().cache = make(map[string]CacheItem)

	// Create SET command with PX option: SET mykey myvalue PX 1000
	cmd := &RedisCommand{
		Type: CmdSET,
		Args: []string{"mykey", "myvalue", "PX", "1000"},
	}

	// Call HandleSet which goes through the full flow:
	// HandleSet -> handleOptionalArguments -> cache.Set
	result := HandleSet(cmd)

	// Should return OK
	if result != "+OK\r\n" {
		t.Errorf("expected +OK\\r\\n, got %s", result)
	}

	// Verify the key was set with correct expiration
	cache := GetInstance()
	if item, exists := cache.cache["mykey"]; !exists {
		t.Fatal("key should exist in cache")
	} else {
		// PX 1000 means expire in 1000ms = 1 second
		expectedExp := time.Now().Unix() + 1 // 1000ms = 1 second
		if item.Expiration < expectedExp-1 || item.Expiration > expectedExp+1 {
			t.Errorf("expected expiration around %d, got %d", expectedExp, item.Expiration)
		}
		if item.Value != "myvalue" {
			t.Errorf("expected value 'myvalue', got %v", item.Value)
		}
	}

	// Wait for expiration (PX 1000 = 1 second, wait 1.5 seconds to be safe)
	time.Sleep(1500 * time.Millisecond)

	// Key should be gone after expiration (Get should return nil and delete it)
	if v := cache.Get("mykey"); v != nil {
		t.Errorf("key should have expired, got %v", v)
	}
}
