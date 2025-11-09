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
```go
http.Handle("/", limiter.RateLimiter(limiter))
```

## Gin
```go
r := gin.Default()
r.Use(ginmiddleware.RateLimiter(limiter))
```

## Echo
```go
e := echo.New()
e.Use(echomiddleware.RateLimiter(limiter))
```

## Chi
```go
r := chi.NewRouter()
r.Use(chimiddleware.RateLimiter(limiter))
```

---

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