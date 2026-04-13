package ratelimiter

import "context"

type RateLimiter interface {
	AllowN(n int) bool
	AllowNResult(n int) Result
}

type BucketStore interface {
	AllowN(ctx context.Context, key string, n int) Result
}
