package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu             sync.Mutex
	capacity       float64
	tokens         float64
	refillRate     float64
	lastAccessTime time.Time
}

func (tb *TokenBucket) AllowNResult(n int) Result {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return Result{
			Allowed:   true,
			Limit:     int(tb.capacity),
			Remaining: int(tb.tokens),
		}
	}
	return Result{
		Allowed:    false,
		Limit:      int(tb.capacity),
		Remaining:  int(tb.tokens),
		RetryAfter: time.Duration((float64(n) - tb.tokens) / tb.refillRate * float64(time.Second)),
	}
}

func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastAccessTime).Seconds()
	tokensToAdd := elapsed * tb.refillRate
	tb.tokens += tokensToAdd
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastAccessTime = now
}
