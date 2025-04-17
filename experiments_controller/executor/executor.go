package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
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

type TaskOptions struct {
	TaskID  string
	TraceID string
	GroupID string
	Type    consts.TaskType
}

type DatasetOptions struct {
	Dataset string
	Service string
}

type ExecutionOptions struct {
	Algorithm   string
	Dataset     string
	ExecutionID int
}

type Executor struct {
}

var Exec *Executor

func (e *Executor) HandleCRDFailed(name, errorMsg string, labels map[string]string) {
	parsedLabels, _ := parseCRDLabels(labels)

	updateTaskError(
		parsedLabels.TaskID,
		parsedLabels.TraceID,
		consts.TaskTypeFaultInjection,
		errorMsg,
	)
}

func (e *Executor) HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time, labels map[string]string) {
	parsedLabels, _ := parseCRDLabels(labels)

	if err := repository.UpdateTimeByDataset(name, startTime, endTime); err != nil {
		logrus.WithFields(logrus.Fields{
			"task_id":  parsedLabels.TaskID,
			"trace_id": parsedLabels.TraceID,
		}).Error(err)

		updateTaskError(
			parsedLabels.TaskID,
			parsedLabels.TraceID,
			consts.TaskTypeFaultInjection,
			"update execution times failed",
		)

		return
	}

	updateTaskStatus(
		parsedLabels.TaskID,
		parsedLabels.TraceID,
		fmt.Sprintf(consts.TaskMsgCompleted, parsedLabels.TaskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusCompleted,
			consts.RdbMsgTaskID:   parsedLabels.TaskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	envVars := map[string]string{
		consts.BuildEnvVarNamespace: namespace,
		consts.BuildEnvVarService:   pod,
	}
	datasetPayload := map[string]any{
		consts.BuildBenchmark:   parsedLabels.Benchmark,
		consts.BuildDataset:     name,
		consts.BuildPreDuration: parsedLabels.PreDuration,
		consts.BuildEnvVars:     envVars,
		consts.BuildStartTime:   startTime,
		consts.BuildEndTime:     endTime,
	}
	taskID, traceID, err := SubmitTask(context.Background(), &UnifiedTask{
		Type:      consts.TaskTypeBuildDataset,
		Payload:   datasetPayload,
		Immediate: true,
		TraceID:   parsedLabels.TraceID,
		GroupID:   parsedLabels.GroupID,
	})
	if err != nil {
		if taskID == "" && traceID == "" {
			logrus.Error(err)
			return
		}

		logrus.WithFields(logrus.Fields{
			"task_id":  taskID,
			"trace_id": traceID,
		}).Error(err)

		updateTaskError(
			taskID,
			traceID,
			consts.BuildDataset,
			"failed to submit task",
		)
	}
}

func (e *Executor) HandleJobAdd(labels map[string]string) {
	taskOptions, _ := parseTaskOptions(labels)

	var message string
	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		message = fmt.Sprintf("building dataset for task %s", taskOptions.TaskID)
	case consts.TaskTypeRunAlgorithm:
		message = fmt.Sprintf("running algorithm for task %s", taskOptions.TaskID)
	}

	updateTaskStatus(
		taskOptions.TaskID,
		taskOptions.TraceID,
		message,
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskID:   taskOptions.TaskID,
			consts.RdbMsgTaskType: taskOptions.Type,
		})
}

func (e *Executor) HandleJobFailed(labels map[string]string, errorMsg string) {
	taskOptions, _ := parseTaskOptions(labels)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  taskOptions.TaskID,
		"trace_id": taskOptions.TraceID,
	})

	if taskOptions.Type == consts.TaskTypeBuildDataset {
		options, _ := parseDatasetOptions(labels)

		logEntry.WithField("dataset", options.Dataset).Errorf("dataset build failed: %v", errorMsg)

		fields := map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskID:   taskOptions.TaskID,
			consts.RdbMsgTaskType: taskOptions.Type,
			consts.RdbMsgError:    errorMsg,
		}

		if err := e.updateDataset(taskOptions, options, consts.TaskStatusError, consts.DatasetBuildFailed, fields); err != nil {
			updateTaskError(
				taskOptions.TaskID,
				taskOptions.TraceID,
				taskOptions.Type,
				"failed to udpate dataset",
			)
		}
	}
}

func (e *Executor) HandleJobSucceeded(labels map[string]string) {
	taskOptions, _ := parseTaskOptions(labels)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  taskOptions.TaskID,
		"trace_id": taskOptions.TraceID,
	})

	baseFields := map[string]any{
		consts.RdbMsgStatus:   consts.TaskStatusCompleted,
		consts.RdbMsgTaskID:   taskOptions.TaskID,
		consts.RdbMsgTaskType: taskOptions.Type,
	}

	switch taskOptions.Type {
	case consts.TaskTypeRunAlgorithm:
		options, _ := parseExecutionOptions(labels)

		if err := e.updateAlgorithm(logEntry, taskOptions, options, baseFields); err != nil {
			updateTaskError(
				taskOptions.TaskID,
				taskOptions.TraceID,
				taskOptions.Type,
				"failed to udpate algorithm",
			)
		}

	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)

		logEntry.WithField("dataset", options.Dataset).Info("dataset build successfully")
		if err := e.updateDataset(taskOptions, options, consts.TaskStatusCompleted, consts.DatasetBuildSuccess, baseFields); err != nil {
			updateTaskError(
				taskOptions.TaskID,
				taskOptions.TraceID,
				taskOptions.Type,
				"failed to udpate dataset",
			)
		}
	}
}

func (e *Executor) updateDataset(taskOptions *TaskOptions, options *DatasetOptions, taskStatus string, datasetStatus int, fields map[string]any) error {
	if datasetStatus != consts.DatasetBuildSuccess {
		updateTaskStatus(
			taskOptions.TaskID,
			taskOptions.TraceID,
			fmt.Sprintf("[%s] %s", taskStatus, taskOptions.TaskID),
			fields,
		)
	} else {
		updateFields := utils.CloneMap(fields)
		updateFields[consts.RdbMsgDataset] = options.Dataset
		updateTaskStatus(
			taskOptions.TaskID,
			taskOptions.TraceID,
			fmt.Sprintf(consts.TaskMsgCompleted, taskOptions.TaskID),
			updateFields,
		)
	}

	if err := repository.UpdateStatusByDataset(options.Dataset, datasetStatus); err != nil {
		return fmt.Errorf("update dataset status to %v failed: %v", datasetStatus, err)
	}

	if datasetStatus == consts.DatasetBuildSuccess {
		image := "detector"
		_, err := client.GetHarborClient().GetLatestTag(image)
		if err != nil {
			logrus.Errorf("failed to get latest tag of %s: %v", image, err)
			return err
		}

		envVars := map[string]string{
			consts.ExecuteEnvVarService: options.Service,
		}
		executionPayload := map[string]any{
			consts.ExecuteImage:   image,
			consts.ExecuteTag:     "latest",
			consts.ExecuteDataset: options.Dataset,
			consts.ExecuteEnvVars: envVars,
		}

		if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   executionPayload,
			Immediate: true,
			TraceID:   taskOptions.TraceID,
			GroupID:   taskOptions.GroupID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) updateAlgorithm(logEntry *logrus.Entry, taskOptions *TaskOptions, options *ExecutionOptions, baseFields map[string]any) error {
	logEntry.WithField("algorithm", options.Algorithm).Info("algorithm execute successfully")

	updateFields := utils.CloneMap(baseFields)
	updateFields[consts.RdbMsgExecutionID] = options.ExecutionID
	updateTaskStatus(
		taskOptions.TaskID,
		taskOptions.TraceID,
		fmt.Sprintf("[%s] %s", consts.TaskMsgCompleted, taskOptions.TaskID),
		updateFields,
	)

	if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
		Type: consts.TaskTypeCollectResult,
		Payload: map[string]any{
			consts.CollectAlgorithm:   options.Algorithm,
			consts.CollectDataset:     options.Dataset,
			consts.CollectExecutionID: options.ExecutionID,
		},
		Immediate: true,
		TraceID:   taskOptions.TraceID,
		GroupID:   taskOptions.GroupID,
	}); err != nil {
		return fmt.Errorf("submit result collection task failed")
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

func parseTaskOptions(labels map[string]string) (*TaskOptions, error) {
	message := "missing or invalid '%s' key in job labels"

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

	return &TaskOptions{
		TaskID:  taskID,
		TraceID: traceID,
		GroupID: groupID,
		Type:    consts.TaskType(taskType),
	}, nil
}

func parseDatasetOptions(labels map[string]string) (*DatasetOptions, error) {
	message := "missing or invalid '%s' key in job labels"

	dataset, ok := labels[consts.LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.LabelDataset)
	}

	service, ok := labels[consts.LabelService]
	if !ok || service == "" {
		return nil, fmt.Errorf(message, consts.LabelService)
	}

	return &DatasetOptions{
		Dataset: dataset,
		Service: service,
	}, nil
}

func parseExecutionOptions(labels map[string]string) (*ExecutionOptions, error) {
	message := "missing or invalid '%s' key in job labels"

	algorithm, ok := labels[consts.LabelAlgorithm]
	if !ok || algorithm == "" {
		return nil, fmt.Errorf(message, consts.LabelAlgorithm)
	}

	dataset, ok := labels[consts.LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.LabelDataset)
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

	return &ExecutionOptions{
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
	}, nil
}
