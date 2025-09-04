package otp

import (
	"fmt"
	"sync"
	"time"
)

type OTPStateManager interface {
	SetX(key string, val string) error
	Get(key string) (string, error)
}

type otpEntry struct {
	value  string
	expiry time.Time
}

type MemStateManager struct {
	TTL   time.Duration
	store map[string]otpEntry
	mu    sync.RWMutex
}

func NewMemStateManager(ttl time.Duration) *MemStateManager {
	sm := &MemStateManager{
		TTL:   ttl,
		store: make(map[string]otpEntry),
	}

	go sm.cleanup()

	return sm
}

func (ms *MemStateManager) SetX(key string, val string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.store[key] = otpEntry{
		value:  val,
		expiry: time.Now().Add(ms.TTL),
	}
	return nil
}

func (ms *MemStateManager) Get(key string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	entry, exists := ms.store[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	if time.Now().After(entry.expiry) {
		delete(ms.store, key)
		return "", fmt.Errorf("key expired")
	}

	return entry.value, nil
}

func (ms *MemStateManager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ms.mu.Lock()
		now := time.Now()
		for key, entry := range ms.store {
			if !entry.expiry.IsZero() && now.After(entry.expiry) {
				delete(ms.store, key)
			}
		}
		ms.mu.Unlock()
	}
}
