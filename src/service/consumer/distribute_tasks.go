package consumer

import (
	"context"
	"fmt"
	"runtime/debug"

	"aegis/consts"
	"aegis/dto"
	"aegis/tracing"

	"github.com/sirupsen/logrus"
)

func dispatchTask(ctx context.Context, task *dto.UnifiedTask, deps RuntimeDeps) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Task panic: %v\n%s", r, debug.Stack())
		}
	}()

	tracing.SetSpanAttribute(ctx, consts.TaskIDKey, task.TaskID)
	tracing.SetSpanAttribute(ctx, consts.TaskTypeKey, consts.GetTaskTypeName(task.Type))
	tracing.SetSpanAttribute(ctx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskPending))

	publishEvent(deps.RedisGateway, ctx, fmt.Sprintf(consts.StreamTraceLogKey, task.TraceID), dto.TraceStreamEvent{
		TaskID:    task.TaskID,
		TaskType:  task.Type,
		EventName: consts.EventTaskStarted,
		Payload:   task,
	})

	var err error
	switch task.Type {
	case consts.TaskTypeBuildContainer:
		err = executeBuildContainer(ctx, task, deps)
	case consts.TaskTypeRestartPedestal:
		err = executeRestartPedestal(ctx, task, deps)
	case consts.TaskTypeFaultInjection:
		err = executeFaultInjection(ctx, task, deps)
	case consts.TaskTypeBuildDatapack:
		err = executeBuildDatapackWithDeps(ctx, task, deps)
	case consts.TaskTypeRunAlgorithm:
		err = executeAlgorithm(ctx, task, deps)
	case consts.TaskTypeCollectResult:
		err = executeCollectResult(ctx, task, deps)
	default:
		err = fmt.Errorf("unknown task type: %d", task.Type)
	}

	if err != nil {
		return err
	}

	return nil
}
