package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/tracing"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func dispatchTask(ctx context.Context, task *UnifiedTask) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Task panic: %v\n%s", r, debug.Stack())
		}
	}()

	tracing.SetSpanAttribute(ctx, consts.TaskIDKey, task.TaskID)
	tracing.SetSpanAttribute(ctx, consts.TaskTypeKey, string(task.Type))
	tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusRunning))

	client.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), client.StreamEvent{
		TaskID:    task.TaskID,
		TaskType:  task.Type,
		EventName: consts.EventTaskStarted,
		Payload:   task,
	})

	var err error
	switch task.Type {
	case consts.TaskTypeRestartService:
		err = executeRestartService(ctx, task)
	case consts.TaskTypeFaultInjection:
		err = executeFaultInjection(ctx, task)
	case consts.TaskTypeRunAlgorithm:
		err = executeAlgorithm(ctx, task)
	case consts.TaskTypeBuildImages:
		err = executeBuildImages(ctx, task)
	case consts.TaskTypeBuildDataset:
		err = executeBuildDataset(ctx, task)
	case consts.TaskTypeCollectResult:
		err = executeCollectResult(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	if err != nil {
		return err
	}

	return nil
}

func getAnnotations(ctx context.Context, task *UnifiedTask) (map[string]string, error) {
	taskCarrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, taskCarrier)
	taskCarrierBytes, err := json.Marshal(taskCarrier)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mapcarrier of task context")
	}

	traceCarrier := task.TraceCarrier
	traceCarrierBytes, err := json.Marshal(traceCarrier)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mapcarrier of trace context")
	}

	return map[string]string{
		consts.TaskCarrier:  string(taskCarrierBytes),
		consts.TraceCarrier: string(traceCarrierBytes),
	}, nil
}
