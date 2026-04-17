package consumer

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/repository"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type stateStore struct {
	db *gorm.DB
}

func newStateStore(db *gorm.DB) *stateStore {
	return &stateStore{db: db}
}

func (s *stateStore) updateExecutionState(executionID int, newState consts.ExecutionState) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		execution, err := repository.GetExecutionByID(tx, executionID)
		if err != nil {
			if errorsIsRecordNotFound(err) {
				return fmt.Errorf("%w: execution %d not found", consts.ErrNotFound, executionID)
			}
			return fmt.Errorf("execution %d not found: %w", executionID, err)
		}

		if execution.State != consts.ExecutionInitial {
			return fmt.Errorf("cannot change state of execution %d from %s to %s", executionID, consts.GetExecutionStateName(execution.State), consts.GetExecutionStateName(newState))
		}

		if err := repository.UpdateExecution(tx, executionID, map[string]any{
			"state": newState,
		}); err != nil {
			return fmt.Errorf("failed to update execution %d duration: %w", executionID, err)
		}

		return nil
	})
}

func (s *stateStore) updateInjectionState(injectionName string, newState consts.DatapackState) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByName(tx, injectionName, false)
		if err != nil {
			return fmt.Errorf("failed to get injection %s: %w", injectionName, err)
		}

		if err := repository.UpdateInjection(tx, injection.ID, map[string]any{
			"state": newState,
		}); err != nil {
			return fmt.Errorf("failed to update injection %s state: %w", injectionName, err)
		}

		return nil
	})
}

func (s *stateStore) updateInjectionTimestamp(injectionName string, startTime time.Time, endTime time.Time) (*dto.InjectionItem, error) {
	var updatedInjection *model.FaultInjection
	err := s.db.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByName(tx, injectionName, false)
		if err != nil {
			if errorsIsRecordNotFound(err) {
				return fmt.Errorf("injection %s not found", injectionName)
			}
			return fmt.Errorf("failed to get injection %s: %w", injectionName, err)
		}

		if err = repository.UpdateInjection(tx, injection.ID, map[string]any{
			"start_time": startTime,
			"end_time":   endTime,
		}); err != nil {
			return fmt.Errorf("update injection timestamps failed: %w", err)
		}

		reloadedInjection, err := repository.GetInjectionByID(tx, injection.ID)
		if err != nil {
			return fmt.Errorf("failed to reload injection %d after update: %w", injection.ID, err)
		}

		updatedInjection = reloadedInjection
		return nil
	})
	if err != nil {
		return nil, err
	}

	injectionItem := dto.NewInjectionItem(updatedInjection)
	return &injectionItem, nil
}

func errorsIsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
