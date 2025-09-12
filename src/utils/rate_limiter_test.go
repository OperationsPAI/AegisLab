package utils

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/LGU-SE-Internal/rcabench/config"
)

func TestTokenBucketRateLimiter_AcquireAndRelease(t *testing.T) {

	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_DB", "0")
	os.Setenv("REDIS_PASSWORD", "")
	config.Init("../")

	limiter := NewTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   "test:token_bucket",
		MaxTokensKey:     "rate_limiting.max_concurrent_test",
		DefaultMaxTokens: 2,
		DefaultTimeout:   3,
		ServiceName:      "test_service",
	})

	ctx := context.Background()

	limiter.redisClient.Del(ctx, "test:token_bucket")

	taskID1 := "task1"
	taskID2 := "task2"
	taskID3 := "task3"
	traceID := "trace-test"

	acquired1, err := limiter.AcquireToken(ctx, taskID1, traceID)
	if err != nil || !acquired1 {
		t.Fatalf("expected to acquire token for task1, got %v, err: %v", acquired1, err)
	}
	acquired2, err := limiter.AcquireToken(ctx, taskID2, traceID)
	if err != nil || !acquired2 {
		t.Fatalf("expected to acquire token for task2, got %v, err: %v", acquired2, err)
	}
	acquired3, err := limiter.AcquireToken(ctx, taskID3, traceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acquired3 {
		t.Fatalf("should not acquire token for task3, bucket full")
	}

	err = limiter.ReleaseToken(ctx, taskID1, traceID)
	if err != nil {
		t.Fatalf("failed to release token: %v", err)
	}
	acquired3, err = limiter.AcquireToken(ctx, taskID3, traceID)
	if err != nil || !acquired3 {
		t.Fatalf("expected to acquire token for task3 after release, got %v, err: %v", acquired3, err)
	}
}

func TestTokenBucketRateLimiter_WaitForToken(t *testing.T) {
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_DB", "0")
	os.Setenv("REDIS_PASSWORD", "")
	config.Init("../")

	limiter := NewTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   "test:token_bucket_wait",
		MaxTokensKey:     "rate_limiting.max_concurrent_test_wait",
		DefaultMaxTokens: 1,
		DefaultTimeout:   2,
		ServiceName:      "test_service_wait",
	})

	ctx := context.Background()
	limiter.redisClient.Del(ctx, "test:token_bucket_wait")

	taskID1 := "task1"
	taskID2 := "task2"
	traceID := "trace-test"

	acquired1, err := limiter.AcquireToken(ctx, taskID1, traceID)
	if err != nil || !acquired1 {
		t.Fatalf("expected to acquire token for task1, got %v, err: %v", acquired1, err)
	}
	start := time.Now()
	acquired2, err := limiter.WaitForToken(ctx, taskID2, traceID)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acquired2 {
		t.Fatalf("should not acquire token for task2, should timeout")
	}
	if duration < 2*time.Second {
		t.Fatalf("wait duration too short, expected at least 2s, got %v", duration)
	}

	err = limiter.ReleaseToken(ctx, taskID1, traceID)
	if err != nil {
		t.Fatalf("failed to release token: %v", err)
	}
	acquired2, err = limiter.WaitForToken(ctx, taskID2, traceID)
	if err != nil || !acquired2 {
		t.Fatalf("expected to acquire token for task2 after release, got %v, err: %v", acquired2, err)
	}
}
