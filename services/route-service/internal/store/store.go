package store

import "time"

// Store is the persistence interface for route data.
type Store interface {
	Health() string
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	createdAt time.Time
}

// NewMemoryStore creates an in-memory store.
func NewMemoryStore(now time.Time) Store {
	return &MemoryStore{createdAt: now}
}

// Health returns a simple health indicator.
func (m *MemoryStore) Health() string {
	return "ok"
}
