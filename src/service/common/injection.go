package common

import (
	"aegis/consts"
	"aegis/dto"
	redis "aegis/infra/redis"
	"aegis/utils"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func ProduceFaultInjectionTasksWithDB(ctx context.Context, db *gorm.DB, redisGateway *redis.Gateway, task *dto.UnifiedTask, injectTime time.Time, payload map[string]any) error {
	newTask := &dto.UnifiedTask{
		Type:         consts.TaskTypeFaultInjection,
		Immediate:    false,
		ExecuteTime:  injectTime.Unix(),
		Payload:      payload,
		ParentTaskID: utils.StringPtr(task.TaskID),
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		State:        consts.TaskPending,
		TraceCarrier: task.TraceCarrier,
		GroupCarrier: task.GroupCarrier,
	}
	err := SubmitTaskWithDB(ctx, db, redisGateway, newTask)
	if err != nil {
		return fmt.Errorf("failed to submit fault injection task: %w", err)
	}
	return nil
}
