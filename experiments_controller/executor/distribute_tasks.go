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
		logrus.Debug("executeFaultInjection")
		err = executeFaultInjection(ctx, task)
	case consts.TaskTypeRunAlgorithm:
		logrus.Debug("executeAlgorithm")
		err = executeAlgorithm(ctx, task)
	case consts.TaskTypeBuildImages:
		logrus.Debug("executeBuildImages")
		err = executeBuildImages(ctx, task)
	case consts.TaskTypeBuildDataset:
		logrus.Debug("executeBuildDataset")
		err = executeBuildDataset(ctx, task)
	case consts.TaskTypeCollectResult:
		logrus.Debug("executeCollectResult")
		err = executeCollectResult(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	if err != nil {
		return err
	}

	return nil
}
