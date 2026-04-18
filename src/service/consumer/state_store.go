package consumer

import (
	"context"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	executionmodule "aegis/module/execution"
	injectionmodule "aegis/module/injection"
)

type stateStore struct {
	execution ExecutionOwner
	injection InjectionOwner
}

func newStateStore(execution ExecutionOwner, injection InjectionOwner) *stateStore {
	return &stateStore{
		execution: execution,
		injection: injection,
	}
}

func (s *stateStore) updateExecutionState(ctx context.Context, executionID int, newState consts.ExecutionState) error {
	if s.execution == nil {
		return fmt.Errorf("execution owner service is nil")
	}
	return s.execution.UpdateExecutionState(ctx, &executionmodule.RuntimeUpdateExecutionStateReq{
		ExecutionID: executionID,
		State:       newState,
	})
}

func (s *stateStore) updateInjectionState(ctx context.Context, injectionName string, newState consts.DatapackState) error {
	if s.injection == nil {
		return fmt.Errorf("injection owner service is nil")
	}
	return s.injection.UpdateInjectionState(ctx, &injectionmodule.RuntimeUpdateInjectionStateReq{
		Name:  injectionName,
		State: newState,
	})
}

func (s *stateStore) updateInjectionTimestamp(ctx context.Context, injectionName string, startTime time.Time, endTime time.Time) (*dto.InjectionItem, error) {
	if s.injection == nil {
		return nil, fmt.Errorf("injection owner service is nil")
	}
	return s.injection.UpdateInjectionTimestamps(ctx, &injectionmodule.RuntimeUpdateInjectionTimestampReq{
		Name:      injectionName,
		StartTime: startTime,
		EndTime:   endTime,
	})
}
