package consumer

import (
	"context"
	"fmt"

	"aegis/dto"
	"aegis/internalclient/orchestratorclient"
	executionmodule "aegis/module/execution"
	injectionmodule "aegis/module/injection"

	"go.uber.org/fx"
)

// ExecutionOwner captures the execution owner operations used by runtime code.
type ExecutionOwner interface {
	CreateExecution(context.Context, *executionmodule.RuntimeCreateExecutionReq) (int, error)
	GetExecution(context.Context, int) (*executionmodule.ExecutionDetailResp, error)
	UpdateExecutionState(context.Context, *executionmodule.RuntimeUpdateExecutionStateReq) error
}

// InjectionOwner captures the injection owner operations used by runtime code.
type InjectionOwner interface {
	CreateInjection(context.Context, *injectionmodule.RuntimeCreateInjectionReq) (*dto.InjectionItem, error)
	UpdateInjectionState(context.Context, *injectionmodule.RuntimeUpdateInjectionStateReq) error
	UpdateInjectionTimestamps(context.Context, *injectionmodule.RuntimeUpdateInjectionTimestampReq) (*dto.InjectionItem, error)
}

type executionOwnerAdapter struct {
	orchestrator  *orchestratorclient.Client
	local         *executionmodule.Service
	requireRemote bool
}

type executionOwnerParams struct {
	fx.In

	Orchestrator *orchestratorclient.Client
	Local        *executionmodule.Service `optional:"true"`
}

type injectionOwnerParams struct {
	fx.In

	Orchestrator *orchestratorclient.Client
	Local        *injectionmodule.Service `optional:"true"`
}

func NewExecutionOwner(params executionOwnerParams) ExecutionOwner {
	return executionOwnerAdapter{
		orchestrator:  params.Orchestrator,
		local:         params.Local,
		requireRemote: false,
	}
}

func NewInjectionOwner(params injectionOwnerParams) InjectionOwner {
	return injectionOwnerAdapter{
		orchestrator:  params.Orchestrator,
		local:         params.Local,
		requireRemote: false,
	}
}

func newRemoteExecutionOwner(params executionOwnerParams) ExecutionOwner {
	return executionOwnerAdapter{
		orchestrator:  params.Orchestrator,
		local:         params.Local,
		requireRemote: true,
	}
}

func newRemoteInjectionOwner(params injectionOwnerParams) InjectionOwner {
	return injectionOwnerAdapter{
		orchestrator:  params.Orchestrator,
		local:         params.Local,
		requireRemote: true,
	}
}

func (a executionOwnerAdapter) CreateExecution(ctx context.Context, req *executionmodule.RuntimeCreateExecutionReq) (int, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.CreateExecution(ctx, req)
	}
	if a.requireRemote {
		return 0, fmt.Errorf("orchestrator-service owner is not configured")
	}
	if a.local == nil {
		return 0, fmt.Errorf("missing execution owner service")
	}
	return a.local.CreateExecutionRecord(ctx, req)
}

func (a executionOwnerAdapter) GetExecution(ctx context.Context, executionID int) (*executionmodule.ExecutionDetailResp, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.GetExecution(ctx, executionID)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("orchestrator-service owner is not configured")
	}
	if a.local == nil {
		return nil, fmt.Errorf("missing execution owner service")
	}
	return a.local.GetExecution(ctx, executionID)
}

func (a executionOwnerAdapter) UpdateExecutionState(ctx context.Context, req *executionmodule.RuntimeUpdateExecutionStateReq) error {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.UpdateExecutionState(ctx, req)
	}
	if a.requireRemote {
		return fmt.Errorf("orchestrator-service owner is not configured")
	}
	if a.local == nil {
		return fmt.Errorf("missing execution owner service")
	}
	return a.local.UpdateExecutionState(ctx, req)
}

type injectionOwnerAdapter struct {
	orchestrator  *orchestratorclient.Client
	local         *injectionmodule.Service
	requireRemote bool
}

func (a injectionOwnerAdapter) CreateInjection(ctx context.Context, req *injectionmodule.RuntimeCreateInjectionReq) (*dto.InjectionItem, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.CreateInjection(ctx, req)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("orchestrator-service owner is not configured")
	}
	if a.local == nil {
		return nil, fmt.Errorf("missing injection owner service")
	}
	return a.local.CreateInjectionRecord(ctx, req)
}

func (a injectionOwnerAdapter) UpdateInjectionState(ctx context.Context, req *injectionmodule.RuntimeUpdateInjectionStateReq) error {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.UpdateInjectionState(ctx, req)
	}
	if a.requireRemote {
		return fmt.Errorf("orchestrator-service owner is not configured")
	}
	if a.local == nil {
		return fmt.Errorf("missing injection owner service")
	}
	return a.local.UpdateInjectionState(ctx, req)
}

func (a injectionOwnerAdapter) UpdateInjectionTimestamps(ctx context.Context, req *injectionmodule.RuntimeUpdateInjectionTimestampReq) (*dto.InjectionItem, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.UpdateInjectionTimestamps(ctx, req)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("orchestrator-service owner is not configured")
	}
	if a.local == nil {
		return nil, fmt.Errorf("missing injection owner service")
	}
	return a.local.UpdateInjectionTimestamps(ctx, req)
}

// RemoteOwnerOptions forces the dedicated runtime-worker-service path to use orchestrator RPC only.
func RemoteOwnerOptions() fx.Option {
	return fx.Options(
		fx.Decorate(newRemoteExecutionOwner),
		fx.Decorate(newRemoteInjectionOwner),
	)
}
