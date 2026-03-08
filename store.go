package ratelimiter

import (
	"context"
	"fmt"
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

func (s *Store) StartCleanup(ctx context.Context, interval time.Duration, ttl time.Duration) {
	go func() {
		fmt.Println("[cleanup] goroutine started")
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanup(ttl)
			case <-ctx.Done():
				fmt.Println("[cleanup] goroutine stopped")
				return
			}

		}
	}()
}

func (s *Store) cleanup(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	fmt.Printf("[cleanup] sweeping %d buckets\n", len(s.buckets))
	for clientID, bucket := range s.buckets {
		bucket.mu.Lock()
		lastAccess := bucket.lastAccessTime
		bucket.mu.Unlock()

		if now.Sub(lastAccess) > ttl {
			fmt.Printf("[cleanup] deleting stale bucket: %s (idle %s)\n", clientID, now.Sub(lastAccess))
			delete(s.buckets, clientID)
		}
	}
}
