package ratelimiter

type RateLimiter interface {
	AllowN(n int) bool
	AllowNResult(n int) Result
}

type BucketStore interface {
	AllowN(key string, n int) Result
}
