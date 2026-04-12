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
)

func main() {
	// Create a context that is canceled when an interrupt signal is received
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create store with capacity of 10 tokens and refill rate of 1 token per second
	store := rateLimiter.NewStore(10, 1)

	store.StartCleanup(ctx, time.Second*12, time.Second*10)

	// create handler function that responds with "Hello, World!"
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	// create HTTP server with rate limiting middleware
	server := &http.Server{
		Addr:    ":8080",
		Handler: store.Middleware(hello),
	}

	fmt.Println("Server is running on http://localhost:8080")
	go server.ListenAndServe()

	<-ctx.Done()
	fmt.Println("Shutting down...")
	server.Shutdown(context.Background())
}
