package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/jassus213/go-limiter/ratelimiter"
	"github.com/jassus213/go-limiter/store"
)

func TestTokenBucketLimiter_Allow(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemory(ctx, 0)

	rate := 1.0
	burst := int64(3)
	limiter := ratelimiter.NewTokenBucket(memStore, rate, burst)

	key := "user:token"

	// Первый запрос — должно быть разрешено
	res, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Fatal("expected allowed=true on first request")
	}
	if res.Remaining != burst-1 {
		t.Fatalf("expected remaining=%d, got %d", burst-1, res.Remaining)
	}

	res, _ = limiter.Allow(ctx, key)
	if !res.Allowed {
		t.Fatal("expected allowed=true on second request")
	}
	if res.Remaining != burst-2 {
		t.Fatalf("expected remaining=%d, got %d", burst-2, res.Remaining)
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

	time.Sleep(time.Second + 10*time.Millisecond)
	res, _ = limiter.Allow(ctx, key)
	if !res.Allowed {
		t.Fatal("expected allowed=true after token refill")
	}
	if res.Remaining != 0 {
		t.Fatalf("expected remaining=0 after taking one token, got %d", res.Remaining)
	}
}

func TestTokenBucketLimiter_ResetAfter(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemory(ctx, 0)

	limiter := ratelimiter.NewTokenBucket(memStore, 1.0, 1)
	key := "user:reset"

	res, _ := limiter.Allow(ctx, key)
	if res.Allowed == false {
		t.Fatal("expected allowed=true on first token")
	}

	res, _ = limiter.Allow(ctx, key)
	if res.Allowed {
		t.Fatal("expected allowed=false when no tokens")
	}
	if res.ResetAfter <= 0 {
		t.Fatalf("expected positive ResetAfter, got %v", res.ResetAfter)
	}
}
