package cache

import (
	"context"
	"encoding/gob"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// ErrNotFound returns when key not found in cache
var ErrNotFound = errors.New("key not found")

// InMem in-memory cache
type InMem struct {
	mu    sync.RWMutex
	data  data
	stats stats
}

type data struct {
	Index btree
	KV    map[string]string
}

type stats struct {
	miss int64
	hit  int64
}

// New returns new in-memory cache
func New() *InMem {
	return &InMem{
		mu: sync.RWMutex{},
		data: data{
			Index: btree{},
			KV:    make(map[string]string),
		},
		stats: stats{
			miss: 0,
			hit:  0,
		},
	}
}

// Get returns value by key
func (m *InMem) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if v, ok := m.data.KV[key]; ok {
		atomic.AddInt64(&m.stats.hit, 1)
		return v, nil
	}

	atomic.AddInt64(&m.stats.miss, 1)
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

// Stats returns hits, misses and cache size
func (m *InMem) Stats() (int64, int64, int64) {
	m.mu.RLock()
	hit := m.stats.hit
	miss := m.stats.miss
	size := int64(len(m.data.KV))
	m.mu.RUnlock()

	return hit, miss, size
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

	return gob.NewDecoder(r).Decode(&m.data)
}

func (m *InMem) delete(key string) {
	delete(m.data.KV, key)
}
