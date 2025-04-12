package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/sirupsen/logrus"
)

type JobLabel struct {
	TaskID      string
	TraceID     string
	GroupID     string
	Type        consts.TaskType
	Algorithm   string
	Dataset     string
	ExecutionID int
}

type Executor struct {
}

var Exec *Executor

func (e *Executor) HandleCRDFailed(name, errorMsg string) error {
	taskID := checkDatasetIndex(name)
	if taskID == "" {
		return fmt.Errorf("failed to get taskID")
	}

	meta, err := getTaskMeta(taskID)
	if err != nil {
		return fmt.Errorf("failed to obtain task metadata")
	}

	updateTaskStatus(taskID, meta.TraceID,
		fmt.Sprintf(consts.TaskMsgFailed, taskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskID:   taskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
			consts.RdbMsgError:    errorMsg,
		})

	return nil
}

func (e *Executor) HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time) error {
	taskID := checkDatasetIndex(name)
	if taskID == "" {
		return fmt.Errorf("failed to get taskID")
	}

	meta, err := getTaskMeta(taskID)
	if err != nil {
		return fmt.Errorf("failed to obtain task metadata")
	}

	updateTaskStatus(taskID, meta.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, taskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusCompleted,
			consts.RdbMsgTaskID:   taskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	if err := repository.UpdateTimeByDataset(name, startTime, endTime); err != nil {
		return fmt.Errorf("update execution times failed: %v", err)
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
		return err
	}

	return nil
}

func (e *Executor) HandleJobAdd(labels map[string]string) error {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		return fmt.Errorf("parse job labels failed: %v", err)
	}

	var message string
	switch jobLabel.Type {
	case consts.TaskTypeBuildDataset:
		message = fmt.Sprintf("building dataset for task %s", jobLabel.TaskID)
	case consts.TaskTypeRunAlgorithm:
		message = fmt.Sprintf("running algorithm for task %s", jobLabel.TaskID)
	}

	updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
		message,
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskID:   jobLabel.TaskID,
			consts.RdbMsgTaskType: jobLabel.Type,
		})

	return nil
}

func (e *Executor) HandleJobFailed(labels map[string]string, errorMsg string) error {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		return fmt.Errorf("parse job labels failed: %v", err)
	}

	logEntry := logrus.WithField("task_id", jobLabel.TaskID).WithField("trace_id", jobLabel.TraceID)

	if jobLabel.Type == consts.TaskTypeBuildDataset {
		logEntry.WithFields(logrus.Fields{
			"task_id":  jobLabel.TaskID,
			"trace_id": jobLabel.TraceID,
			"dataset":  jobLabel.Dataset,
		}).Errorf("dataset build failed: %v", errorMsg)

		fields := map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskID:   jobLabel.TaskID,
			consts.RdbMsgTaskType: jobLabel.Type,
			consts.RdbMsgError:    errorMsg,
		}

		if err := e.updateDataset(jobLabel, consts.TaskStatusError, consts.DatasetBuildFailed, fields); err != nil {
			return fmt.Errorf("failed to udpate dataset: %v", err)
		}
	}

	return nil
}

func (e *Executor) HandleJobSucceeded(labels map[string]string) error {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		message := "parse job labels failed"
		logrus.Errorf("%s: %v", message, err)
		return err
	}

	logEntry := logrus.WithField("task_id", jobLabel.TaskID).WithField("trace_id", jobLabel.TraceID)

	baseFields := map[string]any{
		consts.RdbMsgStatus:   consts.TaskStatusCompleted,
		consts.RdbMsgTaskID:   jobLabel.TaskID,
		consts.RdbMsgTaskType: jobLabel.Type,
	}

	switch jobLabel.Type {
	case consts.TaskTypeBuildDataset:
		logEntry.WithField("dataset", jobLabel.Dataset).Info("dataset build successfully")
		if err := e.updateDataset(jobLabel, consts.TaskStatusCompleted, consts.DatasetBuildSuccess, baseFields); err != nil {
			return fmt.Errorf("failed to udpate dataset: %v", err)
		}

	case consts.TaskTypeRunAlgorithm:
		if err := e.updateAlgorithm(logEntry, jobLabel, baseFields); err != nil {
			return fmt.Errorf("failed to udpate algorithm: %v", err)
		}

	default:
		logEntry.Warnf("unhandled completed task type: %s", jobLabel.Type)
	}

	return nil
}

func parseJobLabel(labels map[string]string) (*JobLabel, error) {
	message := "missing or invalid '%s' key in payload"

	taskID, ok := labels[consts.LabelTaskID]
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.LabelTaskID)
	}

	traceID, ok := labels[consts.LabelTraceID]
	if !ok || traceID == "" {
		return nil, fmt.Errorf(message, consts.LabelTraceID)
	}

	groupID, ok := labels[consts.LabelGroupID]
	if !ok || groupID == "" {
		return nil, fmt.Errorf(message, consts.LabelGroupID)
	}

	taskType, ok := labels[consts.LabelTaskType]
	if !ok || taskType == "" {
		return nil, fmt.Errorf(message, consts.LabelTaskType)
	}

	dataset, ok := labels[consts.LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.LabelDataset)
	}

	algorithm, ok := labels[consts.LabelDataset]
	if !ok || algorithm == "" {
		return nil, fmt.Errorf(message, consts.LabelAlgorithm)
	}

	var executionID int
	executionIDStr, ok := labels[consts.LabelExecutionID]
	if ok && executionIDStr != "" {
		id, err := strconv.Atoi(executionIDStr)
		if err != nil {
			return nil, fmt.Errorf(message, consts.LabelExecutionID)
		}

		executionID = id
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

func (e *Executor) updateDataset(jobLabel *JobLabel, taskStatus string, datasetStatus int, fields map[string]any) error {
	if datasetStatus == consts.DatasetBuildSuccess {
		updateFields := utils.CloneMap(fields)
		updateFields[consts.RdbMsgDataset] = jobLabel.Dataset
		updateTaskStatus(jobLabel.TaskID,
			jobLabel.TraceID,
			fmt.Sprintf(consts.TaskMsgCompleted, jobLabel.TaskID),
			updateFields,
		)
	} else {
		updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID, fmt.Sprintf(taskStatus, jobLabel.TaskID), fields)
	}

	if err := repository.UpdateStatusByDataset(jobLabel.Dataset, datasetStatus); err != nil {
		return fmt.Errorf("update dataset status to %v failed: %v", datasetStatus, err)
	}

	return nil
}

func (e *Executor) updateAlgorithm(logEntry *logrus.Entry, jobLabel *JobLabel, baseFields map[string]any) error {
	logEntry.WithField("algorithm", jobLabel.Algorithm).Info("algorithm execute successfully")

	updateFields := utils.CloneMap(baseFields)
	updateFields[consts.RdbMsgExecutionID] = jobLabel.ExecutionID
	updateTaskStatus(jobLabel.TaskID,
		jobLabel.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, jobLabel.TaskID),
		updateFields,
	)

	if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
		Type: consts.TaskTypeCollectResult,
		Payload: map[string]any{
			consts.CollectAlgorithm:   jobLabel.Algorithm,
			consts.CollectDataset:     jobLabel.Dataset,
			consts.CollectExecutionID: jobLabel.ExecutionID,
		},
		Immediate: true,
		TraceID:   jobLabel.TraceID,
		GroupID:   jobLabel.GroupID,
	}); err != nil {
		return fmt.Errorf("submit result collection task failed")
	}

	return nil
}
