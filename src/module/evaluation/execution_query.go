package evaluation

import (
	"context"
	"fmt"

	"aegis/internalclient/orchestratorclient"
	execution "aegis/module/execution"

	"go.uber.org/fx"
)

type executionQuerySource interface {
	ListEvaluationExecutionsByDatapack(context.Context, *execution.EvaluationExecutionsByDatapackReq) ([]execution.EvaluationExecutionItem, error)
	ListEvaluationExecutionsByDataset(context.Context, *execution.EvaluationExecutionsByDatasetReq) ([]execution.EvaluationExecutionItem, error)
}

type executionQueryAdapter struct {
	orchestrator  *orchestratorclient.Client
	local         *execution.Service
	requireRemote bool
}

type executionQuerySourceParams struct {
	fx.In

	Orchestrator *orchestratorclient.Client `optional:"true"`
	Local        *execution.Service         `optional:"true"`
}

func newExecutionQuerySource(params executionQuerySourceParams) executionQuerySource {
	return executionQueryAdapter{
		orchestrator:  params.Orchestrator,
		local:         params.Local,
		requireRemote: false,
	}
}

func newRemoteExecutionQuerySource(params executionQuerySourceParams) executionQuerySource {
	return executionQueryAdapter{
		orchestrator:  params.Orchestrator,
		local:         params.Local,
		requireRemote: true,
	}
}

func (a executionQueryAdapter) ListEvaluationExecutionsByDatapack(ctx context.Context, req *execution.EvaluationExecutionsByDatapackReq) ([]execution.EvaluationExecutionItem, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.ListEvaluationExecutionsByDatapack(ctx, req)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("orchestrator-service query source is not configured")
	}
	if a.local == nil {
		return nil, fmt.Errorf("evaluation execution query source is not configured")
	}
	return a.local.ListEvaluationExecutionsByDatapack(ctx, req)
}

func (a executionQueryAdapter) ListEvaluationExecutionsByDataset(ctx context.Context, req *execution.EvaluationExecutionsByDatasetReq) ([]execution.EvaluationExecutionItem, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.ListEvaluationExecutionsByDataset(ctx, req)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("orchestrator-service query source is not configured")
	}
	if a.local == nil {
		return nil, fmt.Errorf("evaluation execution query source is not configured")
	}
	return a.local.ListEvaluationExecutionsByDataset(ctx, req)
}

// RemoteQueryOption forces the dedicated resource-service path to use orchestrator RPC only.
func RemoteQueryOption() fx.Option {
	return fx.Decorate(newRemoteExecutionQuerySource)
}
