package consumer

import (
	"context"
	"fmt"
	"time"

	"aegis/consts"
	buildkitinfra "aegis/infra/buildkit"
	helminfra "aegis/infra/helm"
	k8sinfra "aegis/infra/k8s"
	redisinfra "aegis/infra/redis"

	"gorm.io/gorm"
)

const (
	RuntimeServiceName = "runtime-worker-service"
	healthCheckTimeout = 2 * time.Second
	runtimeModeWorker  = "runtime-worker"
)

type DependencyStatus struct {
	Available bool
	Healthy   bool
	Error     string
}

type RuntimeStatusSnapshot struct {
	ServiceName   string
	Mode          string
	AppID         string
	StartedAt     time.Time
	UptimeSeconds int64
	DB            DependencyStatus
	Redis         DependencyStatus
	K8s           DependencyStatus
	BuildKit      DependencyStatus
	Helm          DependencyStatus
}

type RuntimeSnapshotService struct {
	db        *gorm.DB
	redis     *redisinfra.Gateway
	k8s       *k8sinfra.Gateway
	buildkit  *buildkitinfra.Gateway
	helm      *helminfra.Gateway
	restart   *TokenBucketRateLimiter
	build     *TokenBucketRateLimiter
	algorithm *TokenBucketRateLimiter
}

func NewRuntimeSnapshotService(
	db *gorm.DB,
	redis *redisinfra.Gateway,
	k8s *k8sinfra.Gateway,
	buildkit *buildkitinfra.Gateway,
	helm *helminfra.Gateway,
	restart *TokenBucketRateLimiter,
	build *TokenBucketRateLimiter,
	algorithm *TokenBucketRateLimiter,
) *RuntimeSnapshotService {
	return &RuntimeSnapshotService{
		db:        db,
		redis:     redis,
		k8s:       k8s,
		buildkit:  buildkit,
		helm:      helm,
		restart:   restart,
		build:     build,
		algorithm: algorithm,
	}
}

func (s *RuntimeSnapshotService) RuntimeStatus(ctx context.Context) RuntimeStatusSnapshot {
	startedAt := time.Now()
	if consts.InitialTime != nil {
		startedAt = *consts.InitialTime
	}

	return RuntimeStatusSnapshot{
		ServiceName:   RuntimeServiceName,
		Mode:          runtimeModeWorker,
		AppID:         consts.AppID,
		StartedAt:     startedAt,
		UptimeSeconds: int64(time.Since(startedAt).Seconds()),
		DB:            s.dbStatus(ctx),
		Redis:         s.redisStatus(ctx),
		K8s:           s.k8sStatus(ctx),
		BuildKit:      s.buildkitStatus(ctx),
		Helm:          s.helmStatus(),
	}
}

func (s *RuntimeSnapshotService) QueueStatus(ctx context.Context) (redisinfra.TaskQueueStats, error) {
	if s.redis == nil {
		return redisinfra.TaskQueueStats{}, fmt.Errorf("redis gateway is nil")
	}
	return s.redis.GetTaskQueueStats(ctx)
}

func (s *RuntimeSnapshotService) LimiterStatus(ctx context.Context) []RateLimiterSnapshot {
	limiters := make([]RateLimiterSnapshot, 0, 3)
	for _, limiter := range []*TokenBucketRateLimiter{s.restart, s.build, s.algorithm} {
		if limiter == nil {
			continue
		}
		limiters = append(limiters, limiter.Snapshot(ctx))
	}
	return limiters
}

func (s *RuntimeSnapshotService) dbStatus(ctx context.Context) DependencyStatus {
	if s.db == nil {
		return DependencyStatus{Available: false}
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		return DependencyStatus{Available: true, Healthy: false, Error: err.Error()}
	}

	checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), healthCheckTimeout)
	defer cancel()
	if err := sqlDB.PingContext(checkCtx); err != nil {
		return DependencyStatus{Available: true, Healthy: false, Error: err.Error()}
	}
	return DependencyStatus{Available: true, Healthy: true}
}

func (s *RuntimeSnapshotService) redisStatus(ctx context.Context) DependencyStatus {
	if s.redis == nil {
		return DependencyStatus{Available: false}
	}

	checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), healthCheckTimeout)
	defer cancel()
	if err := s.redis.Ping(checkCtx); err != nil {
		return DependencyStatus{Available: true, Healthy: false, Error: err.Error()}
	}
	return DependencyStatus{Available: true, Healthy: true}
}

func (s *RuntimeSnapshotService) k8sStatus(ctx context.Context) DependencyStatus {
	if s.k8s == nil {
		return DependencyStatus{Available: false}
	}

	checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), healthCheckTimeout)
	defer cancel()
	if err := s.k8s.CheckHealth(checkCtx); err != nil {
		return DependencyStatus{Available: true, Healthy: false, Error: err.Error()}
	}
	return DependencyStatus{Available: true, Healthy: true}
}

func (s *RuntimeSnapshotService) buildkitStatus(ctx context.Context) DependencyStatus {
	if s.buildkit == nil {
		return DependencyStatus{Available: false}
	}

	checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), healthCheckTimeout)
	defer cancel()
	if err := s.buildkit.CheckHealth(checkCtx, healthCheckTimeout); err != nil {
		return DependencyStatus{Available: true, Healthy: false, Error: err.Error()}
	}
	return DependencyStatus{Available: true, Healthy: true}
}

func (s *RuntimeSnapshotService) helmStatus() DependencyStatus {
	if s.helm == nil {
		return DependencyStatus{Available: false}
	}
	return DependencyStatus{Available: true, Healthy: true}
}
