# Go Rate Limiter

## Overview
In-memory rate limiting library in Go with HTTP middleware. Uses the token bucket algorithm with lazy refill.

## Project Structure
- `limiter.go` — `RateLimiter` interface (`AllowN(n int) bool`)
- `token_bucket.go` — Token bucket implementation (lazy refill, mutex-protected)
- `store.go` — Per-client bucket registry (`sync.Mutex` + `map[string]*TokenBucket`)
- `middleware.go` — HTTP middleware that wraps an `http.Handler`
- `token_bucket_test.go` — Unit tests
- `example/main.go` — Demo HTTP server

## Design Decisions
- **Lazy refill** over background goroutines — tokens calculated on each `AllowN` call based on elapsed time
- **Two mutexes** — one on the store (protects the map), one per bucket (protects token state)
- **`float64` for tokens/rate/capacity** — avoids losing fractional tokens during refill math
- **`net.SplitHostPort`** on `r.RemoteAddr` — strips ephemeral port so clients are identified by IP only
- **Package name:** `ratelimiter` (library, not `main`)
- **Stale bucket cleanup** — without cleanup, every unique client IP adds a bucket to the map that never gets removed (unbounded growth). `Store.StartCleanup(ctx, interval, ttl)` launches a background goroutine that periodically sweeps the map and deletes buckets whose `lastAccessTime` exceeds the TTL. Uses `time.Ticker` for periodic execution and `context.Context` for graceful shutdown (goroutine exits cleanly on cancellation). Lock ordering: store mutex acquired first, then bucket mutex (to read `lastAccessTime` safely), bucket mutex released before delete.
- **Graceful shutdown** — `example/main.go` uses `signal.NotifyContext` to create a context that cancels on SIGINT/SIGTERM. This context is passed to `StartCleanup` and used to block `main()` via `<-ctx.Done()`. On Ctrl+C: cleanup goroutine stops, `http.Server.Shutdown` drains in-flight requests.

## Design Doc
Full design doc at: `/Users/brandonmooney/Downloads/Development_Projects/swe/Go/rate-limiter-design.md`

## Running
- Tests: `go test ./...`
- Demo server: `go run example/main.go` then `curl http://localhost:8080/`

## TODO
- Additional algorithms (sliding window, leaky bucket)
- Response headers (`X-RateLimit-Remaining`, `Retry-After`)
