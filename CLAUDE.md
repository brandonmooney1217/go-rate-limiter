# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview
In-memory rate limiting library in Go with HTTP middleware. Uses the token bucket algorithm with lazy refill. Module name: `go-rate-limiter`, package name: `ratelimiter`.

## Commands
- Run all tests: `go test ./...`
- Run tests verbosely: `go test -v ./...`
- Run a single test: `go test -run TestAllowN_RefillAfterWait ./...`
- Race detector (recommended for concurrency changes): `go test -race ./...`
- Build everything: `go build ./...`
- Run the demo server: `go run example/main.go` then `curl -i http://localhost:8080/`

## Architecture

The library is composed of three layers, each in its own file:

1. **`token_bucket_limiter.go`** — `TokenBucket` is the algorithm. Each bucket holds `tokens float64`, `capacity`, `refillRate`, `lastAccessTime`, and a `sync.Mutex`. `AllowN(n)` and `AllowNResult(n)` both call the unexported `refill()` helper, which lazily adds `elapsed * refillRate` tokens (capped at capacity) before checking availability. There is no background goroutine for refill.

2. **`store.go`** — `Store` is a per-client bucket registry: `map[string]*TokenBucket` protected by its own `sync.Mutex`. `NewStore(capacity, refillRate)` constructs the registry; `GetBucket(clientID)` returns an existing bucket or creates a new one inline (initialized with `tokens = capacity`, `lastAccessTime = time.Now()`). The store stores its own defaults so it can act as a factory.

3. **`middleware.go`** — `Store.Middleware(next http.Handler) http.Handler` wraps a handler. Per request: extracts the client IP via `net.SplitHostPort(r.RemoteAddr)`, calls `s.GetBucket(clientID).AllowNResult(1)`, sets `X-RateLimit-Limit` and `X-RateLimit-Remaining` headers on every response, and on denial sets `Retry-After` (rounded up by `+1` so clients never retry early) before returning 429.

The `RateLimiter` interface in `limiter.go` declares both `AllowN(n int) bool` and `AllowNResult(n int) Result`. `Result` (in `result.go`) carries `Allowed`, `Limit int`, `Remaining int`, and `RetryAfter time.Duration`.

## Critical Design Decisions

- **Lazy refill** over background goroutines — tokens calculated on each `AllowN` call based on elapsed time. Simpler, scales to many buckets without spawning goroutines.

- **Two separate mutexes** — `Store.mu` protects only the map (held only during lookup/insert). Each `TokenBucket.mu` protects only its token state. This lets lookups for different clients run in parallel with token math on other buckets. Lock ordering when both are needed: **store first, then bucket** (prevents deadlock).

- **`float64` for tokens/rate/capacity** — fractional tokens during refill math (e.g. `0.7 sec * 1.5 tokens/sec = 1.05 tokens`) would be lost with `int`. `Result.Limit` and `Result.Remaining` are converted to `int` at the boundary so the public API doesn't leak the internal type.

- **`time.Duration` precision in `AllowNResult`** — `RetryAfter` is calculated as `time.Duration((float64(n) - tb.tokens) / tb.refillRate * float64(time.Second))`. The cast to `time.Duration` happens **last**, after multiplying by `float64(time.Second)`. Casting earlier would truncate fractional seconds to zero nanoseconds.

- **HTTP header ordering rule** — all `w.Header().Set(...)` calls must happen **before** `http.Error(...)` or `next.ServeHTTP(...)`. Once the response body starts being written, headers are flushed and become immutable. Headers set after are silently dropped.

- **`net.SplitHostPort`** on `r.RemoteAddr` — strips the ephemeral client port so clients are identified by IP only. Without this, every connection from the same client gets a fresh bucket.

## TODO
- `Store.StartCleanup(ctx, interval, ttl)` — background goroutine using `time.Ticker` + `context.Context` to sweep stale buckets and prevent unbounded map growth. Lock ordering: store mutex first, then bucket mutex to read `lastAccessTime`, release bucket mutex before `delete`.
- Graceful shutdown in `example/main.go` — `signal.NotifyContext` to cancel on SIGINT/SIGTERM, passed to `StartCleanup` and used to block `main()` via `<-ctx.Done()`.
- Additional algorithms (sliding window, leaky bucket).

## Reference
Full design doc: `/Users/brandonmooney/Downloads/Development_Projects/swe/Go/rate-limiter-design.md`
