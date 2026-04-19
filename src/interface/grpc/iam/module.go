package grpciam

import "go.uber.org/fx"

var Module = fx.Module("grpc_iam",
	fx.Provide(
		newIAMServer,
		newLifecycle,
	),
	fx.Invoke(registerLifecycle),
)
