package receiver

import (
	"context"

	"aegis/config"
	redis "aegis/infra/redis"
	"aegis/service/logreceiver"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var Module = fx.Module("receiver",
	fx.Provide(newLifecycle),
	fx.Invoke(registerLifecycle),
)

type Lifecycle struct {
	receiver  *logreceiver.OTLPLogReceiver
	StartFunc func(context.Context) error
	StopFunc  func()
}

func newLifecycle(redisGateway *redis.Gateway) *Lifecycle {
	otlpPort := config.GetInt("otlp_receiver.port")
	if otlpPort == 0 {
		otlpPort = logreceiver.DefaultPort
	}
	return &Lifecycle{
		receiver: logreceiver.NewOTLPLogReceiver(otlpPort, 0, redisGateway),
	}
}

func (r *Lifecycle) start(ctx context.Context) error {
	if r.StartFunc != nil {
		return r.StartFunc(ctx)
	}
	go func() {
		if err := r.receiver.Start(ctx); err != nil {
			logrus.Errorf("OTLP log receiver error: %v", err)
		}
	}()
	return nil
}

func (r *Lifecycle) stop() {
	if r.StopFunc != nil {
		r.StopFunc()
		return
	}
	if r.receiver != nil {
		r.receiver.Shutdown()
	}
}

func registerLifecycle(lc fx.Lifecycle, runner *Lifecycle) {
	var (
		receiverCtx context.Context
		cancel      context.CancelFunc
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			receiverCtx, cancel = context.WithCancel(context.WithoutCancel(ctx))
			return runner.start(receiverCtx)
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
