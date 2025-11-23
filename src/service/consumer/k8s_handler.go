package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	crdLabelsErrMsg = "missing or invalid '%s' key in k8s CRD labels"
	jobLabelsErrMsg = "missing or invalid '%s' key in k8s job labels"
)

type k8sAnnotationData struct {
	taskCarrier  propagation.MapCarrier
	traceCarrier propagation.MapCarrier

	algorithm *dto.ContainerVersionItem
	benchmark *dto.ContainerVersionItem
	datapack  *dto.InjectionItem
}

type taskIdentifiers struct {
	taskID    string
	taskType  consts.TaskType
	traceID   string
	groupID   string
	projectID int
	userID    int
}

type crdLabels struct {
	taskIdentifiers
	injectionID int
}

type jobLabels struct {
	taskIdentifiers
	ExecutionID *int
}

type K8sHandler struct {
	Monitor *Monitor
}

var Handler *K8sHandler

func NewHandler() *K8sHandler {
	return &K8sHandler{
		Monitor: GetMonitor(),
	}
}

func (h *K8sHandler) HandleCRDAdd(name string, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleCRDAdd: failed to parse annotations: %v", err)
		return
	}

	parsedLabels, err := parseCRDLabels(labels)
	if err != nil {
		logrus.Errorf("HandleCRDAdd: failed to parse CRD labels: %v", err)
		return
	}

	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)
	updateTaskState(
		taskCtx,
		parsedLabels.traceID,
		parsedLabels.taskID,
		"injecting",
		consts.TaskRunning,
		consts.TaskTypeFaultInjection,
	)

	if err := updateInjectionName(parsedLabels.injectionID, name); err != nil {
		handleTolerableError(trace.SpanFromContext(taskCtx), logrus.WithField("injection_id", parsedLabels.injectionID), "update injection name failed", err)
	}
}

func (h *K8sHandler) HandleCRDDelete(namespace string, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleCRDDelete: failed to parse annotations: %v", err)
		return
	}

	parsedLabels, err := parseCRDLabels(labels)
	if err != nil {
		logrus.Errorf("HandleCRDDelete: failed to parse CRD labels: %v", err)
		return
	}

	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)
	if err := GetMonitor().ReleaseLock(taskCtx, namespace, parsedLabels.traceID); err != nil {
		logrus.Errorf("failed to release lock for namespace %s: %v", namespace, err)
	}
}

func (h *K8sHandler) HandleCRDFailed(name string, annotations map[string]string, labels map[string]string, errMsg string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleCRDFailed: failed to parse annotations: %v", err)
		return
	}

	parsedLabels, err := parseCRDLabels(labels)
	if err != nil {
		logrus.Errorf("HandleCRDFailed: failed to parse CRD labels: %v", err)
		return
	}

	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  parsedLabels.taskID,
		"trace_id": parsedLabels.traceID,
	})
	taskSpan := trace.SpanFromContext(taskCtx)

	if err := updateInjectionState(name, consts.DatapackInjectFailed); err != nil {
		handleTolerableError(taskSpan, logEntry, "update injection state failed", err)
	}

	publishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.traceID), dto.StreamEvent{
		TaskID:    parsedLabels.taskID,
		TaskType:  consts.TaskTypeFaultInjection,
		EventName: consts.EventFaultInjectionFailed,
		Payload: dto.InfoPayloadTemplate{
			State: consts.GetTaskStateName(consts.TaskError),
			Msg:   errMsg,
		},
	})

	updateTaskState(
		taskCtx,
		parsedLabels.traceID,
		parsedLabels.taskID,
		errMsg,
		consts.TaskError,
		consts.TaskTypeFaultInjection,
	)
}

func (h *K8sHandler) HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleCRDSucceeded: failed to parse annotations: %v", err)
		return
	}

	parsedLabels, err := parseCRDLabels(labels)
	if err != nil {
		logrus.Errorf("HandleCRDSucceeded: failed to parse CRD labels: %v", err)
		return
	}

	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)
	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.traceCarrier)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  parsedLabels.taskID,
		"trace_id": parsedLabels.traceID,
	})
	taskSpan := trace.SpanFromContext(taskCtx)

	if err := updateInjectionState(name, consts.DatapackInjectSuccess); err != nil {
		handleTolerableError(taskSpan, logEntry, "update injection state failed", err)
	}

	logEntry.Info("fault injected successfully")
	taskSpan.AddEvent("fault injected successfully")
	publishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.traceID), dto.StreamEvent{
		TaskID:    parsedLabels.taskID,
		TaskType:  consts.TaskTypeFaultInjection,
		EventName: consts.EventFaultInjectionCompleted,
	}, withCallerLevel(4))

	updateTaskState(
		taskCtx,
		parsedLabels.traceID,
		parsedLabels.taskID,
		fmt.Sprintf(consts.TaskMsgCompleted, parsedLabels.taskID),
		consts.TaskCompleted,
		consts.TaskTypeFaultInjection,
	)

	datapack, err := updateInjectionTimestamp(name, startTime, endTime)
	if err != nil {
		logEntry.Errorf("update injection timestamps failed: %v", err)
		taskSpan.AddEvent("update injection timestamps failed")
		taskSpan.RecordError(err)
		return
	}

	payload := map[string]any{
		consts.BuildBenchmark:        *parsedAnnotations.benchmark,
		consts.BuildDatapack:         *datapack,
		consts.BuildDatasetVersionID: consts.DefaultInvalidID,
		consts.InjectNamespace:       namespace,
	}

	task := &dto.UnifiedTask{
		Type:      consts.TaskTypeBuildDatapack,
		Immediate: true,
		Payload:   payload,
		TraceID:   parsedLabels.traceID,
		GroupID:   parsedLabels.groupID,
		ProjectID: parsedLabels.projectID,
		UserID:    parsedLabels.userID,
		State:     consts.TaskPending,
	}
	task.SetTraceCtx(traceCtx)

	err = common.SubmitTask(taskCtx, task)
	if err != nil {
		logEntry.Errorf("submit dataset building task failed: %v", err)
		taskSpan.AddEvent("submit dataset building task failed")
		taskSpan.RecordError(err)
	}
}

func (h *K8sHandler) HandleJobAdd(annotations map[string]string, labels map[string]string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleJobAdd: failed to parse annotations: %v", err)
		return
	}

	parsedLabels, err := parseJobLabels(labels)
	if err != nil {
		logrus.Errorf("HandleJobAdd: failed to parse job labels: %v", err)
		return
	}

	var message string
	switch parsedLabels.taskType {
	case consts.TaskTypeBuildDatapack:
		message = fmt.Sprintf("building dataset for task %s", parsedLabels.taskID)
	case consts.TaskTypeRunAlgorithm:
		message = fmt.Sprintf("running algorithm for task %s", parsedLabels.taskID)
	}

	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)
	updateTaskState(
		taskCtx,
		parsedLabels.traceID,
		parsedLabels.taskID,
		message,
		consts.TaskRunning,
		parsedLabels.taskType,
	)
}

func (h *K8sHandler) HandleJobFailed(job *batchv1.Job, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleJobFailed: parse annotations failed: %v", err)
		return
	}

	parsedLabels, err := parseJobLabels(labels)
	if err != nil {
		logrus.Errorf("HandleJobFailed: parse job labels failed: %v", err)
		return
	}

	stream := fmt.Sprintf(consts.StreamLogKey, parsedLabels.traceID)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  parsedLabels.taskID,
		"trace_id": parsedLabels.traceID,
	})
	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)
	taskSpan := trace.SpanFromContext(taskCtx)

	if parsedAnnotations.datapack == nil {
		handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "missing datapack information in annotations", parsedLabels.taskType, nil)
		return
	}

	logMap, err := k8s.GetJobPodLogs(taskCtx, job.Namespace, job.Name)
	if err != nil {
		handleTolerableError(taskSpan, logrus.WithField("job_name", job.Name), "failed to get job logs", err)
	}

	var filePath string
	if len(logMap) > 0 {
		jobLogBytes, err := json.Marshal(logMap)
		if err != nil {
			handleTolerableError(taskSpan, logrus.WithField("job_name", job.Name), "failed to marshal job logs", err)
		}

		jobLog := string(jobLogBytes)
		spanAttrs := []trace.EventOption{
			trace.WithAttributes(
				attribute.String("job_name", job.Name),
				attribute.String("namespace", job.Namespace),
			),
		}

		filePath, err = writeJobLogs(job, parsedLabels.traceID, logMap)
		if err != nil {
			handleTolerableError(taskSpan, logrus.WithFields(logrus.Fields{
				"job_name":  job.Name,
				"pod_count": len(logMap),
			}), "job failed - logs available but file writing failed", err)

			for podName, log := range logMap {
				logrus.WithField("pod_name", podName).Warn("job logs:")
				logrus.Warn(log)
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

		taskSpan.AddEvent("job failed", spanAttrs...)
	}

	publishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.traceID), dto.StreamEvent{
		TaskID:    parsedLabels.taskID,
		TaskType:  parsedLabels.taskType,
		EventName: consts.EventJobFailed,
		Payload: dto.JobMessage{
			JobName:   job.Name,
			Namespace: job.Namespace,
			LogFile:   filePath,
		},
	}, withCallerLevel(4))

	switch parsedLabels.taskType {
	case consts.TaskTypeBuildDatapack:
		logEntry.Error("datapack build failed")
		taskSpan.AddEvent("datapack build failed")
		publishEvent(taskCtx, stream, dto.StreamEvent{
			TaskID:    parsedLabels.taskID,
			TaskType:  parsedLabels.taskType,
			EventName: consts.EventDatapackBuildFailed,
			Payload: dto.DatapackResult{
				Datapack:  parsedAnnotations.datapack,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}, withCallerLevel(4))

		if err := updateInjectionState(parsedAnnotations.datapack.Name, consts.DatapackBuildFailed); err != nil {
			handleTolerableError(taskSpan, logEntry, "update injection state failed", err)
		}

	case consts.TaskTypeRunAlgorithm:
		rateLimiter := GetAlgoExecutionRateLimiter()
		if releaseErr := rateLimiter.ReleaseToken(taskCtx, parsedLabels.taskID, parsedLabels.traceID); releaseErr != nil {
			handleTolerableError(taskSpan, logEntry, "failed to release algorithm execution token on job failure", releaseErr)
		} else {
			logEntry.Info("successfully released algorithm execution token on job failure")
			taskSpan.AddEvent("successfully released algorithm execution token on job failure")
		}

		if parsedAnnotations.algorithm == nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "missing algorithm information in annotations", parsedLabels.taskType, nil)
			return
		}

		if parsedLabels.ExecutionID == nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "missing execution ID in job labels", parsedLabels.taskType, nil)
			return
		}

		logEntry.Error("algorithm execute failed")
		taskSpan.AddEvent("algorithm execute failed")
		publishEvent(taskCtx, stream, dto.StreamEvent{
			TaskID:    parsedLabels.taskID,
			TaskType:  parsedLabels.taskType,
			EventName: consts.EventAlgoRunFailed,
			Payload: dto.ExecutionResult{
				Algorithm:   parsedAnnotations.algorithm,
				Datapack:    parsedAnnotations.datapack,
				ExecutionID: *parsedLabels.ExecutionID,
				Timestamp:   time.Now().Format(time.RFC3339),
			},
		}, withCallerLevel(4))

		if err := updateExecutionState(*parsedLabels.ExecutionID, consts.ExecutionFailed); err != nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "update execution state failed", parsedLabels.taskType, err)
			return
		}
	}

	updateTaskState(
		taskCtx,
		parsedLabels.traceID,
		parsedLabels.taskID,
		fmt.Sprintf(consts.TaskMsgFailed, parsedLabels.taskID),
		consts.TaskError,
		parsedLabels.taskType,
	)
}

func (h *K8sHandler) HandleJobSucceeded(job *batchv1.Job, annotations map[string]string, labels map[string]string) {
	parsedAnnotations, err := parseAnnotations(annotations)
	if err != nil {
		logrus.Errorf("HandleJobSucceeded: failed to parse annotations: %v", err)
		return
	}

	parsedLabels, err := parseJobLabels(labels)
	if err != nil {
		logrus.Errorf("HandleJobSucceeded: failed to parse job labels: %v", err)
		return
	}

	stream := fmt.Sprintf(consts.StreamLogKey, parsedLabels.traceID)

	taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.taskCarrier)
	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), parsedAnnotations.traceCarrier)

	logEntry := logrus.WithFields(logrus.Fields{
		"task_id":  parsedLabels.taskID,
		"trace_id": parsedLabels.traceID,
	})
	taskSpan := trace.SpanFromContext(taskCtx)

	if parsedAnnotations.datapack == nil {
		handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "missing datapack information in annotations", parsedLabels.taskType, nil)
		return
	}

	publishEvent(taskCtx, stream, dto.StreamEvent{
		TaskID:    parsedLabels.taskID,
		TaskType:  parsedLabels.taskType,
		EventName: consts.EventJobSucceed,
		Payload: dto.JobMessage{
			JobName:   job.Name,
			Namespace: job.Namespace,
		},
	}, withCallerLevel(4))

	switch parsedLabels.taskType {
	case consts.TaskTypeBuildDatapack:
		logEntry.Info("datapack build successfully")
		taskSpan.AddEvent("datapack build successfully")
		publishEvent(taskCtx, stream, dto.StreamEvent{
			TaskID:    parsedLabels.taskID,
			TaskType:  parsedLabels.taskType,
			EventName: consts.EventDatapackBuildSucceed,
			Payload: dto.DatapackResult{
				Datapack:  parsedAnnotations.datapack,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}, withCallerLevel(4))

		if err := updateInjectionState(parsedAnnotations.datapack.Name, consts.DatapackBuildSuccess); err != nil {
			handleTolerableError(taskSpan, logEntry, "update dataset status failed", err)
		}

		updateTaskState(
			taskCtx,
			parsedLabels.traceID,
			parsedLabels.taskID,
			fmt.Sprintf(consts.TaskMsgCompleted, parsedLabels.taskID),
			consts.TaskCompleted,
			parsedLabels.taskType,
		)

		ref := &dto.ContainerRef{
			Name: config.GetString("algo.detector"),
		}

		algorithmVersionResults, err := common.MapRefsToContainerVersions([]*dto.ContainerRef{ref}, consts.ContainerTypeAlgorithm, parsedLabels.userID)
		if err != nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "failed to map container refs to versions", parsedLabels.taskType, err)
			return
		}
		if len(algorithmVersionResults) == 0 {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "no valid algorithm versions found", parsedLabels.taskType, nil)
			return
		}

		algorithmVersion, exists := algorithmVersionResults[ref]
		if !exists {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "algorithm version not found for item", parsedLabels.taskType, nil)
			return
		}

		payload := map[string]any{
			consts.ExecuteAlgorithm:        dto.NewContainerVersionItem(&algorithmVersion),
			consts.ExecuteDatapack:         parsedAnnotations.datapack,
			consts.ExecuteDatasetVersionID: consts.DefaultInvalidID,
		}

		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Immediate: true,
			Payload:   payload,
			TraceID:   parsedLabels.traceID,
			GroupID:   parsedLabels.groupID,
			ProjectID: parsedLabels.projectID,
			UserID:    parsedLabels.userID,
		}
		task.SetTraceCtx(traceCtx)

		if err := common.SubmitTask(taskCtx, task); err != nil {
			handleTolerableError(taskSpan, logEntry, "submit algorithm execution task failed", err)
		}

	case consts.TaskTypeRunAlgorithm:
		rateLimiter := GetAlgoExecutionRateLimiter()
		if releaseErr := rateLimiter.ReleaseToken(taskCtx, parsedLabels.taskID, parsedLabels.traceID); releaseErr != nil {
			handleTolerableError(taskSpan, logEntry, "failed to release algorithm execution token on job success", releaseErr)
		} else {
			logEntry.Info("successfully released algorithm execution token on job success")
			taskSpan.AddEvent("successfully released algorithm execution token on job success")
		}

		if parsedAnnotations.algorithm == nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "missing algorithm information in annotations", parsedLabels.taskType, nil)
			return
		}

		if parsedLabels.ExecutionID == nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "missing execution ID in job labels", parsedLabels.taskType, nil)
			return
		}

		logEntry.Info("algorithm execute successfully")
		taskSpan.AddEvent("algorithm execute successfully")
		publishEvent(taskCtx, fmt.Sprintf(consts.StreamLogKey, parsedLabels.traceID), dto.StreamEvent{
			TaskID:    parsedLabels.taskID,
			TaskType:  parsedLabels.taskType,
			EventName: consts.EventAlgoRunSucceed,
			Payload: dto.ExecutionResult{
				Algorithm:   parsedAnnotations.algorithm,
				Datapack:    parsedAnnotations.datapack,
				ExecutionID: *parsedLabels.ExecutionID,
				Timestamp:   time.Now().Format(time.RFC3339),
			},
		}, withCallerLevel(4))

		if err := updateExecutionState(*parsedLabels.ExecutionID, consts.ExecutionSuccess); err != nil {
			handleFatalError(taskCtx, taskSpan, logEntry, parsedLabels.traceID, parsedLabels.taskID, "update execution state failed", parsedLabels.taskType, err)
			return
		}

		updateTaskState(
			taskCtx,
			parsedLabels.traceID,
			parsedLabels.taskID,
			fmt.Sprintf(consts.TaskMsgCompleted, parsedLabels.taskID),
			consts.TaskCompleted,
			parsedLabels.taskType,
		)

		payload := map[string]any{
			consts.CollectAlgorithm:   parsedAnnotations.algorithm,
			consts.CollectDatapack:    parsedAnnotations.datapack,
			consts.CollectExecutionID: *parsedLabels.ExecutionID,
		}

		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeCollectResult,
			Immediate: true,
			Payload:   payload,
			TraceID:   parsedLabels.traceID,
			GroupID:   parsedLabels.groupID,
			ProjectID: parsedLabels.projectID,
			UserID:    parsedLabels.userID,
		}
		task.SetTraceCtx(traceCtx)

		if err := common.SubmitTask(taskCtx, task); err != nil {
			handleTolerableError(taskSpan, logEntry, "submit result collection task failed", err)
		}
	}
}

// handleFatalError is a helper function to handle fatal errors consistently
//
// It logs the error, adds span events, updates task state, and returns true to indicate the caller should return
func handleFatalError(ctx context.Context, span trace.Span, logEntry *logrus.Entry, traceID, taskID, message string, taskType consts.TaskType, err error) {
	if err != nil {
		logEntry.Errorf("%s: %v", message, err)
		span.RecordError(err)
	} else {
		logEntry.Error(message)
	}

	span.AddEvent(message)

	updateTaskState(
		ctx,
		traceID,
		taskID,
		message,
		consts.TaskError,
		taskType,
	)
}

// handleTolerableError is a helper function to handle tolerable errors consistently
// It logs a warning and adds span events but doesn't update task state
func handleTolerableError(span trace.Span, logEntry *logrus.Entry, message string, err error) {
	if err != nil {
		logEntry.Warnf("%s (non-fatal): %v", message, err)
		span.RecordError(err)
	} else {
		logEntry.Warn(message)
	}

	span.AddEvent(fmt.Sprintf("%s (continuing)", message))
}

func parseAnnotations(annotations map[string]string) (*k8sAnnotationData, error) {
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

	data := &k8sAnnotationData{
		taskCarrier:  taskCarrier,
		traceCarrier: traceCarrier,
	}

	if itemJson, exists := annotations[consts.CRDAnnotationBenchmark]; exists {
		var benchmark dto.ContainerVersionItem
		if err := json.Unmarshal([]byte(itemJson), &benchmark); err != nil {
			return nil, fmt.Errorf("failed to unmarshal '%s' to ContainerVersionItem: %w", consts.CRDAnnotationBenchmark, err)
		}
		data.benchmark = &benchmark
	}

	if itemJson, exists := annotations[consts.JobAnnotationAlgorithm]; exists {
		var algorithm dto.ContainerVersionItem
		if err := json.Unmarshal([]byte(itemJson), &algorithm); err != nil {
			return nil, fmt.Errorf("failed to unmarshal '%s' to ContainerVersionItem: %w", consts.JobAnnotationAlgorithm, err)
		}
		data.algorithm = &algorithm
	}

	if itemJson, exists := annotations[consts.JobAnnotationDatapack]; exists {
		var datapack dto.InjectionItem
		if err := json.Unmarshal([]byte(itemJson), &datapack); err != nil {
			return nil, fmt.Errorf("failed to unmarshal '%s' to ContainerVersionItem: %w", consts.JobAnnotationDatapack, err)
		}
		data.datapack = &datapack
	}

	return data, nil
}

func parseTaskIdentifiers(message string, labels map[string]string) (*taskIdentifiers, error) {
	taskID, ok := labels[consts.JobLabelTaskID]
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.JobLabelTaskID)
	}

	taskTypeStr, ok := labels[consts.JobLabelTaskType]
	if !ok || taskTypeStr == "" {
		return nil, fmt.Errorf(message, consts.JobLabelTaskType)
	}
	taskType := consts.GetTaskTypeByName(taskTypeStr)
	if taskType == nil {
		return nil, fmt.Errorf(message, consts.JobLabelTaskType)
	}

	traceID, ok := labels[consts.JobLabelTraceID]
	if !ok || traceID == "" {
		return nil, fmt.Errorf(message, consts.JobLabelTraceID)
	}

	groupID, ok := labels[consts.JobLabelGroupID]
	if !ok || groupID == "" {
		return nil, fmt.Errorf(message, consts.JobLabelGroupID)
	}

	projectIDStr, ok := labels[consts.JobLabelProjectID]
	if !ok || projectIDStr == "" {
		return nil, fmt.Errorf(message, consts.JobLabelGroupID)
	}
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		return nil, fmt.Errorf(message, consts.JobLabelProjectID)
	}

	userIDStr, ok := labels[consts.JobLabelUserID]
	if !ok || userIDStr == "" {
		return nil, fmt.Errorf(message, consts.JobLabelUserID)
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, fmt.Errorf(message, consts.JobLabelUserID)
	}

	return &taskIdentifiers{
		taskID:    taskID,
		taskType:  *taskType,
		traceID:   traceID,
		groupID:   groupID,
		projectID: projectID,
		userID:    userID,
	}, nil
}

func parseCRDLabels(labels map[string]string) (*crdLabels, error) {
	identifiers, err := parseTaskIdentifiers(crdLabelsErrMsg, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task identifiers: %w", err)
	}

	injectionIDStr, ok := labels[consts.CRDLabelInjectionID]
	if !ok || injectionIDStr == "" {
		return nil, fmt.Errorf(crdLabelsErrMsg, consts.CRDLabelInjectionID)
	}
	injectionID, err := strconv.Atoi(injectionIDStr)
	if err != nil {
		return nil, fmt.Errorf(crdLabelsErrMsg, consts.CRDLabelInjectionID)
	}

	return &crdLabels{
		taskIdentifiers: *identifiers,
		injectionID:     injectionID,
	}, nil
}

func parseJobLabels(labels map[string]string) (*jobLabels, error) {
	identifiers, err := parseTaskIdentifiers(jobLabelsErrMsg, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task identifiers: %w", err)
	}

	data := &jobLabels{
		taskIdentifiers: *identifiers,
	}

	if executionIDStr, exists := labels[consts.JobLabelExecutionID]; exists {
		executionID, err := strconv.Atoi(executionIDStr)
		if err != nil {
			return nil, fmt.Errorf(jobLabelsErrMsg, consts.JobLabelExecutionID)
		}
		data.ExecutionID = &executionID
	}

	return data, nil
}

// updateInjectionName updates the name of a fault injection
func updateInjectionName(injectionID int, newName string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByID(tx, injectionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("injection %d not found", injectionID)
			}
			return fmt.Errorf("failed to get injection %d: %w", injectionID, err)
		}

		if err = repository.UpdateInjection(tx, injection.ID, map[string]any{
			"name": newName,
		}); err != nil {
			return fmt.Errorf("update injection timestamps failed: %w", err)
		}

		return nil
	})
}

// updateExecutionState updates the state of an execution
func updateExecutionState(executionID int, newState consts.ExecutionState) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		execution, err := repository.GetExecutionByID(tx, executionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: execution %d not found", consts.ErrNotFound, executionID)
			}
			return fmt.Errorf("execution %d not found: %w", executionID, err)
		}

		if execution.State != consts.ExecutionInitial {
			return fmt.Errorf("cannot change state of execution %d from %s to %s", executionID, consts.GetExecuteStateName(execution.State), consts.GetExecuteStateName(newState))
		}

		if err := repository.UpdateExecution(tx, executionID, map[string]any{
			"state": newState,
		}); err != nil {
			return fmt.Errorf("failed to update execution %d duration: %w", executionID, err)
		}

		return nil
	})
}

// updateInjectionState updates the state of a fault injection
func updateInjectionState(injectionName string, newState consts.DatapackState) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByName(tx, injectionName)
		if err != nil {
			return fmt.Errorf("failed to get injection %s: %w", injectionName, err)
		}

		if injection.State != consts.DatapackInitial {
			return fmt.Errorf("cannot change state of injection %s from %s to %s", injectionName, consts.GetDatapackStateName(injection.State), consts.GetDatapackStateName(newState))
		}

		if err := repository.UpdateInjection(tx, injection.ID, map[string]any{
			"state": newState,
		}); err != nil {
			return fmt.Errorf("failed to update injection %s state: %w", injectionName, err)
		}

		return nil
	})
}

// updateInjectionTimestamp updates the start and end timestamps of a fault injection
func updateInjectionTimestamp(injectionName string, startTime time.Time, endTime time.Time) (*dto.InjectionItem, error) {
	var updatedInjection *database.FaultInjection
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByName(tx, injectionName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("injection %s not found", injectionName)
			}
			return fmt.Errorf("failed to get injection %s: %w", injectionName, err)
		}

		if err = repository.UpdateInjection(tx, injection.ID, map[string]any{
			"start_time": startTime,
			"end_time":   endTime,
		}); err != nil {
			return fmt.Errorf("update injection timestamps failed: %w", err)
		}

		reloadedInjection, err := repository.GetInjectionByID(tx, injection.ID)
		if err != nil {
			return fmt.Errorf("failed to reload injection %d after update: %w", injection.ID, err)
		}

		updatedInjection = reloadedInjection
		return nil
	})
	if err != nil {
		return nil, err
	}

	injectionItem := dto.NewInjectionItem(updatedInjection)
	return &injectionItem, err
}

func writeJobLogs(job *batchv1.Job, traceID string, logMap map[string][]string) (string, error) {
	logWriter, err := utils.NewJobLogWriter()
	if err != nil {
		return "", fmt.Errorf("failed to create job log writer: %w", err)
	}

	filePath, err := logWriter.WriteJobLogs(job.Name, job.Namespace, traceID, logMap)
	if err != nil {
		return "", fmt.Errorf("failed to write job logs to file: %w", err)
	}

	return filePath, nil
}
