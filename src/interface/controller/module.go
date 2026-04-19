package controller

import (
	"context"
	"log"
	"os"

	k8s "aegis/infra/k8s"
	redis "aegis/infra/redis"
	"aegis/service/consumer"

	"github.com/go-logr/stdr"
	"go.uber.org/fx"
	"gorm.io/gorm"
	k8slogger "sigs.k8s.io/controller-runtime/pkg/log"
)

var Module = fx.Module("controller",
	fx.Provide(newLifecycle),
	fx.Invoke(registerLifecycle),
)

type Params struct {
	fx.In

	Controller     *k8s.Controller
	K8sGateway     *k8s.Gateway
	RedisGateway   *redis.Gateway
	DB             *gorm.DB
	Monitor        consumer.NamespaceMonitor
	AlgoLimiter    *consumer.TokenBucketRateLimiter `name:"algo_limiter"`
	BatchManager   *consumer.FaultBatchManager
	ExecutionOwner consumer.ExecutionOwner
	InjectionOwner consumer.InjectionOwner
}

type Lifecycle struct {
	params   Params
	RunFunc  func(context.Context, context.CancelFunc) error
	StopFunc func()
}

func newLifecycle(params Params) *Lifecycle {
	return &Lifecycle{params: params}
}

func (r *Lifecycle) start(ctx context.Context, cancel context.CancelFunc) error {
	if r.RunFunc != nil {
		return r.RunFunc(ctx, cancel)
	}
	k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
	go r.params.Controller.Initialize(
		ctx,
		cancel,
		consumer.NewHandler(
			r.params.DB,
			r.params.Monitor,
			r.params.AlgoLimiter,
			r.params.K8sGateway,
			r.params.RedisGateway,
			r.params.BatchManager,
			r.params.ExecutionOwner,
			r.params.InjectionOwner,
		),
	)
	return nil
}

func (r *Lifecycle) stop() {
	if r.StopFunc != nil {
		r.StopFunc()
	}
}

func registerLifecycle(lc fx.Lifecycle, runner *Lifecycle) {
	var (
		controllerCtx context.Context
		cancel        context.CancelFunc
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			controllerCtx, cancel = context.WithCancel(context.WithoutCancel(ctx))
			return runner.start(controllerCtx, cancel)
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
