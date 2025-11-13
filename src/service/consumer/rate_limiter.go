package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aegis/client"
	"aegis/config"
	"aegis/consts"

	"github.com/redis/go-redis/v9"
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
	redisClient *redis.Client
	bucketKey   string
	maxTokens   int
	waitTimeout time.Duration
	serviceName string
}

// AcquireToken acquires a token
func (r *TokenBucketRateLimiter) AcquireToken(ctx context.Context, taskID, traceID string) (bool, error) {
	span := trace.SpanFromContext(ctx)

	script := redis.NewScript(`
		local bucket_key = KEYS[1]
		local max_tokens = tonumber(ARGV[1])
		local task_id = ARGV[2]
		local trace_id = ARGV[3]
		local expire_time = tonumber(ARGV[4])
		
		local current_tokens = redis.call('SCARD', bucket_key)
		
		if current_tokens < max_tokens then
			redis.call('SADD', bucket_key, task_id)
			redis.call('EXPIRE', bucket_key, expire_time)
			return 1
		else
			return 0
		end
	`)

	expireTime := 10 * 60

	result, err := script.Run(ctx, r.redisClient, []string{r.bucketKey},
		r.maxTokens, taskID, traceID, expireTime).Result()
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to acquire token: %v", err)
	}

	acquired := result.(int64) == 1
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

	result, err := r.redisClient.SRem(ctx, r.bucketKey, taskID).Result()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to release token: %v", err)
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

	timeoutCtx, cancel := context.WithTimeout(ctx, r.waitTimeout)
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
				"timeout":    r.waitTimeout,
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

var (
	restartPedestalRateLimiter *TokenBucketRateLimiter
	buildContainerRateLimiter  *TokenBucketRateLimiter
	algoExecutionRateLimiter   *TokenBucketRateLimiter
	rateLimiterOnce            sync.Once
)

// GetRestartPedestalRateLimiter returns the singleton restart pedestal rate limiter
func GetRestartPedestalRateLimiter() *TokenBucketRateLimiter {
	rateLimiterOnce.Do(initRateLimiters)
	return restartPedestalRateLimiter
}

// GetBuildContainerRateLimiter returns the singleton build container rate limiter
func GetBuildContainerRateLimiter() *TokenBucketRateLimiter {
	rateLimiterOnce.Do(initRateLimiters)
	return buildContainerRateLimiter
}

// GetAlgoExecutionRateLimiter returns the singleton algorithm execution rate limiter
func GetAlgoExecutionRateLimiter() *TokenBucketRateLimiter {
	rateLimiterOnce.Do(initRateLimiters)
	return algoExecutionRateLimiter
}

// initRateLimiters initializes all rate limiters
func initRateLimiters() {
	restartPedestalRateLimiter = newTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   consts.RestartPedestalTokenBucket,
		MaxTokensKey:     consts.MaxTokensKeyRestartPedestal,
		DefaultMaxTokens: consts.MaxConcurrentRestartPedestal,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      consts.RestartPedestalServiceName,
	})

	buildContainerRateLimiter = newTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   consts.BuildContainerTokenBucket,
		MaxTokensKey:     consts.MaxTokensKeyBuildContainer,
		DefaultMaxTokens: consts.MaxConcurrentBuildContainer,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      consts.BuildContainerServiceName,
	})

	algoExecutionRateLimiter = newTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   consts.AlgoExecutionTokenBucket,
		MaxTokensKey:     consts.MaxTokensKeyAlgoExecution,
		DefaultMaxTokens: consts.MaxConcurrentAlgoExecution,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      consts.AlgoExecutionServiceName,
	})
}

// newTokenBucketRateLimiter creates a new token bucket rate limiter
func newTokenBucketRateLimiter(cfg RateLimiterConfig) *TokenBucketRateLimiter {
	maxTokens := config.GetInt(cfg.MaxTokensKey)
	if maxTokens <= 0 {
		maxTokens = cfg.DefaultMaxTokens
	}

	waitTimeout := config.GetInt("rate_limiting.token_wait_timeout")
	if waitTimeout <= 0 {
		waitTimeout = cfg.DefaultTimeout
	}

	return &TokenBucketRateLimiter{
		redisClient: client.GetRedisClient(),
		bucketKey:   cfg.TokenBucketKey,
		maxTokens:   maxTokens,
		waitTimeout: time.Duration(waitTimeout) * time.Second,
		serviceName: cfg.ServiceName,
	}
}
