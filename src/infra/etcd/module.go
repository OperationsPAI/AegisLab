package etcdinfra

import "go.uber.org/fx"

var Module = fx.Module("etcd",
	fx.Provide(NewGatewayWithLifecycle),
)
