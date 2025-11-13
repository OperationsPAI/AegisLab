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

func dispatchTask(ctx context.Context, task *dto.UnifiedTask) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Task panic: %v\n%s", r, debug.Stack())
		}
	}()

	tracing.SetSpanAttribute(ctx, consts.TaskIDKey, task.TaskID)
	tracing.SetSpanAttribute(ctx, consts.TaskTypeKey, consts.GetTaskTypeName(task.Type))
	tracing.SetSpanAttribute(ctx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskPending))

	publishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
		TaskID:    task.TaskID,
		TaskType:  task.Type,
		EventName: consts.EventTaskStarted,
		Payload:   task,
	})

	var err error
	switch task.Type {
	case consts.TaskTypeBuildContainer:
		err = executeBuildContainer(ctx, task)
	case consts.TaskTypeRestartPedestal:
		err = executeRestartPedestal(ctx, task)
	case consts.TaskTypeFaultInjection:
		err = executeFaultInjection(ctx, task)
	case consts.TaskTypeBuildDatapack:
		err = executeBuildDatapack(ctx, task)
	case consts.TaskTypeRunAlgorithm:
		err = executeAlgorithm(ctx, task)
	case consts.TaskTypeCollectResult:
		err = executeCollectResult(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %d", task.Type)
	}

	if err != nil {
		return err
	}

	return nil
}
