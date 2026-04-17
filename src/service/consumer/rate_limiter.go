package consumer

import (
	"context"
	"sync"
	"time"

	"aegis/config"
	"aegis/consts"
	redisinfra "aegis/infra/redis"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// RateLimiterConfig rate limiter configuration
type RateLimiterConfig struct {
	TokenBucketKey   string
	MaxTokensKey     string
	DefaultMaxTokens int
	DefaultTimeout   int
	ServiceName      string
}

// TokenBucketRateLimiter token bucket rate limiter
type TokenBucketRateLimiter struct {
	bucketKey   string
	store       tokenBucketStore
	mu          sync.RWMutex
	maxTokens   int
	waitTimeout time.Duration
	serviceName string
}

// GetConfig returns the current configuration
func (r *TokenBucketRateLimiter) GetConfig() (maxTokens int, waitTimeout time.Duration) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.maxTokens, r.waitTimeout
}

// UpdateConfig dynamically updates the rate limiter configuration
func (r *TokenBucketRateLimiter) UpdateConfig(maxTokens int, waitTimeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	logFields := logrus.Fields{
		"service": r.serviceName,
	}

	oldMaxTokens := r.maxTokens
	oldWaitTimeout := r.waitTimeout

	if maxTokens > 0 {
		r.maxTokens = maxTokens
		logFields["old_max_tokens"] = oldMaxTokens
		logFields["new_max_tokens"] = r.maxTokens
	}
	if waitTimeout > 0 {
		r.waitTimeout = waitTimeout
		logFields["old_wait_timeout"] = oldWaitTimeout
		logFields["new_wait_timeout"] = r.waitTimeout
	}

	logrus.WithFields(logFields).Info("Rate limiter configuration updated")
}

// AcquireToken acquires a token
func (r *TokenBucketRateLimiter) AcquireToken(ctx context.Context, taskID, traceID string) (bool, error) {
	span := trace.SpanFromContext(ctx)

	r.mu.RLock()
	maxTokens := r.maxTokens
	r.mu.RUnlock()

	acquired, err := r.store.acquire(ctx, maxTokens, taskID, traceID)
	if err != nil {
		span.RecordError(err)
		return false, err
	}
	if acquired {
		span.AddEvent("token acquired successfully")
		logrus.WithFields(logrus.Fields{
			"task_id":    taskID,
			"trace_id":   traceID,
			"service":    r.serviceName,
			"bucket_key": r.bucketKey,
		}).Info("Successfully acquired token")
	}

	return acquired, nil
}

// ReleaseToken releases a token
func (r *TokenBucketRateLimiter) ReleaseToken(ctx context.Context, taskID, traceID string) error {
	span := trace.SpanFromContext(ctx)

	result, err := r.store.release(ctx, taskID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if result > 0 {
		span.AddEvent("token released successfully")
		logrus.WithFields(logrus.Fields{
			"task_id":    taskID,
			"trace_id":   traceID,
			"service":    r.serviceName,
			"bucket_key": r.bucketKey,
		}).Info("Successfully released token")
	}

	return nil
}

// WaitForToken waits for a token, returns false if timeout
func (r *TokenBucketRateLimiter) WaitForToken(ctx context.Context, taskID, traceID string) (bool, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("waiting for token")

	r.mu.RLock()
	waitTimeout := r.waitTimeout
	r.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			span.AddEvent("token wait timeout")
			logrus.WithFields(logrus.Fields{
				"task_id":    taskID,
				"trace_id":   traceID,
				"timeout":    waitTimeout,
				"service":    r.serviceName,
				"bucket_key": r.bucketKey,
			}).Warn("Token wait timeout")
			return false, nil
		case <-ticker.C:
			acquired, err := r.AcquireToken(ctx, taskID, traceID)
			if err != nil {
				return false, err
			}
			if acquired {
				return true, nil
			}
		}
	}
}

func NewRestartPedestalRateLimiter(gateway *redisinfra.Gateway) *TokenBucketRateLimiter {
	return newTokenBucketRateLimiter(gateway, RateLimiterConfig{
		TokenBucketKey:   consts.RestartPedestalTokenBucket,
		MaxTokensKey:     consts.MaxTokensKeyRestartPedestal,
		DefaultMaxTokens: consts.MaxConcurrentRestartPedestal,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      consts.RestartPedestalServiceName,
	})
}

func NewBuildContainerRateLimiter(gateway *redisinfra.Gateway) *TokenBucketRateLimiter {
	return newTokenBucketRateLimiter(gateway, RateLimiterConfig{
		TokenBucketKey:   consts.BuildContainerTokenBucket,
		MaxTokensKey:     consts.MaxTokensKeyBuildContainer,
		DefaultMaxTokens: consts.MaxConcurrentBuildContainer,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      consts.BuildContainerServiceName,
	})
}

func NewAlgoExecutionRateLimiter(gateway *redisinfra.Gateway) *TokenBucketRateLimiter {
	return newTokenBucketRateLimiter(gateway, RateLimiterConfig{
		TokenBucketKey:   consts.AlgoExecutionTokenBucket,
		MaxTokensKey:     consts.MaxTokensKeyAlgoExecution,
		DefaultMaxTokens: consts.MaxConcurrentAlgoExecution,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      consts.AlgoExecutionServiceName,
	})
}

// newTokenBucketRateLimiter creates a new token bucket rate limiter
func newTokenBucketRateLimiter(gateway *redisinfra.Gateway, cfg RateLimiterConfig) *TokenBucketRateLimiter {
	maxTokens := config.GetInt(cfg.MaxTokensKey)
	if maxTokens <= 0 {
		maxTokens = cfg.DefaultMaxTokens
	}

	waitTimeout := config.GetInt("rate_limiting.token_wait_timeout")
	if waitTimeout <= 0 {
		waitTimeout = cfg.DefaultTimeout
	}

	return &TokenBucketRateLimiter{
		bucketKey:   cfg.TokenBucketKey,
		store:       newTokenBucketStore(gateway, cfg.TokenBucketKey),
		maxTokens:   maxTokens,
		waitTimeout: time.Duration(waitTimeout) * time.Second,
		serviceName: cfg.ServiceName,
	}
}
