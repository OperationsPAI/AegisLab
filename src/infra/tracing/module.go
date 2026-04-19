package tracing

import (
	"context"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

var Module = fx.Module("tracing",
	fx.Provide(NewTraceProvider),
)

func NewTraceProvider(lc fx.Lifecycle) *trace.TracerProvider {
	provider, err := NewProvider()
	if err != nil {
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ShutdownProvider(ctx, provider)
			return nil
		},
	})

	return provider
}
