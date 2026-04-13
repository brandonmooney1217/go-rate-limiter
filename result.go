package ratelimiter

import "time"

type Result struct {
	Allowed    bool
	Remaining  int
	Limit      int
	RetryAfter time.Duration
}
