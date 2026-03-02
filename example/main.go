package main

import (
	"fmt"
	ratelimiter "go-rate-limiter"
	"net/http"
)

func main() {

	// create a new store with a capacity of 10 tokens and a refill rate of 2 tokens per second
	store := ratelimiter.NewStore(5, 0.2)

	// handler that will be wrapped by the rate limiter middleware
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello!")
	})

	http.Handle("/", store.Middleware(hello))
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)

}
