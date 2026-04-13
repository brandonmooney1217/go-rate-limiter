package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client         *redis.Client
	bucketCapacity float64
	refillRate     float64
	ttl            int64
}

var rateLimitScript = redis.NewScript(
	`
	local key = KEYS[1]
	local capacity = tonumber(ARGV[1])
	local refill_rate = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])
	local ttl = tonumber(ARGV[5])

	-- read current state (or defaults for new keys)
	local tokens = tonumber(redis.call('HGET', key,
	'tokens')) or capacity
	local last_refill = tonumber(redis.call('HGET',
	key, 'last_refill')) or now

	-- refill (same as your tb.refill())
	local elapsed = (now - last_refill) / 1000
	tokens = math.min(capacity, tokens + elapsed *
	refill_rate)

	-- check and decrement (same as your tb.AllowN())
	local allowed = 0
	if tokens >= requested then
		tokens = tokens - requested
		allowed = 1
	end

	-- write state back + auto-expire stale keys
	redis.call('HSET', key, 'tokens', tokens,
	'last_refill', now)
	redis.call('EXPIRE', key, ttl)

	-- calculate retry-after if denied
	local retry_after = 0
	if allowed == 0 then
		retry_after = (requested - tokens) /
	refill_rate
	end

	return {allowed, math.floor(tokens), capacity,
	math.floor(retry_after * 1000)}`)

func NewRedisStore(client *redis.Client, capacity, refillRate float64, ttl int64) *RedisStore {
	return &RedisStore{
		client:         client,
		bucketCapacity: capacity,
		refillRate:     refillRate,
		ttl:            ttl,
	}
}

func (s *RedisStore) AllowN(ctx context.Context, key string, n int) Result {
	res, err := rateLimitScript.Run(ctx, s.client, []string{key},
		s.bucketCapacity, s.refillRate, time.Now().UnixMilli(), n, s.ttl).Int64Slice()

	if err != nil {
		fmt.Printf("Redis error: %v\n", err)
		return Result{
			Allowed: false,
		}
	}

	return Result{
		Allowed:    res[0] == 1,
		Remaining:  int(res[1]),
		Limit:      int(res[2]),
		RetryAfter: time.Duration(res[3]) * time.Millisecond,
	}
}
