package main

import (
	"sync"
	"testing"
)

func TestGetInstance(t *testing.T) {
	cache := GetInstance()

	if cache == nil {
		t.Fatal("GetInstance() returned nil")
	}

	// Test that multiple calls return the same instance
	cache2 := GetInstance()
	if cache != cache2 {
		t.Fatal("GetInstance() returned different instances")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := GetInstance()

	// Test setting and getting a string value
	cache.Set("name", "Alice")
	value := cache.Get("name")

	if value != "Alice" {
		t.Errorf("Expected 'Alice', got %v", value)
	}

	// Test setting and getting different types
	cache.Set("age", 25)
	age := cache.Get("age")

	if age != 25 {
		t.Errorf("Expected 25, got %v", age)
	}

	cache.Set("active", true)
	active := cache.Get("active")

	if active != true {
		t.Errorf("Expected true, got %v", active)
	}
}

func TestCacheGetNonExistentKey(t *testing.T) {
	cache := GetInstance()

	// Clear any existing data for this test
	cache.cache = make(map[string]interface{})

	value := cache.Get("nonexistent")

	if value != nil {
		t.Errorf("Expected nil for non-existent key, got %v", value)
	}
}

func TestCacheOverwriteValue(t *testing.T) {
	cache := GetInstance()

	// Clear any existing data for this test
	cache.cache = make(map[string]interface{})

	cache.Set("key", "first")
	firstValue := cache.Get("key")

	if firstValue != "first" {
		t.Errorf("Expected 'first', got %v", firstValue)
	}

	cache.Set("key", "second")
	secondValue := cache.Get("key")

	if secondValue != "second" {
		t.Errorf("Expected 'second', got %v", secondValue)
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := GetInstance()

	// Clear any existing data for this test
	cache.cache = make(map[string]interface{})

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines performing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + id))
				cache.Set(key, j)
				_ = cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	for i := 0; i < numGoroutines; i++ {
		key := string(rune('a' + i))
		value := cache.Get(key)
		if value != numOperations-1 {
			t.Errorf("Expected final value %d for key %s, got %v", numOperations-1, key, value)
		}
	}
}

func TestCacheMultipleKeys(t *testing.T) {
	cache := GetInstance()

	// Clear any existing data for this test
	cache.cache = make(map[string]interface{})

	testData := map[string]interface{}{
		"user1": "Alice",
		"user2": "Bob",
		"user3": "Charlie",
		"count": 42,
		"rate":  3.14,
	}

	// Set all values
	for key, value := range testData {
		cache.Set(key, value)
	}

	// Verify all values
	for key, expectedValue := range testData {
		actualValue := cache.Get(key)
		if actualValue != expectedValue {
			t.Errorf("Expected %v for key %s, got %v", expectedValue, key, actualValue)
		}
	}
}

func TestCacheEmptyKey(t *testing.T) {
	cache := GetInstance()

	// Clear any existing data for this test
	cache.cache = make(map[string]interface{})

	cache.Set("", "empty")
	value := cache.Get("")

	if value != "empty" {
		t.Errorf("Expected 'empty' for empty key, got %v", value)
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Clear any existing data for this test
	GetInstance().cache = make(map[string]interface{})

	// Test global Set function
	Set("global_key", "global_value")

	// Test global Get function
	value := Get("global_key")

	if value != "global_value" {
		t.Errorf("Expected 'global_value', got %v", value)
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache := GetInstance()
	// Clear cache for benchmark
	cache.cache = make(map[string]interface{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := GetInstance()
	// Clear cache and set test value
	cache.cache = make(map[string]interface{})
	cache.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Get("key")
	}
}

func BenchmarkCacheConcurrent(b *testing.B) {
	cache := GetInstance()
	// Clear cache for benchmark
	cache.cache = make(map[string]interface{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set("key", "value")
			_ = cache.Get("key")
		}
	})
}
