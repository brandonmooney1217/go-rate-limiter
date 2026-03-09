package ratelimiter

type RateLimiter interface {
	AllowN(n int) bool

	AllowNResult(n int) Result
}
