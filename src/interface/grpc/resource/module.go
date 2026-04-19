package grpcresource

import "go.uber.org/fx"

var Module = fx.Module("grpc_resource",
	fx.Provide(
		newResourceServer,
		newLifecycle,
	),
	fx.Invoke(registerLifecycle),
)
