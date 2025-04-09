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

func (e *Executor) HandleCRDUpdate(namespace, pod, name string, startTime, endTime time.Time) {
	logEntry := logrus.WithField("dataset", name)
	taskID := checkDatasetIndex(name)
	if taskID == "" {
		return
	}

	meta, err := getTaskMeta(taskID)
	if err != nil {
		logEntry.Errorf("failed to obtain task metadata")
		return
	}

	updateTaskStatus(taskID, meta.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, taskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusCompleted,
			consts.RdbMsgTaskID:   taskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	if err := repository.UpdateTimeByDataset(name, startTime, endTime); err != nil {
		logEntry.Errorf("update execution times failed: %v", err)
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
			consts.RdbMsgTaskID:   jobLabel.TaskID,
			consts.RdbMsgTaskType: jobLabel.Type,
		})
}

func (e *Executor) HandleJobUpdate(labels map[string]string, status, errorMsg string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Errorf("parse job labels failed: %v", err)
		return
	}

	logEntry := logrus.WithField("task_id", jobLabel.TaskID).WithField("trace_id", jobLabel.TraceID)

	switch status {
	case consts.TaskStatusCompleted:
		e.handleJobCompleted(logEntry, jobLabel, errorMsg)
	case consts.TaskStatusError:
		e.handleJobError(logEntry, jobLabel, errorMsg)
	default:
		logEntry.Warnf("unhandled job status: %s", status)
	}
}

func (e *Executor) handleJobCompleted(logEntry *logrus.Entry, jobLabel *JobLabel, errorMsg string) {
	baseFields := map[string]any{
		consts.RdbMsgStatus:   consts.TaskStatusCompleted,
		consts.RdbMsgTaskID:   jobLabel.TaskID,
		consts.RdbMsgTaskType: jobLabel.Type,
	}

	switch jobLabel.Type {
	case consts.TaskTypeBuildDataset:
		logEntry.WithField("dataset", jobLabel.Dataset).Info("dataset build successfully")
		e.updateDataset(logEntry, jobLabel, consts.TaskStatusCompleted, consts.DatasetBuildSuccess, baseFields)

	case consts.TaskTypeRunAlgorithm:
		e.handleAlgorithmCompletion(logEntry, jobLabel, baseFields)

	default:
		logEntry.Warnf("Unhandled completed task type: %s", jobLabel.Type)
	}
}

func (e *Executor) handleJobError(logEntry *logrus.Entry, jobLabel *JobLabel, errorMsg string) {
	if jobLabel.Type == consts.TaskTypeBuildDataset {
		logEntry.WithField("dataset", jobLabel.Dataset).Errorf("dataset build failed: %v", errorMsg)

		fields := map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskType: jobLabel.Type,
			consts.RdbMsgError:    errorMsg,
		}

		e.updateDataset(logEntry, jobLabel, consts.TaskStatusError, consts.DatasetBuildFailed, fields)
	}
}

func (e *Executor) updateDataset(logEntry *logrus.Entry, jobLabel *JobLabel, taskStatus string, datasetStatus int, fields map[string]any) {
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
		logEntry.Errorf("update dataset status to %v failed: %v", datasetStatus, err)
	}
}

func (e *Executor) handleAlgorithmCompletion(logEntry *logrus.Entry, jobLabel *JobLabel, baseFields map[string]any) {
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
		logEntry.Error("submit result collection task failed")
	}
}
