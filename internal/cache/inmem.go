package cache

import (
	"context"
	"encoding/gob"
	"errors"
	"io"
	"sync"
	"time"
)

// ErrNotFound returns when key not found in cache
var ErrNotFound = errors.New("key not found")

// InMem in-memory cache
type InMem struct {
	mu   *sync.Mutex
	data *data
}

type data struct {
	Index *btree
	KV    map[string]string
}

// New returns new in-memory cache
func New() *InMem {
	return &InMem{
		mu: new(sync.Mutex),
		data: &data{
			Index: new(btree),
			KV:    make(map[string]string),
		},
	}
}

// Get returns value by key
func (m *InMem) Get(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v, ok := m.data.KV[key]; ok {
		return v, nil
	}

	return "", ErrNotFound
}

// Set sets the value and TTL for the specified key
func (m *InMem) Set(key, value string, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ts := time.Now().Add(ttl).Unix()
	m.data.KV[key] = value
	m.data.Index.insert(ts, key)
}

// Delete deletes value by key
func (m *InMem) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.delete(key)
}

// Eviction starts the process of removing keys whose life has come to an end
func (m *InMem) Eviction(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-ticker.C:
			m.mu.Lock()
			{
				for k := range m.data.Index.removeNodesLessThan(ts.Unix()) {
					m.delete(k)
				}
			}
			m.mu.Unlock()
		}
	}
}

// Dump writes cached data into a specified writer
func (m *InMem) Dump(w io.Writer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return gob.NewEncoder(w).Encode(m.data)
}

// Restore restores cached data from a specified reader
func (m *InMem) Restore(r io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return gob.NewDecoder(r).Decode(m.data)
}

func (m *InMem) delete(key string) {
	delete(m.data.KV, key)
}
