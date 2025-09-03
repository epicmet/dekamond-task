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
	return &MemStateManager{
		TTL:   ttl,
		store: make(map[string]otpEntry),
	}
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

	// TODO: We are keeping the expired key in mem. It's technically a memory leak.
	// Maybe instead, check on an interval for expired keys? For now, it good enough
	if time.Now().After(entry.expiry) {
		delete(ms.store, key)
		return "", fmt.Errorf("key expired")
	}

	return entry.value, nil
}
