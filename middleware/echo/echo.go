// Package echo provides an Echo middleware adapter for
// github.com/jassus213/go-limiter.
//
// This package allows you to easily integrate rate limiting
// into your Echo HTTP server using any Limiter implementation.
//
// Example usage:
//
//	import (
//	    "net/http"
//	    "time"
//
//	    "github.com/labstack/echo/v4"
//	    ratelimiter "github.com/jassus213/go-limiter"
//	    "github.com/jassus213/go-limiter/middleware/echo"
//	)
//
//	func main() {
//	    store := ratelimiter.NewMemoryStore()
//	    limiter := ratelimiter.NewFixedWindow(store, 100, time.Minute)
//
//	    e := echo.New()
//	    e.Use(echo.RateLimiter(limiter))
//
//	    e.GET("/ping", func(c echo.Context) error {
//	        return c.String(http.StatusOK, "pong")
//	    })
//
//	    e.Start(":8080")
//	}
package echo

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jassus213/go-limiter/ratelimiter"
	"github.com/labstack/echo/v4"
)

// RateLimiter returns an Echo middleware that enforces rate limiting.
func RateLimiter(limiter ratelimiter.Limiter, options ...ratelimiter.Option) echo.MiddlewareFunc {
	cfg := ratelimiter.NewConfig(options...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key, err := cfg.KeyFunc(c.Request())
			if err != nil {
				cfg.Logger.Errorf("[RateLimiter] Failed to extract key: %v", err)
				return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}

			result, err := limiter.Allow(c.Request().Context(), key)
			if err != nil {
				cfg.Logger.Errorf("[RateLimiter] Limiter failed for key '%s': %v", key, err)
				return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}

			c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))

			resetTimestamp := time.Now().Add(result.ResetAfter).Unix()
			c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTimestamp, 10))

			if !result.Allowed {
				cfg.Logger.Debugf(
					"[RateLimiter] Request denied for key '%s'. Remaining: %d, Limit: %d",
					key, result.Remaining, result.Limit,
				)
				cfg.ErrorHandler(c.Response(), c.Request(), ratelimiter.ErrorExceeded, result)
				return nil
			}

			cfg.Logger.Debugf(
				"[RateLimiter] Request allowed for key '%s'. Remaining: %d, Limit: %d",
				key, result.Remaining, result.Limit,
			)

			return next(c)
		}
	}
}
