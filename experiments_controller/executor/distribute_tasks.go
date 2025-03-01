package executor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func dispatchTask(ctx context.Context, task *UnifiedTask) error {
	logrus.Infof("Executing task ID: [%s]", task.TaskID)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Task panic: %v\n%s", r, debug.Stack())
		}
	}()

	var err error
	switch task.Type {
	case TaskTypeFaultInjection:
		logrus.Debug("executeFaultInjection")
		err = executeFaultInjection(ctx, task)
	case TaskTypeRunAlgorithm:
		logrus.Debug("executeAlgorithm")
		err = executeAlgorithm(ctx, task)
	case TaskTypeBuildImages:
		logrus.Debug("executeBuildImages")
		err = executeBuildImages(ctx, task)
	case TaskTypeBuildDataset:
		logrus.Debug("executeBuildDataset")
		err = executeBuildDataset(ctx, task)
	case TaskTypeCollectResult:
		logrus.Debug("executeCollectResult")
		err = executeCollectResult(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	if err != nil {
		updateTaskStatus(task.TaskID, task.TraceID,
			err.Error(),
			map[string]any{
				RdbMsgStatus:   TaskStatusError,
				RdbMsgTaskType: TaskTypeCollectResult,
			})

		return err
	}

	return nil
}
