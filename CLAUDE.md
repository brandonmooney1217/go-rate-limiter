# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview
Rate limiting library in Go with HTTP middleware. Uses the token bucket algorithm with lazy refill. Supports pluggable backends via the `BucketStore` interface (in-memory and Redis). Module name: `go-rate-limiter`, package name: `ratelimiter`.

## Commands
- Run all tests: `go test ./...`
- Run tests verbosely: `go test -v ./...`
- Run a single test: `go test -run TestAllowN_RefillAfterWait ./...`
- Race detector (recommended for concurrency changes): `go test -race ./...`
- Build everything: `go build ./...`
- Run the demo server (in-memory): `go run example/main.go` then `curl -i http://localhost:8080/`
- Run the demo server (Redis): `REDIS_ADDR=localhost:6379 go run example/main.go`
- Start Redis locally: `docker run -d -p 6379:6379 redis`

## Architecture

The library is composed of three layers, each in its own file:

1. **`token_bucket_limiter.go`** — `TokenBucket` is the algorithm. Each bucket holds `tokens float64`, `capacity`, `refillRate`, `lastAccessTime`, and a `sync.Mutex`. `AllowN(n)` and `AllowNResult(n)` both call the unexported `refill()` helper, which lazily adds `elapsed * refillRate` tokens (capped at capacity) before checking availability. There is no background goroutine for refill.

2. **`memory_store.go`** — `MemoryStore` is a per-client bucket registry: `map[string]*TokenBucket` protected by its own `sync.Mutex`. `NewStore(capacity, refillRate)` constructs the registry; `GetBucket(clientID)` returns an existing bucket or creates a new one inline. `AllowN(key, n)` implements the `BucketStore` interface by calling `GetBucket(key).AllowNResult(n)`. `StartCleanup(ctx, interval, ttl)` launches a background goroutine that periodically sweeps stale buckets using `time.Ticker` + `context.Context`.

3. **`middleware.go`** — `RateLimitMiddleware(store BucketStore) func(http.Handler) http.Handler` is a standalone function (not a method) that accepts any `BucketStore` implementation. Per request: extracts the client IP via `net.SplitHostPort(r.RemoteAddr)`, calls `store.AllowN(clientID, 1)`, sets `X-RateLimit-Limit` and `X-RateLimit-Remaining` headers on every response, and on denial sets `Retry-After` (rounded up by `+1` so clients never retry early) before returning 429.

4. **`redis_store.go`** — `RedisStore` implements `BucketStore` using Redis as the backend. `NewRedisStore(client, capacity, refillRate, ttl)` takes an externally-created `*redis.Client`. `AllowN(ctx, key, n)` executes an atomic Lua script that performs the entire token bucket algorithm (read state → refill → check → decrement → write → set TTL) in a single Redis call. No mutex needed — Redis single-threaded execution + Lua atomicity handles concurrency. Timestamps use `UnixMilli` for sub-second precision; elapsed time is divided by 1000 in the Lua script to convert to seconds for refill math. On error, fails closed (denies request).

**Interfaces** in `limiter.go`: `RateLimiter` (bucket-level: `AllowN(n)`, `AllowNResult(n)`) and `BucketStore` (store-level: `AllowN(ctx, key, n) Result`). `Result` (in `result.go`) carries `Allowed`, `Remaining int`, `Limit int`, and `RetryAfter time.Duration`. The Lua script return array matches this field order.

## Critical Design Decisions

- **Lazy refill** over background goroutines — tokens calculated on each `AllowN` call based on elapsed time. Simpler, scales to many buckets without spawning goroutines.

- **Two separate mutexes** — `MemoryStore.mu` protects only the map (held only during lookup/insert). Each `TokenBucket.mu` protects only its token state. This lets lookups for different clients run in parallel with token math on other buckets. Lock ordering when both are needed: **store first, then bucket** (prevents deadlock).

- **`float64` for tokens/rate/capacity** — fractional tokens during refill math (e.g. `0.7 sec * 1.5 tokens/sec = 1.05 tokens`) would be lost with `int`. `Result.Limit` and `Result.Remaining` are converted to `int` at the boundary so the public API doesn't leak the internal type.

- **`time.Duration` precision in `AllowNResult`** — `RetryAfter` is calculated as `time.Duration((float64(n) - tb.tokens) / tb.refillRate * float64(time.Second))`. The cast to `time.Duration` happens **last**, after multiplying by `float64(time.Second)`. Casting earlier would truncate fractional seconds to zero nanoseconds.

- **HTTP header ordering rule** — all `w.Header().Set(...)` calls must happen **before** `http.Error(...)` or `next.ServeHTTP(...)`. Once the response body starts being written, headers are flushed and become immutable. Headers set after are silently dropped.

- **`net.SplitHostPort`** on `r.RemoteAddr` — strips the ephemeral client port so clients are identified by IP only. Without this, every connection from the same client gets a fresh bucket.

- **BucketStore interface** — abstracts at the "check rate limit" level (`AllowN(ctx, key, n) Result`), not the "get bucket" level. Redis doesn't return bucket objects — it runs the entire check atomically server-side. Both `MemoryStore` and `RedisStore` implement this interface. Middleware is decoupled from any concrete store type. Context parameter enables Redis network timeouts; `MemoryStore` accepts but ignores it.

- **Redis Lua script atomicity** — the entire read-modify-write cycle runs as a single atomic Lua script on Redis. Without this, two servers could read the same token count simultaneously and both allow a request (race condition). Redis `EXPIRE` replaces `MemoryStore.StartCleanup` — stale keys auto-delete after the configured TTL.

- **Redis client passed in, not created internally** — `NewRedisStore` takes a `*redis.Client` for separation of concerns (connection config is the caller's responsibility), testability, and connection reuse.

- **Stale bucket cleanup** — `MemoryStore.StartCleanup(ctx, interval, ttl)` launches a background goroutine using `time.Ticker` + `context.Context` to sweep stale buckets and prevent unbounded map growth. Lock ordering: store mutex first, then bucket mutex to read `lastAccessTime`, release bucket mutex before `delete`. Explicit lock/unlock (not defer) on bucket mutex inside the loop to avoid stacking defers.

- **Graceful shutdown** — `example/main.go` uses `signal.NotifyContext` to create a context that cancels on SIGINT/SIGTERM. This context is passed to `StartCleanup` and used to block `main()` via `<-ctx.Done()`. On Ctrl+C: cleanup goroutine stops, `http.Server.Shutdown` drains in-flight requests.

## TODO
- Redis integration tests (`redis_store_test.go`).
- Additional algorithms (sliding window, leaky bucket).

## Reference
Full design doc: `/Users/brandonmooney/Downloads/Development_Projects/swe/Go/rate-limiter-design.md`
