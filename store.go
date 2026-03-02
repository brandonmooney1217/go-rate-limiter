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

func (s *Store) GetBucket(clientID string) *TokenBucket {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[clientID]
	if !exists {
		bucket = &TokenBucket{
			capacity:       s.capacity,
			tokens:         s.capacity,
			refillRate:     s.refillRate,
			lastAccessTime: time.Now(),
		}
		s.buckets[clientID] = bucket
	}

	return bucket
}
