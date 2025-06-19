package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	batchv1 "k8s.io/api/batch/v1"
)

type Carriers struct {
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
		parsedLabels.TraceID,
		parsedLabels.TaskID,
		"injecting",
		consts.TaskStatusRunning,
		consts.TaskTypeFaultInjection,
	)
}

func (e *Executor) HandleCRDFailed(name string, annotations map[string]string, labels map[string]string, err error, errMsg string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	parsedLabels, _ := parseCRDLabels(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)

	updateTaskStatus(
		ctx,
		parsedLabels.TraceID,
		parsedLabels.TaskID,
		errMsg,
		consts.TaskStatusError,
		consts.TaskTypeFaultInjection,
	)

	repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.TraceID), dto.StreamEvent{
		TaskID:    parsedLabels.TaskID,
		TaskType:  consts.TaskTypeFaultInjection,
		EventName: consts.EventFaultInjectionFailed,
		Payload: dto.InfoPayloadTemplate{
			Status: consts.TaskStatusError,
			Msg:    errMsg,
		},
	})
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

		updateTaskStatus(
			ctx,
			parsedLabels.TraceID,
			parsedLabels.TaskID,
			"update execution times failed",
			consts.TaskStatusError,
			consts.TaskTypeFaultInjection,
		)

		return
	}

	updateTaskStatus(
		ctx,
		parsedLabels.TraceID,
		parsedLabels.TaskID,
		"injection completed",
		consts.TaskStatusCompleted,
		consts.TaskTypeFaultInjection,
	)

	envVars := map[string]string{
		consts.BuildEnvVarNamespace: namespace,
	}
	datasetPayload := map[string]any{
		consts.BuildBenchmark:   parsedLabels.Benchmark,
		consts.BuildDataset:     name,
		consts.BuildPreDuration: parsedLabels.PreDuration,
		consts.BuildEnvVars:     envVars,
		consts.BuildStartTime:   startTime,
		consts.BuildEndTime:     endTime,
	}

	repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.TraceID), dto.StreamEvent{
		TaskID:    parsedLabels.TaskID,
		TaskType:  consts.TaskTypeFaultInjection,
		EventName: consts.EventFaultInjectionCompleted,
	})

	taskID, traceID, err := SubmitTask(ctx, &dto.UnifiedTask{
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
		taskOptions.TraceID,
		taskOptions.TaskID,
		message,
		consts.TaskStatusRunning,
		taskOptions.Type,
	)
}

func (e *Executor) HandleJobFailed(job *batchv1.Job, annotations map[string]string, labels map[string]string, errC error, errMsg string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	taskOptions, _ := parseTaskOptions(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)
	span := trace.SpanFromContext(ctx)

	logs, err := k8s.GetJobPodLogs(ctx, job.Namespace, job.Name)
	if err != nil {
		logrus.WithField("job_name", job.Name).Errorf("failed to get job logs: %v", err)
	}

	for podName, log := range logs {
		logrus.WithField("pod_name", podName).Error("job logs:")
		logrus.Error(log)
	}

	podLog := logs[job.Name]
	span.AddEvent("job failed", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "logs",
			Value: attribute.StringValue(podLog),
		},
	))

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  taskOptions.TaskID,
		"trace_id": taskOptions.TraceID,
	})
	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)

		logEntry.WithField("dataset", options.Dataset).Errorf("dataset build failed: %v", errMsg)
		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeBuildDataset,
			EventName: consts.EventDatasetBuildFailed,
			Payload: dto.InfoPayloadTemplate{
				Status: consts.TaskStatusError,
				Msg:    errC.Error(),
			},
		}, repository.WithCallerLevel(4))

		if err := repository.UpdateStatusByDataset(options.Dataset, consts.DatasetBuildFailed); err != nil {
			span.AddEvent("update dataset status failed")
			span.RecordError(err)
			updateTaskStatus(
				ctx,
				taskOptions.TraceID,
				taskOptions.TaskID,
				"update dataset status failed",
				consts.TaskStatusError,
				taskOptions.Type,
			)
		}

	case consts.TaskTypeRunAlgorithm:
		options, _ := parseExecutionOptions(labels)

		logEntry.WithFields(logrus.Fields{
			"algorithm": options.Algorithm,
			"dataset":   options.Dataset,
		}).Errorf("algorithm execute failed: %v", errMsg) //TODO errMsg为空

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeRunAlgorithm,
			EventName: consts.EventAlgoRunFailed,
			Payload: dto.InfoPayloadTemplate{
				Status: consts.TaskStatusError,
				Msg:    errC.Error(),
			},
		}, repository.WithCallerLevel(4))

		if err := repository.UpdateStatusByExecID(options.ExecutionID, consts.ExecutionFailed); err != nil {
			span.AddEvent("update execution status failed")
			span.RecordError(err)
			updateTaskStatus(
				ctx,
				taskOptions.TraceID,
				taskOptions.TaskID,
				"update execution status failed",
				consts.TaskStatusError,
				taskOptions.Type,
			)
		}
	}

	updateTaskStatus(
		ctx,
		taskOptions.TraceID,
		taskOptions.TaskID,
		fmt.Sprintf(consts.TaskMsgFailed, taskOptions.TaskID),
		consts.TaskStatusError,
		taskOptions.Type,
	)
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

	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)
		logEntry = logEntry.WithFields(logrus.Fields{
			"dataset": options.Dataset,
		})

		logEntry.Info("dataset build successfully")
		repository.PublishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeBuildDataset,
			EventName: consts.EventDatasetBuildSucceed,
			Payload:   options.Dataset,
		}, repository.WithCallerLevel(4))

		updateTaskStatus(
			taskCtx,
			taskOptions.TraceID,
			taskOptions.TaskID,
			fmt.Sprintf(consts.TaskMsgCompleted, taskOptions.TaskID),
			consts.TaskStatusCompleted,
			taskOptions.Type,
		)

		if err := repository.UpdateStatusByDataset(options.Dataset, consts.DatasetBuildSuccess); err != nil {
			logEntry.Errorf("update dataset status failed: %v", err)
			taskSpan.AddEvent("update dataset status failed")
			return
		}

		task := &dto.UnifiedTask{
			Type: consts.TaskTypeRunAlgorithm,
			Payload: map[string]any{
				consts.ExecuteAlgorithm: config.GetString("algo.detector"),
				consts.ExecuteDataset:   options.Dataset,
			},
			Immediate: true,
			TraceID:   taskOptions.TraceID,
			GroupID:   taskOptions.GroupID,
		}
		task.SetTraceCtx(traceCtx)

		_, _, err := SubmitTask(traceCtx, task)
		if err != nil {
			logEntry.Errorf("submit algorithm execution task failed: %v", err)
			taskSpan.AddEvent("submit algorithm execution task failed")
			taskSpan.RecordError(err)
		}

	case consts.TaskTypeRunAlgorithm:
		options, _ := parseExecutionOptions(labels)
		logEntry = logEntry.WithFields(logrus.Fields{
			"algorithm": options.Algorithm,
			"dataset":   options.Dataset,
		})

		logEntry.Info("algorithm execute successfully")
		repository.PublishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeRunAlgorithm,
			EventName: consts.EventAlgoRunSucceed,
			Payload:   options,
		}, repository.WithCallerLevel(4))

		updateTaskStatus(
			taskCtx,
			taskOptions.TraceID,
			taskOptions.TaskID,
			fmt.Sprintf(consts.TaskMsgCompleted, taskOptions.TaskID),
			consts.TaskStatusCompleted,
			taskOptions.Type,
		)

		if err := repository.UpdateStatusByExecID(options.ExecutionID, consts.ExecutionSuccess); err != nil {
			logEntry.Errorf("update execution status failed: %v", err)
			taskSpan.AddEvent("update execution status failed")
			return
		}

		task := &dto.UnifiedTask{
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
			logEntry.Errorf("submit result collection task failed: %v", err)
			taskSpan.AddEvent("submit result collection task failed")
			taskSpan.RecordError(err)
		}
	}
}

func parseAnnotations(annotations map[string]string) (*Carriers, error) {
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

	return &Carriers{
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

	return &DatasetOptions{
		Dataset: dataset,
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
