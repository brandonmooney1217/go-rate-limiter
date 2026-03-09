package ratelimiter

import "time"

type Result struct {
	Allowed    bool
	Limit      int           // X-RateLimit-Limit: total capacity
	Remaining  int           // X-RateLimit-Remaining: tokens left
	RetryAfter time.Duration // Retry-After: time until enough tokens are available (only set when Allowed=false)
}
