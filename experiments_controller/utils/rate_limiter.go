package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	TokenBucketKey   string
	MaxTokensKey     string
	DefaultMaxTokens int
	DefaultTimeout   int
	ServiceName      string
}

// TokenBucketRateLimiter 基于 Redis 的令牌桶限流器
type TokenBucketRateLimiter struct {
	redisClient *redis.Client
	bucketKey   string
	maxTokens   int
	waitTimeout time.Duration
	serviceName string
}

// NewTokenBucketRateLimiter 创建新的令牌桶限流器
func NewTokenBucketRateLimiter(cfg RateLimiterConfig) *TokenBucketRateLimiter {
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

// AcquireToken 获取令牌
func (r *TokenBucketRateLimiter) AcquireToken(ctx context.Context, taskID, traceID string) (bool, error) {
	span := trace.SpanFromContext(ctx)

	script := redis.NewScript(`
		local bucket_key = KEYS[1]
		local max_tokens = tonumber(ARGV[1])
		local task_id = ARGV[2]
		local trace_id = ARGV[3]
		local expire_time = tonumber(ARGV[4])
		
		-- 获取当前令牌数量
		local current_tokens = redis.call('SCARD', bucket_key)
		
		if current_tokens < max_tokens then
			-- 有可用令牌，添加任务ID到集合中
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

// ReleaseToken 释放令牌
func (r *TokenBucketRateLimiter) ReleaseToken(ctx context.Context, taskID, traceID string) error {
	span := trace.SpanFromContext(ctx)

	// 从集合中移除任务ID
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

// WaitForToken 等待获取令牌，如果超时则返回 false
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

// NewRestartServiceRateLimiter 创建重启服务限流器
func NewRestartServiceRateLimiter() *TokenBucketRateLimiter {
	return NewTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   consts.RestartServiceTokenBucket,
		MaxTokensKey:     "rate_limiting.max_concurrent_restarts",
		DefaultMaxTokens: consts.MaxConcurrentRestarts,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      "restart_service",
	})
}

// NewBuildContainerRateLimiter 创建构建容器限流器
func NewBuildContainerRateLimiter() *TokenBucketRateLimiter {
	return NewTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   consts.BuildContainerTokenBucket,
		MaxTokensKey:     "rate_limiting.max_concurrent_builds",
		DefaultMaxTokens: consts.MaxConcurrentBuilds,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      "build_container",
	})
}

// NewAlgoExecutionRateLimiter 创建算法执行限流器
func NewAlgoExecutionRateLimiter() *TokenBucketRateLimiter {
	return NewTokenBucketRateLimiter(RateLimiterConfig{
		TokenBucketKey:   consts.AlgoExecutionTokenBucket,
		MaxTokensKey:     "rate_limiting.max_concurrent_algo_execution",
		DefaultMaxTokens: consts.MaxConcurrentAlgoExecution,
		DefaultTimeout:   consts.TokenWaitTimeout,
		ServiceName:      "algo_execution",
	})
}
