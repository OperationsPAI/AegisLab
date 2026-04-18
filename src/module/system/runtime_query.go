package systemmodule

import (
	"context"
	"fmt"

	"aegis/internalclient/runtimeclient"
	systemmetricmodule "aegis/module/systemmetric"
	taskmodule "aegis/module/task"

	"go.uber.org/fx"
)

type runtimeQuerySource interface {
	ListNamespaceLocks(context.Context) (*ListNamespaceLockResp, error)
	ListQueuedTasks(context.Context) (*taskmodule.QueuedTasksResp, error)
}

type runtimeQueryAdapter struct {
	runtime       *runtimeclient.Client
	local         *systemmetricmodule.Service
	requireRemote bool
}

type runtimeQuerySourceParams struct {
	fx.In

	Runtime *runtimeclient.Client `optional:"true"`
	Local   *systemmetricmodule.Service
}

func newRuntimeQuerySource(params runtimeQuerySourceParams) runtimeQuerySource {
	return runtimeQueryAdapter{
		runtime:       params.Runtime,
		local:         params.Local,
		requireRemote: false,
	}
}

func newRemoteRuntimeQuerySource(params runtimeQuerySourceParams) runtimeQuerySource {
	return runtimeQueryAdapter{
		runtime:       params.Runtime,
		local:         params.Local,
		requireRemote: true,
	}
}

func (a runtimeQueryAdapter) ListNamespaceLocks(ctx context.Context) (*ListNamespaceLockResp, error) {
	if a.runtime != nil && a.runtime.Enabled() {
		return a.runtime.GetNamespaceLocks(ctx)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("runtime-worker-service query source is not configured")
	}
	return a.local.ListNamespaceLocks(ctx)
}

func (a runtimeQueryAdapter) ListQueuedTasks(ctx context.Context) (*taskmodule.QueuedTasksResp, error) {
	if a.runtime != nil && a.runtime.Enabled() {
		return a.runtime.GetQueuedTasks(ctx)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("runtime-worker-service query source is not configured")
	}
	return a.local.ListQueuedTasks(ctx)
}

// RemoteRuntimeQueryOption forces the dedicated system-service path to use runtime RPC only.
func RemoteRuntimeQueryOption() fx.Option {
	return fx.Decorate(newRemoteRuntimeQuerySource)
}
