package main

import (
	"context"
	"fmt"
	rateLimiter "go-rate-limiter"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Create a context that is canceled when an interrupt signal is received
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var store rateLimiter.BucketStore

	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		client := redis.NewClient(&redis.Options{Addr: addr})
		if err := client.Ping(ctx).Err(); err != nil {
			fmt.Printf("Failed to connect to Redis at %s: %v\n", addr, err)
			os.Exit(1)
		}
		store = rateLimiter.NewRedisStore(client, 10, 1, 60)
		fmt.Printf("Using Redis store at %s\n", addr)
	} else {
		memStore := rateLimiter.NewStore(10, 1)
		memStore.StartCleanup(ctx, time.Second*12,
			time.Second*10)
		store = memStore
		fmt.Println("Using in-memory store")
	}

	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	// create HTTP server with rate limiting middleware
	server := &http.Server{
		Addr:    ":8080",
		Handler: rateLimiter.RateLimitMiddleware(store)(hello),
	}
	fmt.Println("Server is running on http://localhost:8080")
	go server.ListenAndServe()

	<-ctx.Done()
	fmt.Println("Shutting down...")
	server.Shutdown(context.Background())
}
