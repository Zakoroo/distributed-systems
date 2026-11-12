// Package locks provides synchronization by resource key, without resource policy.
package locks

import "sync"

// LockMap associates each key with a read/write lock. Its zero value is usable.
// A LockMap must not be copied after first use.
type LockMap[K comparable] struct {
	mu      sync.Mutex
	entries map[K]*entry
}

type entry struct {
	mu   sync.RWMutex
	refs int // Holders and waiters both keep this entry alive.
}

func NewLockMap[K comparable]() *LockMap[K] {
	return &LockMap[K]{}
}

// WithReadLock runs handler while holding the key's shared lock.
func (m *LockMap[K]) WithReadLock(key K, handler func() error) error {
	e := m.retain(key)
	e.mu.RLock()
	defer func() {
		e.mu.RUnlock()
		m.release(key, e)
	}()
	return handler()
}

// WithWriteLock runs handler while holding the key's exclusive lock.
func (m *LockMap[K]) WithWriteLock(key K, handler func() error) error {
	e := m.retain(key)
	e.mu.Lock()
	defer func() {
		e.mu.Unlock()
		m.release(key, e)
	}()
	return handler()
}

func (m *LockMap[K]) retain(key K) *entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries == nil {
		m.entries = make(map[K]*entry)
	}
	e := m.entries[key]
	if e == nil {
		e = &entry{}
		m.entries[key] = e
	}
	e.refs++
	return e
}

func (m *LockMap[K]) release(key K, e *entry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e.refs--
	if e.refs == 0 {
		delete(m.entries, key)
	}
}
