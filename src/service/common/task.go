package common

import (
	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// cronNextTime calculates the next execution time from a cron expression
func CronNextTime(expr string) (time.Time, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(time.Now()), nil
}

// SubmitTask submits a task for execution, either immediate or delayed
//
// Task Context Hierarchy:
//  1. If GroupCarrier is not nil: task is an initial task that spawns several traces
//     1.2. If TraceCarrier is nil, create a new one
//  2. If TraceCarrier is not nil: task is within a task trace
//
// When calling SubmitTask:
// - For initial task: fill in the GroupCarrier (parent's parent)
// - For subsequent task: fill in the TraceCarrier (parent)
// - The context itself is the youngest span
//
// Hierarchy example:
//
//	Group -> Trace -> Task 1
//	               -> Task 2
//	               -> Task 3
//	               -> Task 4
//	               -> Task 5
func SubmitTask(ctx context.Context, t *dto.UnifiedTask) error {
	if t.TraceID == "" {
		t.TraceID = uuid.NewString()
	}

	if t.TaskID == "" {
		t.TaskID = uuid.NewString()
	}

	if t.ParentTaskID != nil && t.State != consts.TaskRescheduled {
		parentLevel, err := repository.GetParentTaskLevelByID(database.DB, *t.ParentTaskID)
		if err != nil {
			return fmt.Errorf("failed to get parent task level: %w", err)
		}
		t.Level = parentLevel + 1
	}

	if !t.Immediate {
		if err := calculateExecuteTime(t); err != nil {
			return fmt.Errorf("failed to calculate execute time: %w", err)
		}
	}

	var trace *database.Trace
	var err error
	if t.ParentTaskID == nil && t.State != consts.TaskRescheduled {
		withAlgorithms := false
		leafNum := 1
		if t.Type == consts.TaskTypeRestartPedestal {
			var algorithms []dto.ContainerVersionItem
			if err := client.GetHashField(ctx, consts.InjectionAlgorithmsKey, t.GroupID, &algorithms); err != nil {
				return fmt.Errorf("failed to get algorithms from redis: %w", err)
			}

			if len(algorithms) > 0 {
				withAlgorithms = true
				leafNum = len(algorithms)
			}
		}

		trace, err = t.ConvertToTrace(withAlgorithms, leafNum)
		if err != nil {
			return fmt.Errorf("failed to convert to trace: %w", err)
		}
	}

	task, err := t.ConvertToTask()
	if err != nil {
		return fmt.Errorf("failed to convert to task: %w", err)
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if trace != nil {
			if err := repository.UpsertTrace(tx, trace); err != nil {
				return fmt.Errorf("failed to upsert trace to database: %w", err)
			}
		}

		if err := repository.UpsertTask(tx, task); err != nil {
			return fmt.Errorf("failed to upsert task to database: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	taskData, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	if t.Immediate {
		err = repository.SubmitImmediateTask(ctx, taskData, t.TaskID)
	} else {
		err = repository.SubmitDelayedTask(ctx, taskData, t.TaskID, t.ExecuteTime)
	}

	if err != nil {
		return fmt.Errorf("failed to submit task to queue (task saved in DB): %w", err)
	}

	return nil
}

// calculateExecuteTime calculates the execution time for a task
func calculateExecuteTime(task *dto.UnifiedTask) error {
	if task.Type == consts.TaskTypeCronJob {
		next, err := CronNextTime(task.CronExpr)
		if err != nil {
			return err
		}

		task.ExecuteTime = next.Unix()
	}
	return nil
}
