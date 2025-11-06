package main

import "sync"

type Cache struct {
	cache map[string]interface{}
	mutex sync.RWMutex
}

var (
	instance *Cache
	once     sync.Once
)

// GetInstance returns the global singleton cache instance
func GetInstance() *Cache {
	once.Do(func() {
		instance = &Cache{
			cache: make(map[string]interface{}),
		}
	})
	return instance
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[key] = value
}

func (c *Cache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.cache[key]
}

// Global convenience functions for direct access
func Set(key string, value interface{}) {
	GetInstance().Set(key, value)
}

func Get(key string) interface{} {
	return GetInstance().Get(key)
}
