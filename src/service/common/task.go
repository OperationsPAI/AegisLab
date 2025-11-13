package common

import (
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
func SubmitTask(ctx context.Context, task *dto.UnifiedTask) error {
	if task.TaskID == "" {
		task.TaskID = uuid.NewString()
	}

	if task.TraceID == "" {
		task.TraceID = uuid.NewString()
	}

	t, err := task.ConvertToTask()
	if err != nil {
		return fmt.Errorf("failed to convert to task: %w", err)
	}

	if err := repository.UpsertTask(database.DB, t); err != nil {
		return fmt.Errorf("failed to upsert task to database: %w", err)
	}

	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	if task.Immediate {
		err = repository.SubmitImmediateTask(ctx, taskData, task.TaskID)
	} else {
		if err = calculateExecuteTime(task); err != nil {
			return err
		}
		err = repository.SubmitDelayedTask(ctx, taskData, task.TaskID, task.ExecuteTime)
	}

	if err != nil {
		return err
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
