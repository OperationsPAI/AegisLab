package executor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/sirupsen/logrus"
)

type JobLabel struct {
	TaskID      string
	TraceID     string
	GroupID     string
	Type        consts.TaskType
	Algorithm   *string
	Dataset     string
	ExecutionID *int
}

type Executor struct {
}

var Exec *Executor

func parseJobLabel(labels map[string]string) (*JobLabel, error) {
	taskID, ok := labels[consts.LabelTaskID]
	if !ok || taskID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.LabelTaskID)
	}

	traceID, ok := labels[consts.LabelTraceID]
	if !ok || traceID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.LabelTraceID)
	}

	groupID, ok := labels[consts.LabelGroupID]
	if !ok || groupID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.LabelGroupID)
	}

	taskType, ok := labels[consts.LabelTaskType]
	if !ok || taskType == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.LabelTaskType)
	}

	dataset, ok := labels[consts.LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.LabelDataset)
	}

	var algorithm *string
	if algo, ok := labels[consts.LabelAlgorithm]; ok {
		algorithm = &algo
	}

	var executionID *int
	executionIDStr, ok := labels[consts.LabelExecutionID]
	if ok && executionIDStr != "" {
		id, err := strconv.Atoi(executionIDStr)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.LabelExecutionID)
		}
		executionID = &id
	}

	return &JobLabel{
		TaskID:      taskID,
		TraceID:     traceID,
		GroupID:     groupID,
		Type:        consts.TaskType(taskType),
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
	}, nil
}

func (e *Executor) HandleCRDUpdate(namespace, pod, name string) {
	taskID := checkDatasetIndex(name)
	if taskID == "" {
		return
	}

	meta, err := getTaskMeta(taskID)
	if err != nil {
		logrus.Errorf("failed to obtain task metadata")
		return
	}

	updateTaskStatus(taskID, meta.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, taskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusCompleted,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	startTime, endTime, err := checkExecutionTime(name, namespace)
	logrus.Infof("check execution time for dataset %s, startTime: %v, endTime: %v", name, startTime, endTime)
	if err != nil {
		logrus.WithField("dataset", name).Errorf("failed to check execution time for dataset: %v", err)
		return
	}

	datasetPayload := map[string]any{
		consts.BuildBenchmark:   meta.Benchmark,
		consts.BuildDataset:     name,
		consts.BuildNamespace:   namespace,
		consts.BuildPreDuration: meta.PreDuration,
		consts.BuildService:     pod,
		consts.BuildStartTime:   startTime,
		consts.BuildEndTime:     endTime,
	}
	if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
		Type:      consts.TaskTypeBuildDataset,
		Payload:   datasetPayload,
		Immediate: true,
		TraceID:   meta.TraceID,
		GroupID:   meta.GroupID,
	}); err != nil {
		logrus.Error(err)
		return
	}
}

func (e *Executor) HandleJobAdd(labels map[string]string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
		return
	}

	var message string
	switch jobLabel.Type {
	case consts.TaskTypeBuildDataset:
		message = fmt.Sprintf("Building dataset for task %s", jobLabel.TaskID)
	case consts.TaskTypeRunAlgorithm:
		message = fmt.Sprintf("Running algorithm for task %s", jobLabel.TaskID)
	}

	updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
		message,
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskType: jobLabel.Type,
		})
}

func (e *Executor) HandleJobUpdate(labels map[string]string, status string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
		return
	}

	if status == consts.TaskStatusCompleted {
		if jobLabel.Type == consts.TaskTypeBuildDataset {
			logrus.Infof(fmt.Sprintf("Dataset %s built", jobLabel.Dataset))

			updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
				fmt.Sprintf(consts.TaskMsgCompleted, jobLabel.TaskID),
				map[string]any{
					consts.RdbMsgStatus:   consts.TaskStatusCompleted,
					consts.RdbMsgTaskType: jobLabel.Type,
					consts.RdbMsgDataset:  jobLabel.Dataset,
				})

			if err := repository.UpdateStatusByDataset(jobLabel.Dataset, consts.DatasetBuildSuccess); err != nil {
				logrus.WithField("dataset", jobLabel.Dataset).Errorf("Failed to update record for dataset %s: %v", jobLabel.Dataset, err)
			}
		}

		if jobLabel.Type == consts.TaskTypeRunAlgorithm {
			algorithm := *jobLabel.Algorithm
			executionID := *jobLabel.ExecutionID
			logrus.Infof(fmt.Sprintf("Algorithm %s executed", algorithm))

			updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
				fmt.Sprintf(consts.TaskMsgCompleted, jobLabel.TaskID),
				map[string]any{
					consts.RdbMsgStatus:      consts.TaskStatusCompleted,
					consts.RdbMsgTaskType:    jobLabel.Type,
					consts.RdbMsgExecutionID: executionID,
				})

			payload := map[string]any{
				consts.CollectAlgorithm:   algorithm,
				consts.CollectDataset:     jobLabel.Dataset,
				consts.CollectExecutionID: executionID,
			}
			if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
				Type:      consts.TaskTypeCollectResult,
				Payload:   payload,
				Immediate: true,
				TraceID:   jobLabel.TraceID,
				GroupID:   jobLabel.GroupID,
			}); err != nil {
				logrus.Error(err)
				return
			}

			return
		}
	}

	if status == consts.TaskStatusError {
	}
}
