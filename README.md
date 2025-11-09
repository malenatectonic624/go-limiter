# go‑limiter

🚀 **go-limiter** is a lightweight, flexible, and extensible rate limiting library for Go.  
It provides multiple algorithms, pluggable stores, and ready-to-use middleware integrations.

---

A flexible and modular rate‑limiting library for Go, with pluggable storage back‑ends (in‑memory, Redis) and middleware adapters for popular frameworks like Gin, Echo, Chi and the standard net/http.

---

## Features
* Two built‑in algorithms:
  * Fixed Window — simple time‑window counter limiting.
  * Token Bucket — allows bursts while enforcing steady rate.
* Pluggable store/back‑end interface (Store):
  * In‑memory store (single‑instance): fast, no external service needed.
  * Redis store (distributed apps): shared state across instances.
* Middleware adapters out‑of‑the‑box for:
  * `net/http`
  * `Gin`
  * `Echo`
  * `Chi`

---

# Getting Started

```bash
go get github.com/jassus213/go‑limiter
```

### Using Fixed Window
```go
ctx := context.Background()
store := store.NewMemory(ctx, time.Minute) // in‑memory
limiter := ratelimiter.NewFixedWindow(store, 100, time.Minute)

result, err := limiter.Allow(ctx, "user:123")
if err != nil {
    // handle error
}
if result.Allowed {
    // proceed request
} else {
    // reject or throttle
}
```

### Using Token Bucket
```go
ctx := context.Background()
store := store.NewMemory(ctx, time.Minute)
limiter := ratelimiter.NewTokenBucket(store, 1.0, 5) // 1 token/sec, burst 5

result, err := limiter.Allow(ctx, "user:123")
if err != nil {
    // handle error
}
if result.Allowed {
    // proceed
} else {
    // throttle until result.ResetAfter
}
```

### Redis Store (distributed mode)
```go
client := redis.NewClient(&redis.Options{ Addr: "localhost:6379" })
store := store.NewRedis(client)
limiter := ratelimiter.NewFixedWindow(store, 100, time.Minute)
```

---

# Middleware Usage
## net/http
```bash
go get github.com/jassus213/go-limiter/middleware/nethttp
```

```go
http.Handle("/", limiter.RateLimiter(limiter))
```

## Gin
```bash
go get github.com/jassus213/go-limiter/middleware/gin
```

```go
r := gin.Default()
r.Use(ginmiddleware.RateLimiter(limiter))
```

## Echo
```bash
go get github.com/jassus213/go-limiter/middleware/echo
```

```go
e := echo.New()
e.Use(echomiddleware.RateLimiter(limiter))
```

## Chi
```bash
go get github.com/jassus213/go-limiter/middleware/chi
```

```go
r := chi.NewRouter()
r.Use(chimiddleware.RateLimiter(limiter))
```

---

# 🧾 Logging
`go-limiter` provides pluggable logging modules under the `logger/` package.  
You can use the standard `log` package or popular structured loggers like Logrus, Zap, or Zerolog.

Install only what you need:

```bash
# standard log (no extra deps)
go get github.com/jassus213/go-limiter/logger/log

# logrus integration
go get github.com/jassus213/go-limiter/logger/logrus

# zap integration
go get github.com/jassus213/go-limiter/logger/zap

# zerolog integration
go get github.com/jassus213/go-limiter/logger/zerolog
```

Each logger implements a common interface, so you can pass it into limiter options:

## 🧱 Example — Standard Logger with Gin Middleware

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	stdlogadapter "github.com/jassus213/go-limiter/logger/log"
	ginMiddleware "github.com/jassus213/go-limiter/middleware/gin"
	"github.com/jassus213/go-limiter/ratelimiter"
	"github.com/jassus213/go-limiter/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a standard logger adapter
	stdLogger := stdlogadapter.New(log.Default())

	// In-memory token bucket limiter
	limiterStore := store.NewMemory(ctx, 10*time.Minute)
	limiter := ratelimiter.NewTokenBucket(limiterStore, 1.0, 5)

	// Configure limiter options
	config := []ratelimiter.Option{
		ratelimiter.WithLogger(stdLogger),
		ratelimiter.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error, result ratelimiter.Result) {
			stdLogger.Errorf(
				"Rate limit exceeded for key: %s | Remaining: %d | Limit: %d",
				r.RemoteAddr, result.Remaining, result.Limit,
			)
			retryAfter := int(result.ResetAfter.Seconds())
			if retryAfter <= 0 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}),
	}

	// Setup Gin router with middleware
	router := gin.Default()
	router.Use(ginMiddleware.RateLimiter(limiter, config...))

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	log.Println("Server running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
```

## 🔹Logrus Example
```go
import (
	"github.com/sirupsen/logrus"
	llog "github.com/jassus213/go-limiter/logger/logrus"
)

base := logrus.New()
logger := llog.New(base)

config := []ratelimiter.Option{
	ratelimiter.WithLogger(logger),
}

router.Use(ginMiddleware.RateLimiter(limiter, config...))
```

## 🔹Zap Example
```go
import (
	"go.uber.org/zap"
	zlog "github.com/jassus213/go-limiter/logger/zap"
)

base, _ := zap.NewProduction()
logger := zlog.New(base)

config := []ratelimiter.Option{
	ratelimiter.WithLogger(logger),
}

router.Use(ginMiddleware.RateLimiter(limiter, config...))
```

## 🔹Zerolog Example
```go
import (
	"os"
	"github.com/rs/zerolog"
	zl "github.com/jassus213/go-limiter/logger/zerolog"
)

base := zerolog.New(os.Stdout).With().Timestamp().Logger()
logger := zl.New(base)

config := []ratelimiter.Option{
	ratelimiter.WithLogger(logger),
}

router.Use(ginMiddleware.RateLimiter(limiter, config...))
```

# HTTP Headers
When a request is allowed/denied, middleware sets HTTP headers for rate‑limit info:
* `X‑RateLimit‑Limit` - the maximum number of requests (or tokens) allowed.
* `X‑RateLimit‑Remaining` - how many left in current window/bucket.
* `X‑RateLimit‑Reset` - time (as Unix timestamp or duration) when the quota resets.

---

# Configuration & Options
You can pass optional configuration for the limiter (via Option pattern) to customize:
* Key extraction (e.g., by IP, user ID, header).
* Error handler behaviour.
* Logging implementation.

---

# License
This project is licensed under the MIT License. See the LICENSE file for details.