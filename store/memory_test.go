package store

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestMemoryStore_Increment(t *testing.T) {
	ctx := context.Background()
	s := NewMemory(ctx, 0)

	key := "user:123"
	window := 100 * time.Millisecond

	count, err := s.Increment(ctx, key, window)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected count 1, got %d", count)
	}

	count, _ = s.Increment(ctx, key, window)
	if count != 2 {
		t.Fatalf("Expected count 2, got %d", count)
	}

	time.Sleep(window + 10*time.Millisecond)
	count, _ = s.Increment(ctx, key, window)
	if count != 1 {
		t.Fatalf("Expected count reset to 1, got %d", count)
	}
}

func TestMemoryStore_TakeToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemory(ctx, 0)

	key := "user:abc"
	rate := 1.0
	burst := int64(2)
	epsilon := 1e-6

	allowed, remaining, _ := s.TakeToken(ctx, key, rate, burst)
	if !allowed {
		t.Fatalf("Expected allowed true")
	}
	if math.Abs(remaining-float64(burst-1)) > epsilon {
		t.Fatalf("Expected remaining %f, got %f", float64(burst-1), remaining)
	}

	allowed, remaining, _ = s.TakeToken(ctx, key, rate, burst)
	if !allowed {
		t.Fatalf("Expected allowed true")
	}
	if math.Abs(remaining-0) > epsilon {
		t.Fatalf("Expected remaining 0, got %f", remaining)
	}

	allowed, remaining, _ = s.TakeToken(ctx, key, rate, burst)
	if allowed {
		t.Fatalf("Expected allowed false")
	}

	time.Sleep(time.Second + 10*time.Millisecond)

	allowed, remaining, _ = s.TakeToken(ctx, key, rate, burst)
	if !allowed {
		t.Fatalf("Expected allowed true after refill")
	}
	if remaining <= 0 {
		t.Fatalf("Expected remaining > 0 after refill, got %f", remaining)
	}
}

func TestMemoryStore_RunCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanupInterval := 20 * time.Millisecond
	s := NewMemory(ctx, cleanupInterval)

	key := "user:cleanup"
	s.Increment(ctx, key, 10*time.Millisecond)
	s.TakeToken(ctx, "token:cleanup", 1, 1)

	time.Sleep(25 * cleanupInterval)

	ms := s.(*MemoryStore)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, found := ms.fixedWindowEntries[key]; found {
		t.Fatalf("Expected fixed window entry to be cleaned up")
	}
	if _, found := ms.tokenBucketEntries["token:cleanup"]; found {
		t.Fatalf("Expected token bucket entry to be cleaned up")
	}
}
