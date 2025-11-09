//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jassus213/go-limiter/store"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (context.Context, testcontainers.Container, *redis.Client) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis host: %v", err)
	}

	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get Redis port: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Ждём, пока Redis будет доступен
	for i := 0; i < 10; i++ {
		if err := client.Ping(ctx).Err(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return ctx, redisC, client
}

func TestRedisStore_WithTestcontainers(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	defer redisC.Terminate(ctx)

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis host: %v", err)
	}

	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get redis port: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})
	defer rdb.Close()

	redisStore := store.NewRedis(rdb)

	key := "token:test"
	rate := 1.0
	burst := int64(2)

	allowed, remaining, err := redisStore.TakeToken(ctx, key, rate, burst)
	if err != nil || !allowed {
		t.Fatalf("expected first token allowed, got allowed=%v, err=%v", allowed, err)
	}

	allowed, remaining, err = redisStore.TakeToken(ctx, key, rate, burst)
	if err != nil || !allowed {
		t.Fatalf("expected second token allowed, got allowed=%v, err=%v", allowed, err)
	}

	allowed, remaining, err = redisStore.TakeToken(ctx, key, rate, burst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatalf("expected third token not allowed, got remaining=%f", remaining)
	}

	time.Sleep(time.Second + 50*time.Millisecond)

	allowed, remaining, err = redisStore.TakeToken(ctx, key, rate, burst)
	if err != nil {
		t.Fatalf("unexpected error after refill: %v", err)
	}

	epsilon := 0.01
	if !allowed || remaining < epsilon {
		t.Fatalf("expected token allowed after refill, got allowed=%v, remaining=%f", allowed, remaining)
	}
}
