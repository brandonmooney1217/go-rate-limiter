package main

import (
	"context"
	"fmt"
	ratelimiter "go-rate-limiter"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Create a context that will be canceled on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create a new store with a capacity of 10 tokens and a refill rate of 2 tokens per second
	store := ratelimiter.NewStore(5, 0.2)

	// cleanup old buckets every 5 seconds, removing buckets idle for more than 10 seconds (short for testing)
	store.StartCleanup(ctx, time.Second*5, time.Second*10)

	// handler that will be wrapped by the rate limiter middleware
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello!")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: store.Middleware(hello),
	}

	go server.ListenAndServe()
	fmt.Println("Server running on :8080")

	<-ctx.Done()
	fmt.Println("Shutting down...")
	server.Shutdown(context.Background())

}
