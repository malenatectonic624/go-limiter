package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/jassus213/go-limiter/ratelimiter"
	"github.com/jassus213/go-limiter/store"
)

func TestFixedWindowLimiter_Allow(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemory(ctx, 0)

	limit := int64(3)
	window := 100 * time.Millisecond
	limiter := ratelimiter.NewFixedWindow(memStore, limit, window)

	key := "user:123"

	res, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Fatal("expected allowed=true on first request")
	}
	if res.Remaining != limit-1 {
		t.Fatalf("expected remaining=%d, got %d", limit-1, res.Remaining)
	}

	res, _ = limiter.Allow(ctx, key)
	if !res.Allowed {
		t.Fatal("expected allowed=true on second request")
	}
	if res.Remaining != limit-2 {
		t.Fatalf("expected remaining=%d, got %d", limit-2, res.Remaining)
	}

	res, _ = limiter.Allow(ctx, key)
	if !res.Allowed {
		t.Fatal("expected allowed=true on third request")
	}
	if res.Remaining != 0 {
		t.Fatalf("expected remaining=0, got %d", res.Remaining)
	}

	res, _ = limiter.Allow(ctx, key)
	if res.Allowed {
		t.Fatal("expected allowed=false on fourth request")
	}
	if res.Remaining != 0 {
		t.Fatalf("expected remaining=0 after limit exceeded, got %d", res.Remaining)
	}

	time.Sleep(window + 10*time.Millisecond)
	res, _ = limiter.Allow(ctx, key)
	if !res.Allowed {
		t.Fatal("expected allowed=true after window reset")
	}
	if res.Remaining != limit-1 {
		t.Fatalf("expected remaining=%d after window reset, got %d", limit-1, res.Remaining)
	}
}

func TestFixedWindowLimiter_ResetAfter(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemory(ctx, 0)

	limiter := ratelimiter.NewFixedWindow(memStore, 1, 200*time.Millisecond)
	key := "user:reset"

	res, _ := limiter.Allow(ctx, key)
	if res.ResetAfter <= 0 {
		t.Fatalf("expected positive ResetAfter, got %v", res.ResetAfter)
	}

	time.Sleep(150 * time.Millisecond)
	res, _ = limiter.Allow(ctx, key)
	if res.ResetAfter <= 0 {
		t.Fatalf("expected positive ResetAfter near window end, got %v", res.ResetAfter)
	}
}
