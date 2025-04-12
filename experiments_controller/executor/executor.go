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

type CRDLabels struct {
	TaskID      string
	TraceID     string
	GroupID     string
	Benchmark   string
	PreDuration int
}

type JobLabels struct {
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

func (e *Executor) HandleCRDFailed(name, errorMsg string, labels map[string]string) {
	parsedLabels, _ := parseCRDLabels(labels)

	updateTaskStatus(parsedLabels.TaskID, parsedLabels.TraceID,
		fmt.Sprintf(consts.TaskMsgFailed, parsedLabels.TaskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskID:   parsedLabels.TaskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
			consts.RdbMsgError:    errorMsg,
		})
}

func (e *Executor) HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time, labels map[string]string) error {
	parsedLabels, err := parseCRDLabels(labels)
	if err != nil {
		return fmt.Errorf("failed to read CRD labels: %v", err)
	}

	if err := repository.UpdateTimeByDataset(name, startTime, endTime); err != nil {
		return fmt.Errorf("update execution times failed: %v", err)
	}

	updateTaskStatus(parsedLabels.TaskID, parsedLabels.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, parsedLabels.TaskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusCompleted,
			consts.RdbMsgTaskID:   parsedLabels.TaskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	datasetPayload := map[string]any{
		consts.BuildBenchmark:   parsedLabels.Benchmark,
		consts.BuildDataset:     name,
		consts.BuildNamespace:   namespace,
		consts.BuildPreDuration: parsedLabels.PreDuration,
		consts.BuildService:     pod,
		consts.BuildStartTime:   startTime,
		consts.BuildEndTime:     endTime,
	}
	if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
		Type:      consts.TaskTypeBuildDataset,
		Payload:   datasetPayload,
		Immediate: true,
		TraceID:   parsedLabels.TraceID,
		GroupID:   parsedLabels.GroupID,
	}); err != nil {
		return err
	}

	return nil
}

func parseCRDLabels(labels map[string]string) (*CRDLabels, error) {
	message := "missing or invalid '%s' key in payload"

	taskID, ok := labels[consts.CRDTaskID]
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.CRDTaskID)
	}

	traceID, ok := labels[consts.CRDTraceID]
	if !ok || traceID == "" {
		return nil, fmt.Errorf(message, consts.CRDTraceID)
	}

	groupID, ok := labels[consts.CRDGroupID]
	if !ok || groupID == "" {
		return nil, fmt.Errorf(message, consts.CRDGroupID)
	}

	benchmark, ok := labels[consts.CRDBenchmark]
	if !ok || benchmark == "" {
		return nil, fmt.Errorf(message, consts.CRDBenchmark)
	}

	var preDuration int
	preDurationStr, ok := labels[consts.CRDPreDuration]
	if ok && preDurationStr != "" {
		duration, err := strconv.Atoi(preDurationStr)
		if err != nil {
			return nil, fmt.Errorf(message, consts.CRDPreDuration)
		}

		preDuration = duration
	}

	return &CRDLabels{
		TaskID:      taskID,
		TraceID:     traceID,
		GroupID:     groupID,
		Benchmark:   benchmark,
		PreDuration: preDuration,
	}, nil
}

func (e *Executor) HandleJobAdd(labels map[string]string) error {
	jobLabel, err := parseJobLabels(labels)
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
	jobLabel, err := parseJobLabels(labels)
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
	jobLabel, err := parseJobLabels(labels)
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

func parseJobLabels(labels map[string]string) (*JobLabels, error) {
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

	return &JobLabels{
		TaskID:      taskID,
		TraceID:     traceID,
		GroupID:     groupID,
		Type:        consts.TaskType(taskType),
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
	}, nil
}

func (e *Executor) updateDataset(jobLabels *JobLabels, taskStatus string, datasetStatus int, fields map[string]any) error {
	if datasetStatus == consts.DatasetBuildSuccess {
		updateFields := utils.CloneMap(fields)
		updateFields[consts.RdbMsgDataset] = jobLabels.Dataset
		updateTaskStatus(jobLabels.TaskID,
			jobLabels.TraceID,
			fmt.Sprintf(consts.TaskMsgCompleted, jobLabels.TaskID),
			updateFields,
		)
	} else {
		updateTaskStatus(jobLabels.TaskID, jobLabels.TraceID, fmt.Sprintf(taskStatus, jobLabels.TaskID), fields)
	}

	if err := repository.UpdateStatusByDataset(jobLabels.Dataset, datasetStatus); err != nil {
		return fmt.Errorf("update dataset status to %v failed: %v", datasetStatus, err)
	}

	return nil
}

func (e *Executor) updateAlgorithm(logEntry *logrus.Entry, jobLabels *JobLabels, baseFields map[string]any) error {
	logEntry.WithField("algorithm", jobLabels.Algorithm).Info("algorithm execute successfully")

	updateFields := utils.CloneMap(baseFields)
	updateFields[consts.RdbMsgExecutionID] = jobLabels.ExecutionID
	updateTaskStatus(jobLabels.TaskID,
		jobLabels.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, jobLabels.TaskID),
		updateFields,
	)

	if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
		Type: consts.TaskTypeCollectResult,
		Payload: map[string]any{
			consts.CollectAlgorithm:   jobLabels.Algorithm,
			consts.CollectDataset:     jobLabels.Dataset,
			consts.CollectExecutionID: jobLabels.ExecutionID,
		},
		Immediate: true,
		TraceID:   jobLabels.TraceID,
		GroupID:   jobLabels.GroupID,
	}); err != nil {
		return fmt.Errorf("submit result collection task failed")
	}

	return nil
}
