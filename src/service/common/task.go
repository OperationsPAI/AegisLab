package common

import (
	"aegis/consts"
	"aegis/dto"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
func SubmitTaskWithDB(ctx context.Context, db *gorm.DB, redisGateway *redisinfra.Gateway, t *dto.UnifiedTask) error {
	if db == nil {
		return fmt.Errorf("task db is nil")
	}
	if redisGateway == nil {
		return fmt.Errorf("task redis gateway is nil")
	}

	if t.TraceID == "" {
		t.TraceID = uuid.NewString()
	}

	if t.TaskID == "" {
		t.TaskID = uuid.NewString()
	}

	if t.ParentTaskID != nil && t.State != consts.TaskRescheduled {
		parentLevel, err := getParentTaskLevelByID(db, *t.ParentTaskID)
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

	var trace *model.Trace
	var err error
	if t.ParentTaskID == nil && t.State != consts.TaskRescheduled {
		withAlgorithms := false
		leafNum := 1
		if t.Type == consts.TaskTypeRestartPedestal && t.Extra != nil {
			if val, ok := t.Extra[consts.TaskExtraInjectionAlgorithms]; ok {
				if algoNums, ok := val.(int); ok && algoNums > 0 {
					withAlgorithms = true
					leafNum = algoNums
				}
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if trace != nil {
			if err := upsertTrace(tx, trace); err != nil {
				return fmt.Errorf("failed to upsert trace to database: %w", err)
			}
		}

		if err := upsertTask(tx, task); err != nil {
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
		err = redisGateway.SubmitImmediateTask(ctx, taskData, t.TaskID)
	} else {
		err = redisGateway.SubmitDelayedTask(ctx, taskData, t.TaskID, t.ExecuteTime)
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

func getParentTaskLevelByID(db *gorm.DB, parentTaskID string) (int, error) {
	var result model.Task
	if err := db.Select("level").
		Where("id = ? AND status != ?", parentTaskID, consts.CommonDeleted).
		First(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to find parent task with id %s: %w", parentTaskID, err)
	}
	return result.Level, nil
}

func upsertTask(db *gorm.DB, task *model.Task) error {
	if err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"execute_time",
				"state",
				"updated_at",
			}),
		},
	).Create(task).Error; err != nil {
		return fmt.Errorf("failed to upsert task: %w", err)
	}
	return nil
}

func upsertTrace(db *gorm.DB, trace *model.Trace) error {
	if err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"last_event",
				"end_time",
				"state",
				"updated_at",
			}),
		},
	).Create(trace).Error; err != nil {
		return fmt.Errorf("failed to upsert trace: %w", err)
	}
	return nil
}
