package ratelimiter

import (
	"sync"
	"time"
)

type Store struct {
	mu         sync.Mutex
	buckets    map[string]*TokenBucket
	capacity   float64
	refillRate float64
}

func NewStore(capacity, refillRate float64) *Store {
	return &Store{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (s *Store) GetBucket(key string) *TokenBucket {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			capacity:       s.capacity,
			refillRate:     s.refillRate,
			tokens:         s.capacity,
			lastAccessTime: time.Now(),
		}
		s.buckets[key] = bucket
	}

	return bucket
}
