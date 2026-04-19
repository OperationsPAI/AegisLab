package grpcruntime

import "go.uber.org/fx"

var Module = fx.Module("grpc_runtime",
	fx.Provide(
		newRuntimeServer,
		newLifecycle,
	),
	fx.Invoke(registerLifecycle),
)
