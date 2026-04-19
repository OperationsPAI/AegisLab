package grpcsystem

import "go.uber.org/fx"

var Module = fx.Module("grpc_system",
	fx.Provide(
		newSystemServer,
		newLifecycle,
	),
	fx.Invoke(registerLifecycle),
)
