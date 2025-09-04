package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type RateLimitStateManager interface {
	Get(key string) (int64, error)
	Set(key string, value int64, expireTime time.Duration) (string, error)
	Decr(key string) (int64, error)
	Incr(key string) (int64, error)
}

type entry struct {
	value  int64
	expiry time.Time
}

type InMemoryStateManager struct {
	store map[string]entry
	mu    sync.RWMutex
}

func NewInMemoryStateManager() *InMemoryStateManager {
	sm := &InMemoryStateManager{
		store: make(map[string]entry),
	}

	go sm.cleanup()

	return sm
}

func (sm *InMemoryStateManager) Get(key string) (int64, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ent, exists := sm.store[key]
	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	if !ent.expiry.IsZero() && time.Now().After(ent.expiry) {
		return 0, fmt.Errorf("key expired")
	}

	return ent.value, nil
}

func (sm *InMemoryStateManager) Set(key string, value int64, expireTime time.Duration) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var expiry time.Time
	if expireTime > 0 {
		expiry = time.Now().Add(expireTime)
	}

	sm.store[key] = entry{
		value:  value,
		expiry: expiry,
	}

	return key, nil
}

func (sm *InMemoryStateManager) Decr(key string) (int64, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ent, exists := sm.store[key]
	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	if !ent.expiry.IsZero() && time.Now().After(ent.expiry) {
		delete(sm.store, key)
		return 0, fmt.Errorf("key expired")
	}

	newValue := ent.value - 1
	sm.store[key] = entry{value: newValue, expiry: ent.expiry}
	return newValue, nil
}

func (sm *InMemoryStateManager) Incr(key string) (int64, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ent, exists := sm.store[key]
	if !exists {
		sm.store[key] = entry{value: 1, expiry: time.Time{}}
		return 1, nil
	}

	if !ent.expiry.IsZero() && time.Now().After(ent.expiry) {
		delete(sm.store, key)
		sm.store[key] = entry{value: 1, expiry: time.Time{}}
		return 1, nil
	}

	newValue := ent.value + 1
	sm.store[key] = entry{value: newValue, expiry: ent.expiry}
	return newValue, nil
}

func (sm *InMemoryStateManager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for key, entry := range sm.store {
			if !entry.expiry.IsZero() && now.After(entry.expiry) {
				delete(sm.store, key)
			}
		}
		sm.mu.Unlock()
	}
}
