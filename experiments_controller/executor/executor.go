package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	batchv1 "k8s.io/api/batch/v1"
)

type Annotations struct {
	TaskCarrier  propagation.MapCarrier
	TraceCarrier propagation.MapCarrier
}

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

func (e *Executor) HandleCRDAdd(annotations map[string]string, labels map[string]string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	parsedLabels, _ := parseCRDLabels(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)

	updateTaskStatus(
		ctx,
		parsedLabels.TaskID,
		parsedLabels.TraceID,
		fmt.Sprintf("executing fault injection for task %s", parsedLabels.TaskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskID:   parsedLabels.TaskID,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})
}

func (e *Executor) HandleCRDFailed(name string, annotations map[string]string, labels map[string]string, err error, errMsg string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	parsedLabels, _ := parseCRDLabels(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)

	updateTaskError(
		ctx,
		parsedLabels.TaskID,
		parsedLabels.TraceID,
		consts.TaskTypeFaultInjection,
		err,
		errMsg,
	)
}

func (e *Executor) HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	parsedLabels, _ := parseCRDLabels(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)

	if err := repository.UpdateTimeByDataset(name, startTime, endTime); err != nil {
		logrus.WithFields(logrus.Fields{
			"task_id":  parsedLabels.TaskID,
			"trace_id": parsedLabels.TraceID,
		}).Error(err)

		updateTaskError(
			ctx,
			parsedLabels.TaskID,
			parsedLabels.TraceID,
			consts.TaskTypeFaultInjection,
			err,
			"update execution times failed",
		)

		return
	}

	updateTaskStatus(
		ctx,
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

	taskID, traceID, err := SubmitTask(ctx, &UnifiedTask{
		Type:         consts.TaskTypeBuildDataset,
		Payload:      datasetPayload,
		Immediate:    true,
		TraceID:      parsedLabels.TraceID,
		GroupID:      parsedLabels.GroupID,
		TraceCarrier: parsedAnnotations.TraceCarrier,
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
	}
}

func (e *Executor) HandleJobAdd(annotations map[string]string, labels map[string]string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	taskOptions, _ := parseTaskOptions(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)

	var message string
	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		message = fmt.Sprintf("building dataset for task %s", taskOptions.TaskID)
	case consts.TaskTypeRunAlgorithm:
		message = fmt.Sprintf("running algorithm for task %s", taskOptions.TaskID)
	}

	updateTaskStatus(
		ctx,
		taskOptions.TaskID,
		taskOptions.TraceID,
		message,
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskID:   taskOptions.TaskID,
			consts.RdbMsgTaskType: taskOptions.Type,
		})
}

func (e *Executor) HandleJobFailed(job *batchv1.Job, annotations map[string]string, labels map[string]string, err error, errMsg string) {
	logs, err := k8s.GetJobPodLogs(context.Background(), job.Namespace, job.Name)
	if err != nil {
		logrus.WithField("job_name", job.Name).Errorf("failed to get job logs: %v", err)
	}

	for podName, log := range logs {
		logrus.WithField("pod_name", podName).Errorf("job logs: %s", log)
	}
	podLog := logs[job.Name]

	parsedAnnotations, _ := parseAnnotations(annotations)
	taskOptions, _ := parseTaskOptions(labels)
	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)
	span := trace.SpanFromContext(taskCtx)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  taskOptions.TaskID,
		"trace_id": taskOptions.TraceID,
	})

	span.AddEvent("job failed", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "logs",
			Value: attribute.StringValue(podLog),
		},
	))

	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)

		logEntry.WithField("dataset", options.Dataset).Errorf("dataset build failed: %v", errMsg)

		fields := map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskID:   taskOptions.TaskID,
			consts.RdbMsgTaskType: taskOptions.Type,
			consts.RdbMsgErr:      err,
			consts.RdbMsgErrMsg:   errMsg,
		}

		updateTaskStatus(
			taskCtx,
			taskOptions.TaskID,
			taskOptions.TraceID,
			fmt.Sprintf(consts.TaskMsgFailed, taskOptions.TaskID),
			fields,
		)
		if err := repository.UpdateStatusByDataset(options.Dataset, consts.DatasetBuildFailed); err != nil {
			span.AddEvent("update dataset status failed")
			updateTaskError(
				taskCtx,
				taskOptions.TaskID,
				taskOptions.TraceID,
				taskOptions.Type,
				err,
				podLog,
			)
		}
	default:

	}

}

func (e *Executor) HandleJobSucceeded(annotations map[string]string, labels map[string]string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	taskOptions, _ := parseTaskOptions(labels)
	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)
	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TraceCarrier)
	taskSpan := trace.SpanFromContext(taskCtx)

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
		logEntry.WithField("algorithm", options.Algorithm).Info("algorithm execute successfully")
		updateFields := utils.CloneMap(baseFields)
		updateFields[consts.RdbMsgExecutionID] = options.ExecutionID

		updateTaskStatus(
			taskCtx,
			taskOptions.TaskID,
			taskOptions.TraceID,
			fmt.Sprintf(consts.TaskMsgCompleted, taskOptions.TaskID),
			updateFields,
		)

		task := &UnifiedTask{
			Type: consts.TaskTypeCollectResult,
			Payload: map[string]any{
				consts.CollectAlgorithm:   options.Algorithm,
				consts.CollectDataset:     options.Dataset,
				consts.CollectExecutionID: options.ExecutionID,
			},
			Immediate: true,
			TraceID:   taskOptions.TraceID,
			GroupID:   taskOptions.GroupID,
		}
		task.SetTraceCtx(traceCtx)

		_, _, err := SubmitTask(taskCtx, task)

		if err != nil {
			logEntry.WithField("algorithm", options.Algorithm).Errorf("submit result collection task failed: %v", err)
			taskSpan.AddEvent("submit result collection task failed")
			taskSpan.RecordError(err)

		}

	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)

		logEntry.WithField("dataset", options.Dataset).Info("dataset build successfully")

		updateFields := utils.CloneMap(baseFields)
		updateFields[consts.RdbMsgDataset] = options.Dataset
		updateTaskStatus(
			taskCtx,
			taskOptions.TaskID,
			taskOptions.TraceID,
			fmt.Sprintf(consts.TaskMsgCompleted, taskOptions.TaskID),
			updateFields,
		)

		// TODO: replace with config.string, rather than hardcode
		image := "detector"
		if _, err := client.GetHarborClient().GetLatestTag(image); err != nil {
			logrus.Errorf("failed to get latest tag of %s: %v", image, err)
		}

		if err := repository.UpdateStatusByDataset(options.Dataset, consts.DatasetBuildSuccess); err != nil {
			logrus.WithField("dataset", options.Dataset).Errorf("update dataset status failed: %v", err)
			taskSpan.AddEvent("update dataset status failed")
			return
		}

		executionPayload := map[string]any{
			consts.ExecuteImage:   image,
			consts.ExecuteTag:     "latest",
			consts.ExecuteDataset: options.Dataset,
			consts.ExecuteEnvVars: map[string]string{
				consts.ExecuteEnvVarService: options.Service,
			},
		}

		task := &UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   executionPayload,
			Immediate: true,
			TraceID:   taskOptions.TraceID,
			GroupID:   taskOptions.GroupID,
		}
		task.SetTraceCtx(traceCtx)

		_, _, err := SubmitTask(traceCtx, task)
		if err != nil {
			logEntry.WithField("dataset", options.Dataset).Errorf("submit algorithm execution task failed: %v", err)
			taskSpan.AddEvent("submit algorithm execution task failed")
			taskSpan.RecordError(err)
		}
	}
}

func parseAnnotations(annotations map[string]string) (*Annotations, error) {
	message := "missing or invalid '%s' key in k8s annotations"

	taskCarrierStr, ok := annotations[consts.TaskCarrier]
	if !ok {
		return nil, fmt.Errorf(message, consts.TaskCarrier)
	}

	var taskCarrier propagation.MapCarrier
	if err := json.Unmarshal([]byte(taskCarrierStr), &taskCarrier); err != nil {
		return nil, fmt.Errorf(message, consts.TaskCarrier)
	}

	traceCarrierStr, ok := annotations[consts.TraceCarrier]
	if !ok {
		return nil, fmt.Errorf(message, consts.TraceCarrier)
	}

	var traceCarrier propagation.MapCarrier
	if err := json.Unmarshal([]byte(traceCarrierStr), &traceCarrier); err != nil {
		return nil, fmt.Errorf(message, consts.TraceCarrier)
	}

	return &Annotations{
		TaskCarrier:  taskCarrier,
		TraceCarrier: traceCarrier,
	}, nil
}

func parseCRDLabels(labels map[string]string) (*CRDLabels, error) {
	message := "missing or invalid '%s' key in k8s labels"

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
