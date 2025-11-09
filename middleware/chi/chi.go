// Package chi provides a Chi middleware adapter for
// github.com/jassus213/go-limiter.
//
// This package allows you to easily integrate rate limiting
// into your Chi HTTP server using any Limiter implementation.
//
// Example usage:
//
//	import (
//	    "net/http"
//	    "time"
//
//	    "github.com/go-chi/chi/v5"
//	    ratelimiter "github.com/jassus213/go-limiter"
//	    "github.com/jassus213/go-limiter/middleware/chi"
//	)
//
//	func main() {
//	    store := ratelimiter.NewMemoryStore()
//	    limiter := ratelimiter.NewFixedWindow(store, 100, time.Minute)
//
//	    r := chi.NewRouter()
//	    r.Use(chi.RateLimiter(limiter))
//
//	    r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
//	        w.Write([]byte("pong"))
//	    })
//
//	    http.ListenAndServe(":8080", r)
//	}
package chi

import (
	"net/http"
	"strconv"
	"time"

	ratelimiter "github.com/jassus213/go-limiter/ratelimiter"
)

// RateLimiter returns a Chi middleware that enforces rate limiting.
func RateLimiter(limiter ratelimiter.Limiter, options ...ratelimiter.Option) func(next http.Handler) http.Handler {
	cfg := ratelimiter.NewConfig(options...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, err := cfg.KeyFunc(r)
			if err != nil {
				cfg.Logger.Errorf("[RateLimiter] Failed to extract key: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			result, err := limiter.Allow(r.Context(), key)
			if err != nil {
				cfg.Logger.Errorf("[RateLimiter] Limiter failed for key '%s': %v", key, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))

			resetTimestamp := time.Now().Add(result.ResetAfter).Unix()
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTimestamp, 10))

			if !result.Allowed {
				cfg.Logger.Debugf(
					"[RateLimiter] Request denied for key '%s'. Remaining: %d, Limit: %d",
					key, result.Remaining, result.Limit,
				)
				cfg.ErrorHandler(w, r, ratelimiter.ErrorExceeded, result)
				return
			}

			cfg.Logger.Debugf(
				"[RateLimiter] Request allowed for key '%s'. Remaining: %d, Limit: %d",
				key, result.Remaining, result.Limit,
			)

			next.ServeHTTP(w, r)
		})
	}
}
