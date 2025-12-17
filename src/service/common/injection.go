package common

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/utils"
	"context"
	"fmt"
	"time"
)

// ProduceFaultInjectionTasks produces fault injection tasks into Redis based on the request specifications
func ProduceFaultInjectionTasks(ctx context.Context, task *dto.UnifiedTask, injectTime time.Time, payload map[string]any) error {
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
	err := SubmitTask(ctx, newTask)
	if err != nil {
		return fmt.Errorf("failed to submit fault injection task: %w", err)
	}
	return nil
}
