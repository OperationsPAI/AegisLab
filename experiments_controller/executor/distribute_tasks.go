package executor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/sirupsen/logrus"
)

func dispatchTask(ctx context.Context, task *UnifiedTask) error {
	logrus.WithField("task_id", task.TaskID).Info("Executing task")
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Task panic: %v\n%s", r, debug.Stack())
		}
	}()

	var err error
	switch task.Type {
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
