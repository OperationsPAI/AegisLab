package evaluationmodule

import (
	"context"
	"fmt"

	"aegis/internalclient/orchestratorclient"
	executionmodule "aegis/module/execution"

	"go.uber.org/fx"
)

type executionQuerySource interface {
	ListEvaluationExecutionsByDatapack(context.Context, *executionmodule.EvaluationExecutionsByDatapackReq) ([]executionmodule.EvaluationExecutionItem, error)
	ListEvaluationExecutionsByDataset(context.Context, *executionmodule.EvaluationExecutionsByDatasetReq) ([]executionmodule.EvaluationExecutionItem, error)
}

type executionQueryAdapter struct {
	orchestrator  *orchestratorclient.Client
	local         *executionmodule.Service
	requireRemote bool
}

type executionQuerySourceParams struct {
	fx.In

	Orchestrator *orchestratorclient.Client `optional:"true"`
	Local        *executionmodule.Service   `optional:"true"`
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

func (a executionQueryAdapter) ListEvaluationExecutionsByDatapack(ctx context.Context, req *executionmodule.EvaluationExecutionsByDatapackReq) ([]executionmodule.EvaluationExecutionItem, error) {
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

func (a executionQueryAdapter) ListEvaluationExecutionsByDataset(ctx context.Context, req *executionmodule.EvaluationExecutionsByDatasetReq) ([]executionmodule.EvaluationExecutionItem, error) {
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
