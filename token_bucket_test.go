package ratelimiter

import (
	"testing"
	"time"
)

func TestAllowN_FreshBucket(t *testing.T) {
	store := NewStore(10, 1)
	bucket := store.GetBucket("client1")

	if !bucket.AllowN(5) {
		t.Errorf("Expected to allow 5 tokens, but was denied")
	}
}

func TestAllowN_InsufficientTokens(t *testing.T) {
	store := NewStore(10, 1)
	bucket := store.GetBucket("client1")

	if !bucket.AllowN(5) {
		t.Errorf("Expected to allow 5 tokens, but was denied")
	}

	if bucket.AllowN(6) {
		t.Errorf("Expected to deny 6 tokens, but was allowed")
	}
}

func TestAllowN_ExhaustedBucket(t *testing.T) {
	store := NewStore(10, 1)
	bucket := store.GetBucket("client1")

	if !bucket.AllowN(10) {
		t.Errorf("Expected to allow 10 tokens, but was denied")
	}

	if bucket.AllowN(1) {
		t.Errorf("Expected to deny 1 token from empty bucket, but was allowed")
	}
}

func TestAllowN_RefillAfterWait(t *testing.T) {
	store := NewStore(10, 100)
	bucket := store.GetBucket("client1")

	if !bucket.AllowN(10) {
		t.Errorf("Expected to allow 10 tokens, but was denied")
	}

	time.Sleep(50 * time.Millisecond) // ~5 tokens refill at 100/sec

	if !bucket.AllowN(5) {
		t.Errorf("Expected to allow 5 tokens after refill, but was denied")
	}
}
