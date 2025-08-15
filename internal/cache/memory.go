// Package cache provides a simple in-memory cache.
package cache

import (
	"sync"
	"time"
)

// Memory is a TTL-based in-memory cache safe for concurrent use.
type Memory struct {
	mu    sync.Mutex
	items map[string]entry
	ttl   time.Duration
}

type entry struct {
	v   any
	exp time.Time
}

// NewMemory creates a new Memory cache with the provided TTL for entries.
func NewMemory(ttl time.Duration) *Memory {
	return &Memory{
		items: make(map[string]entry),
		ttl:   ttl,
	}
}

// Get retrieves a value by key if present and not expired.
func (m *Memory) Get(key string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.items[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.exp) {
		delete(m.items, key)
		return nil, false
	}
	return e.v, true
}

// Set stores a value by key with the configured TTL.
func (m *Memory) Set(key string, v any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = entry{v: v, exp: time.Now().Add(m.ttl)}
}

// Delete removes a key from the cache.
func (m *Memory) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}
