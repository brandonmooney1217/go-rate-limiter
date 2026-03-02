package ratelimiter

import (
	"net"
	"net/http"
)

func (s *Store) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID, _, _ := net.SplitHostPort(r.RemoteAddr)

		bucket := s.GetBucket(clientID)

		if !bucket.AllowN(1) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
