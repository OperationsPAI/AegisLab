package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"

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

type TaskIdentifiers struct {
	TaskID    string `json:"task_id"`
	TraceID   string `json:"trace_id"`
	GroupID   string `json:"group_id"`
	ProjectID *int   `json:"project_id,omitempty"`
	UserID    *int   `json:"user_id,omitempty"` // UserID is optional and can be nil
}

type CRDLabels struct {
	TaskIdentifiers
	Benchmark   string
	PreDuration int
}

type TaskOptions struct {
	TaskIdentifiers
	Type consts.TaskType
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

func (e *Executor) HandleCRDFailed(name string, annotations map[string]string, labels map[string]string, errMsg string) {
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
	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)
	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TraceCarrier)
	taskSpan := trace.SpanFromContext(taskCtx)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  parsedLabels.TaskID,
		"trace_id": parsedLabels.TraceID,
	})

	logEntry.Info("fault injected successfully")
	repository.PublishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.TraceID), dto.StreamEvent{
		TaskID:    parsedLabels.TaskID,
		TaskType:  consts.TaskTypeFaultInjection,
		EventName: consts.EventFaultInjectionCompleted,
	}, repository.WithCallerLevel(4))

	updateTaskStatus(
		taskCtx,
		parsedLabels.TraceID,
		parsedLabels.TaskID,
		fmt.Sprintf(consts.TaskMsgCompleted, parsedLabels.TaskID),
		consts.TaskStatusCompleted,
		consts.TaskTypeFaultInjection,
	)

	if err := repository.UpdateTimeByInjectionName(name, startTime, endTime); err != nil {
		logEntry.Errorf("update injection execution times failed: %v", err)
		taskSpan.AddEvent("update injection execution times failed")
		return
	}

	task := &dto.UnifiedTask{
		Type: consts.TaskTypeBuildDataset,
		Payload: map[string]any{
			consts.BuildBenchmark:   parsedLabels.Benchmark,
			consts.BuildDataset:     name,
			consts.BuildPreDuration: parsedLabels.PreDuration,
			consts.BuildEnvVars: map[string]string{
				consts.BuildEnvVarNamespace: namespace,
			},
			consts.BuildStartTime: startTime,
			consts.BuildEndTime:   endTime,
			consts.BuildUserID:    parsedLabels.UserID,
		},
		Immediate: true,
		TraceID:   parsedLabels.TraceID,
		GroupID:   parsedLabels.GroupID,
		ProjectID: parsedLabels.ProjectID,
	}
	task.SetTraceCtx(traceCtx)

	_, _, err := SubmitTask(taskCtx, task)
	if err != nil {
		logEntry.Errorf("submit dataset building task failed: %v", err)
		taskSpan.AddEvent("submit dataset building task failed")
		taskSpan.RecordError(err)
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

func (e *Executor) HandleJobFailed(job *batchv1.Job, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	taskOptions, _ := parseTaskOptions(labels)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)
	span := trace.SpanFromContext(ctx)

	logMap, err := k8s.GetJobPodLogs(ctx, job.Namespace, job.Name)
	if err != nil {
		logrus.WithField("job_name", job.Name).Errorf("failed to get job logs: %v", err)
	}

	// Use PublishEvent to record job logs
	if len(logMap) > 0 {
		jobLogBytes, err := json.Marshal(logMap)
		if err != nil {
			logrus.WithField("job_name", job.Name).Errorf("failed to marshal job logs: %v", err)
		}

		jobLog := string(jobLogBytes)
		spanAttrs := []trace.EventOption{
			trace.WithAttributes(
				attribute.String("job_name", job.Name),
				attribute.String("namespace", job.Namespace),
			),
		}

		filePath, err := writeJobLogs(job, taskOptions, logMap)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"job_name":  job.Name,
				"pod_count": len(logMap),
			}).Errorf("Job failed - logs available but file writing disabled: %v", err)

			for podName, log := range logMap {
				logrus.WithField("pod_name", podName).Error("job logs:")
				logrus.Error(log)
			}

			spanAttrs = append(spanAttrs, trace.WithAttributes(
				attribute.String("logs", jobLog),
			))
		}

		if filePath != "" {
			spanAttrs = append(spanAttrs, trace.WithAttributes(
				attribute.String("log_file", filePath),
			))
		}

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  taskOptions.Type,
			EventName: consts.EventJobFailed,
			Payload: dto.JobMessage{
				JobName:   job.Name,
				Namespace: job.Namespace,
				LogFile:   filePath,
			},
		}, repository.WithCallerLevel(4))

		span.AddEvent("job failed", spanAttrs...)
	}

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  taskOptions.TaskID,
		"trace_id": taskOptions.TraceID,
	})

	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)

		logEntry.Errorf("dataset build failed")
		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeBuildDataset,
			EventName: consts.EventDatapackBuildFailed,
			Payload:   options,
		}, repository.WithCallerLevel(4))

		if err := repository.UpdateStatusByDataset(options.Dataset, consts.DatapackBuildFailed); err != nil {
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
		options, _ := parseExecutionOptions(annotations, labels)

		// Release algorithm execution token
		rateLimiter := utils.NewAlgoExecutionRateLimiter()
		if releaseErr := rateLimiter.ReleaseToken(ctx, taskOptions.TaskID, taskOptions.TraceID); releaseErr != nil {
			logEntry.Error("failed to release algorithm execution token on job failure")
		} else {
			logEntry.Info("successfully released algorithm execution token on job failure")
		}

		logEntry.Errorf("algorithm execute failed")
		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeRunAlgorithm,
			EventName: consts.EventAlgoRunFailed,
			Payload:   options,
		}, repository.WithCallerLevel(4))

		if err := repository.UpdateExecutionResult(options.ExecutionID, map[string]any{
			"status": consts.ExecutionFailed,
		}); err != nil {
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

func (e *Executor) HandleJobSucceeded(job *batchv1.Job, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, _ := parseAnnotations(annotations)
	taskOptions, _ := parseTaskOptions(labels)
	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TaskCarrier)
	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.TraceCarrier)
	taskSpan := trace.SpanFromContext(taskCtx)

	repository.PublishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
		TaskID:    taskOptions.TaskID,
		TaskType:  taskOptions.Type,
		EventName: consts.EventJobSucceed,
		Payload: dto.JobMessage{
			JobName:   job.Name,
			Namespace: job.Namespace,
		},
	}, repository.WithCallerLevel(4))

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  taskOptions.TaskID,
		"trace_id": taskOptions.TraceID,
	})

	switch taskOptions.Type {
	case consts.TaskTypeBuildDataset:
		options, _ := parseDatasetOptions(labels)

		logEntry.Info("dataset build successfully")
		repository.PublishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, taskOptions.TraceID), dto.StreamEvent{
			TaskID:    taskOptions.TaskID,
			TaskType:  consts.TaskTypeBuildDataset,
			EventName: consts.EventDatapackBuildSucceed,
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

		if err := repository.UpdateStatusByDataset(options.Dataset, consts.DatapackBuildSuccess); err != nil {
			logEntry.Errorf("update dataset status failed: %v", err)
			taskSpan.AddEvent("update dataset status failed")
			return
		}

		algorithm, tag, err := repository.GetContainerWithTag(consts.ContainerTypeAlgorithm, config.GetString("algo.detector"), consts.DefaultContainerTag, *taskOptions.UserID)
		if err != nil {
			logEntry.Errorf("get algorithm container failed: %v", err)
			taskSpan.AddEvent("get algorithm container failed")
			return
		}

		task := &dto.UnifiedTask{
			Type: consts.TaskTypeRunAlgorithm,
			Payload: map[string]any{
				consts.ExecuteAlgorithm:    algorithm,
				consts.ExecuteAlgorithmTag: tag,
				consts.ExecuteDataset:      options.Dataset,
				consts.ExecuteEnvVars:      map[string]string{},
			},
			Immediate: true,
			TraceID:   taskOptions.TraceID,
			GroupID:   taskOptions.GroupID,
			ProjectID: taskOptions.ProjectID,
		}
		task.SetTraceCtx(traceCtx)

		if _, _, err := SubmitTask(taskCtx, task); err != nil {
			logEntry.Errorf("submit algorithm execution task failed: %v", err)
			taskSpan.AddEvent("submit algorithm execution task failed")
			taskSpan.RecordError(err)
		}

	case consts.TaskTypeRunAlgorithm:
		options, _ := parseExecutionOptions(annotations, labels)

		// Release algorithm execution token
		rateLimiter := utils.NewAlgoExecutionRateLimiter()
		if releaseErr := rateLimiter.ReleaseToken(taskCtx, taskOptions.TaskID, taskOptions.TraceID); releaseErr != nil {
			logEntry.Errorf("Failed to release algorithm execution token on job success: %v", releaseErr)
		} else {
			logEntry.Info("Successfully released algorithm execution token on job success")
		}

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

		if err := repository.UpdateExecutionResult(options.ExecutionID, map[string]any{
			"status": consts.ExecutionSuccess,
		}); err != nil {
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
				consts.CollectTimestamp:   options.Timestamp,
			},
			Immediate: true,
			TraceID:   taskOptions.TraceID,
			GroupID:   taskOptions.GroupID,
			ProjectID: taskOptions.ProjectID,
		}
		task.SetTraceCtx(traceCtx)

		if _, _, err := SubmitTask(taskCtx, task); err != nil {
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

func parseTaskIdentifiers(message string, labels map[string]string) (*TaskIdentifiers, error) {
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

	var projectID *int
	projectIDStr, ok := labels[consts.LabelProjectID]
	if ok && projectIDStr != "" {
		id, err := strconv.Atoi(projectIDStr)
		if err != nil {
			return nil, fmt.Errorf(message, consts.LabelProjectID)
		}
		projectID = &id
	}

	var userID *int
	userIDStr, ok := labels[consts.LabelUserID]
	if ok && userIDStr != "" {
		id, err := strconv.Atoi(userIDStr)
		if err != nil {
			return nil, fmt.Errorf(message, consts.LabelUserID)
		}
		userID = &id
	}

	return &TaskIdentifiers{
		TaskID:    taskID,
		TraceID:   traceID,
		GroupID:   groupID,
		ProjectID: projectID,
		UserID:    userID,
	}, nil
}

func parseCRDLabels(labels map[string]string) (*CRDLabels, error) {
	message := "missing or invalid '%s' key in k8s CRD labels"

	identifiers, err := parseTaskIdentifiers(message, labels)
	if err != nil {
		return nil, err
	}

	benchmark, ok := labels[consts.LabelBenchmark]
	if !ok || benchmark == "" {
		return nil, fmt.Errorf(message, consts.LabelBenchmark)
	}

	var preDuration int
	preDurationStr, ok := labels[consts.LabelPreDuration]
	if ok && preDurationStr != "" {
		duration, err := strconv.Atoi(preDurationStr)
		if err != nil {
			return nil, fmt.Errorf(message, consts.LabelPreDuration)
		}

		preDuration = duration
	}

	return &CRDLabels{
		TaskIdentifiers: *identifiers,
		Benchmark:       benchmark,
		PreDuration:     preDuration,
	}, nil
}

func parseTaskOptions(labels map[string]string) (*TaskOptions, error) {
	message := "missing or invalid '%s' key in k8s job labels"

	identifiers, err := parseTaskIdentifiers(message, labels)
	if err != nil {
		return nil, err
	}

	taskType, ok := labels[consts.LabelTaskType]
	if !ok || taskType == "" {
		return nil, fmt.Errorf(message, consts.LabelTaskType)
	}

	return &TaskOptions{
		TaskIdentifiers: *identifiers,
		Type:            consts.TaskType(taskType),
	}, nil
}

func parseDatasetOptions(labels map[string]string) (*dto.DatasetOptions, error) {
	message := "missing or invalid '%s' key in job labels"

	dataset, ok := labels[consts.LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.LabelDataset)
	}

	return &dto.DatasetOptions{
		Dataset: dataset,
	}, nil
}

func parseExecutionOptions(annotations, labels map[string]string) (*dto.ExecutionOptions, error) {
	message := "missing or invalid '%s' key in job labels"

	algorithmStr, ok := annotations[consts.AnnotationAlgorithm]
	if !ok || algorithmStr == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in job annotations", consts.AnnotationAlgorithm)
	}

	var algorithm dto.AlgorithmItem
	if err := json.Unmarshal([]byte(algorithmStr), &algorithm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal '%s' to AlgorithmItem: %v", algorithmStr, err)
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

	timestamp, ok := labels[consts.LabelTimestamp]
	if !ok || timestamp == "" {
		return nil, fmt.Errorf(message, consts.LabelTimestamp)
	}

	return &dto.ExecutionOptions{
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
		Timestamp:   timestamp,
	}, nil
}

func writeJobLogs(job *batchv1.Job, taskOptions *TaskOptions, logMap map[string][]string) (string, error) {
	logWriter, err := utils.NewJobLogWriter()
	if err != nil {
		return "", fmt.Errorf("failed to create job log writer: %v", err)
	}

	filePath, err := logWriter.WriteJobLogs(
		job.Name,
		job.Namespace,
		taskOptions.TraceID,
		logMap,
	)
	if err != nil {
		return "", fmt.Errorf("failed to write job logs to file: %v", err)
	}

	return filePath, nil
}
