package main

import (
	"fmt"
	rateLimiter "go-rate-limiter"
	"net/http"
)

func main() {
	// create store with capacity of 10 tokens and refill rate of 1 token per second
	store := rateLimiter.NewStore(10, 1)

	// create handler functionthat responds with "Hello, World!"
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	// create HTTP server with rate limiting middleware
	server := &http.Server{
		Addr:    ":8080",
		Handler: store.Middleware(hello),
	}

	fmt.Println("Server is running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}

}
