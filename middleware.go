package ratelimiter

import (
	"net"
	"net/http"
	"strconv"
)

func RateLimitMiddleware(store BucketStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID, _, _ := net.SplitHostPort(r.RemoteAddr)

			result := store.AllowN(r.Context(), clientID, 1)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())+1))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
