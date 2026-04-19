package systemmetric

import (
	"context"
	"runtime"
	"time"

	"go.uber.org/fx"
)

func RegisterMetricsCollector(lifecycle fx.Lifecycle, service *Service) {
	var cancel context.CancelFunc

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			collectorCtx, collectorCancel := context.WithCancel(context.WithoutCancel(ctx))
			cancel = collectorCancel
			go func() {
				ticker := time.NewTicker(time.Minute)
				defer ticker.Stop()

				for {
					select {
					case <-collectorCtx.Done():
						return
					case <-ticker.C:
					}

					if err := service.StoreSystemMetrics(collectorCtx); err != nil {
						// Keep the collector alive even if a single write fails.
						runtime.Gosched()
					}
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			if cancel != nil {
				cancel()
			}
			return nil
		},
	})
}
