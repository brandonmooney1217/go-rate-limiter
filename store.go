package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Store struct {
	mu             sync.Mutex
	buckets        map[string]*TokenBucket
	bucketCapacity float64
	refillRate     float64
}

func NewStore(capacity, refillRate float64) *Store {
	return &Store{
		buckets:        make(map[string]*TokenBucket),
		bucketCapacity: capacity,
		refillRate:     refillRate,
	}
}

func (s *Store) GetBucket(key string) *TokenBucket {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			capacity:       s.bucketCapacity,
			refillRate:     s.refillRate,
			tokens:         s.bucketCapacity,
			lastAccessTime: time.Now(),
		}
		s.buckets[key] = bucket
	}

	return bucket
}

func (s *Store) StartCleanup(ctx context.Context, interval time.Duration, ttl time.Duration) {
	// Ticker to periodically clean up expired buckets
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			// receive-only channel. Ticker sends curent time on the channel at regular intervals
			case <-ticker.C:
				s.cleanup(ttl)
			// receive only channel that is closed when the context is canceled.
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Store) cleanup(ttl time.Duration) {
	fmt.Println("Cleaning up expired buckets...")
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, bucket := range s.buckets {
		bucket.mu.Lock()
		lastAccess := bucket.lastAccessTime
		bucket.mu.Unlock()

		if now.Sub(lastAccess) > ttl {
			delete(s.buckets, key)
			fmt.Printf("Cleaned up bucket: key=%s, lastAccess=%v, age=%v\n", key, lastAccess, now.Sub(lastAccess))
		}
	}
}
