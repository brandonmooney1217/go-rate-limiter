package ratelimiter

type RateLimiter interface {
	AllowN(n int) bool
}
