package workerinterface

import (
	"context"

	buildkitinfra "aegis/infra/buildkit"
	etcdinfra "aegis/infra/etcd"
	helminfra "aegis/infra/helm"
	k8sinfra "aegis/infra/k8s"
	redisinfra "aegis/infra/redis"
	commonservice "aegis/service/common"
	"aegis/service/consumer"
	"aegis/service/initialization"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Module("worker",
	fx.Provide(newLifecycle),
	fx.Invoke(registerLifecycle),
)

type Params struct {
	fx.In

	DB             *gorm.DB
	RedisGateway   *redisinfra.Gateway
	BuildKit       *buildkitinfra.Gateway
	Helm           *helminfra.Gateway
	K8sGateway     *k8sinfra.Gateway
	Controller     *k8sinfra.Controller
	Etcd           *etcdinfra.Gateway
	Monitor        consumer.NamespaceMonitor
	RestartLimiter *consumer.TokenBucketRateLimiter `name:"restart_limiter"`
	BuildLimiter   *consumer.TokenBucketRateLimiter `name:"build_limiter"`
	AlgoLimiter    *consumer.TokenBucketRateLimiter `name:"algo_limiter"`
	BatchManager   *consumer.FaultBatchManager
	ExecutionOwner consumer.ExecutionOwner
	InjectionOwner consumer.InjectionOwner
}

type Lifecycle struct {
	params    Params
	StartFunc func(context.Context) error
	StopFunc  func()
}

func newLifecycle(params Params) *Lifecycle {
	return &Lifecycle{params: params}
}

func (r *Lifecycle) start(ctx context.Context) error {
	if r.StartFunc != nil {
		return r.StartFunc(ctx)
	}
	params := r.params
	if err := initialization.InitializeConsumer(
		ctx,
		params.DB,
		params.Controller,
		params.Monitor,
		params.RedisGateway,
		commonservice.NewConfigUpdateListener(ctx, params.DB, params.Etcd),
		params.RestartLimiter,
		params.BuildLimiter,
		params.AlgoLimiter,
	); err != nil {
		return err
	}
	if err := params.RedisGateway.InitConcurrencyLock(ctx); err != nil {
		return err
	}

	go consumer.StartScheduler(ctx, params.RedisGateway)
	go consumer.ConsumeTasks(ctx, consumer.RuntimeDeps{
		DB:                   params.DB,
		Monitor:              params.Monitor,
		RestartRateLimiter:   params.RestartLimiter,
		BuildRateLimiter:     params.BuildLimiter,
		AlgorithmRateLimiter: params.AlgoLimiter,
		RedisGateway:         params.RedisGateway,
		K8sGateway:           params.K8sGateway,
		BuildKitGateway:      params.BuildKit,
		HelmGateway:          params.Helm,
		FaultBatchManager:    params.BatchManager,
		ExecutionOwner:       params.ExecutionOwner,
		InjectionOwner:       params.InjectionOwner,
	})
	return nil
}

func (r *Lifecycle) stop() {
	if r.StopFunc != nil {
		r.StopFunc()
	}
}

func registerLifecycle(lc fx.Lifecycle, runner *Lifecycle) {
	var (
		workerCtx context.Context
		cancel    context.CancelFunc
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			workerCtx, cancel = context.WithCancel(context.WithoutCancel(ctx))
			return runner.start(workerCtx)
		},
		OnStop: func(ctx context.Context) error {
			if cancel != nil {
				cancel()
			}
			runner.stop()
			return nil
		},
	})
}
