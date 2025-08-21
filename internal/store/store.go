package store

import (
	"hash/fnv"
	"sync"
)

type shard struct {
	mu sync.RWMutex
	m  map[string]string
}
type Store struct {
	shards []shard
	mask   uint32
}

func New() *Store { return NewSharded(64) } // default 64 shards

func NewSharded(n int) *Store {
	// n must be power of two for fast masking
	if n <= 0 {
		n = 1
	}
	// round up to power of two
	p := 1
	for p < n {
		p <<= 1
	}
	n = p

	s := &Store{shards: make([]shard, n), mask: uint32(n - 1)}
	for i := range s.shards {
		s.shards[i].m = make(map[string]string)
	}
	return s
}

func (s *Store) shardFor(key string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := h.Sum32() & s.mask
	return &s.shards[idx]
}

func (s *Store) Get(k string) (string, bool) {
	sh := s.shardFor(k)
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	v, ok := sh.m[k]
	return v, ok
}
func (s *Store) Set(k, v string) {
	sh := s.shardFor(k)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.m[k] = v
}
func (s *Store) Del(k string) bool {
	sh := s.shardFor(k)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if _, ok := sh.m[k]; ok {
		delete(sh.m, k)
		return true
	}
	return false
}
