package main

import "sync"

type BlockingBlpopWaiter struct {
	ResponseChannel chan string
	listKey         string
}

type BlockingBlpopRegistry struct {
	mutex   sync.Mutex
	waiters map[string][]*BlockingBlpopWaiter
}

func NewBlockingBlpopRegistry() *BlockingBlpopRegistry {
	return &BlockingBlpopRegistry{
		waiters: make(map[string][]*BlockingBlpopWaiter),
	}
}

var blockingBlpopRegistry = NewBlockingBlpopRegistry()

func GetBlockingBlpopRegistry() *BlockingBlpopRegistry {
	return blockingBlpopRegistry
}

func SetBlockingBlpopRegistryForTest(registry *BlockingBlpopRegistry) {
	blockingBlpopRegistry = registry
}

func (registry *BlockingBlpopRegistry) RegisterWaiter(listKey string) *BlockingBlpopWaiter {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	waiter := &BlockingBlpopWaiter{
		ResponseChannel: make(chan string, 1),
		listKey:         listKey,
	}

	registry.waiters[listKey] = append(registry.waiters[listKey], waiter)

	return waiter
}

func (registry *BlockingBlpopRegistry) UnregisterWaiter(listKey string, targetWaiter *BlockingBlpopWaiter) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	waiters := registry.waiters[listKey]
	for index, waiter := range waiters {
		if waiter == targetWaiter {
			registry.waiters[listKey] = append(waiters[:index], waiters[index+1:]...)
			return
		}
	}
}

func (registry *BlockingBlpopRegistry) HasWaiters(listKey string) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	return len(registry.waiters[listKey]) > 0
}

func (registry *BlockingBlpopRegistry) NotifyNextWaiter(listKey string, response string) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	waiters := registry.waiters[listKey]
	if len(waiters) == 0 {
		return false
	}

	nextWaiter := waiters[0]
	registry.waiters[listKey] = waiters[1:]

	nextWaiter.ResponseChannel <- response

	return true
}

func (waiter *BlockingBlpopWaiter) Unregister() {
	GetBlockingBlpopRegistry().UnregisterWaiter(waiter.listKey, waiter)
}
