package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/tracing"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
)

type executionPayload struct {
	algorithmVersion database.ContainerVersion
	dataset          string
	envVars          map[string]string
	labels           *dto.ExecutionLabels
}

func rescheduleAlgoExecutionTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	span := trace.SpanFromContext(ctx)

	var executeTime time.Time

	// Implement random 1 to 5 minute delay
	minDelayMinutes := 1
	maxDelayMinutes := 5
	randomDelayMinutes := minDelayMinutes + rand.Intn(maxDelayMinutes-minDelayMinutes+1)
	executeTime = time.Now().Add(time.Duration(randomDelayMinutes) * time.Minute)

	eventPayload := executeTime.String()

	span.AddEvent(fmt.Sprintf("rescheduling algorithm execution task: %s", reason))
	logrus.WithFields(logrus.Fields{
		"task_id":     task.TaskID,
		"trace_id":    task.TraceID,
		"delay_mins":  randomDelayMinutes,
		"retry_count": task.ReStartNum + 1,
	}).Warnf("%s: scheduled for %s", reason, executeTime.Format("2006-01-02 15:04:05"))

	tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusPending))

	repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
		TaskID:    task.TaskID,
		TaskType:  consts.TaskTypeRunAlgorithm,
		EventName: consts.EventNoTokenAvailable,
		Payload:   eventPayload,
	})

	updateTaskStatus(
		ctx,
		task.TraceID,
		task.TaskID,
		reason,
		consts.TaskStautsRescheduled,
		consts.TaskTypeRunAlgorithm,
	)

	if _, _, err := SubmitTask(ctx, &dto.UnifiedTask{
		TaskID:       task.TaskID,
		Type:         consts.TaskTypeRunAlgorithm,
		Immediate:    false,
		ExecuteTime:  executeTime.Unix(),
		ReStartNum:   task.ReStartNum + 1,
		Payload:      task.Payload,
		Status:       consts.TaskStautsRescheduled,
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		TraceCarrier: task.TraceCarrier,
	}); err != nil {
		span.RecordError(err)
		span.AddEvent("failed to submit rescheduled task")
		return fmt.Errorf("failed to submit rescheduled algorithm execution task: %v", err)
	}

	return nil
}

func executeAlgorithm(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting algorithm execution attempt %d", task.ReStartNum+1))

		rateLimiter := utils.NewAlgoExecutionRateLimiter()

		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to acquire rate limit token")
			return fmt.Errorf("failed to acquire rate limit token: %v", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logrus.WithFields(logrus.Fields{
				"task_id":  task.TaskID,
				"trace_id": task.TraceID,
			}).Info("No algorithm execution token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				span.RecordError(err)
				span.AddEvent("failed to wait for token")
				return fmt.Errorf("failed to wait for token: %v", err)
			}

			if !acquired {
				if err := rescheduleAlgoExecutionTask(childCtx, task, "failed to acquire algorithm execution token within timeout, retrying later"); err != nil {
					return err
				}

				return nil
			}
		}

		// Ensure token is released when function exits
		var tokenAcquired = true
		defer func() {
			if tokenAcquired {
				if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
					logrus.WithFields(logrus.Fields{
						"task_id":  task.TaskID,
						"trace_id": task.TraceID,
						"error":    releaseErr,
					}).Error("Failed to release algorithm execution token")
				}
			}
		}()

		// Note: Token will be released when job completes or fails, not here
		// This ensures proper rate limiting during the entire job execution period
		payload, err := parseExecutionPayload(task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse execution payload")
			return err
		}

		annotations, err := getAnnotations(childCtx, task)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get annotations")
			return err
		}

		record, err := repository.GetDatasetByName(payload.dataset, consts.DatapackBuildSuccess)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to query database for dataset")
			return fmt.Errorf("failed to query database for dataset %s: %v", payload.dataset, err)
		}

		executionID, err := repository.CreateExecutionResult(task.TaskID, payload.algorithmVersion.ID, record.ID, 0, payload.labels)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create execution result")
			return fmt.Errorf("failed to create execution result: %v", err)
		}

		itemJson, err := json.Marshal(dto.AlgorithmItem{
			Name:    payload.algorithmVersion.Container.Name,
			Version: payload.algorithmVersion.Name,
			EnvVars: payload.envVars,
		})
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to marshal algorithm item")
			return fmt.Errorf("failed to marshal algorithm item: %v", err)
		}

		annotations[consts.AnnotationAlgorithm] = string(itemJson)

		jobName := task.TaskID
		labels := map[string]string{
			consts.LabelTaskID:      task.TaskID,
			consts.LabelTraceID:     task.TraceID,
			consts.LabelGroupID:     task.GroupID,
			consts.LabelProjectID:   getDefaultIDString(task.ProjectID),
			consts.LabelUserID:      getDefaultIDString(task.UserID),
			consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
			consts.LabelDataset:     payload.dataset,
			consts.LabelExecutionID: strconv.Itoa(executionID),
		}

		if err := createAlgoJob(childCtx, jobName, payload.algorithmVersion.ImageRef, annotations, labels, executionID, payload, record); err != nil {
			return err
		}

		tokenAcquired = false
		return nil
	})
}

// Parse algorithm execution task Payload
func parseExecutionPayload(payload map[string]any) (*executionPayload, error) {
	message := "missing or invalid '%s' key in payload"

	algorithmVersion, err := utils.ConvertToType[database.ContainerVersion](payload[consts.ExecuteAlgorithmVersion])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to ContainerVersion: %v", consts.ExecuteAlgorithmVersion, err)
	}

	dataset, ok := payload[consts.ExecuteDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.ExecuteDataset)
	}

	envVars, err := utils.ConvertToType[map[string]string](payload[consts.ExecuteEnvVars])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to map[string]string: %v", consts.ExecuteEnvVars, err)
	}

	// Parse labels if provided
	var labels *dto.ExecutionLabels
	if labelsData, exists := payload["labels"]; exists {
		labels, err = utils.ConvertToType[*dto.ExecutionLabels](labelsData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert 'labels' to ExecutionLabels: %v", err)
		}
	}

	return &executionPayload{
		algorithmVersion: algorithmVersion,
		dataset:          dataset,
		envVars:          envVars,
		labels:           labels,
	}, nil
}

func createAlgoJob(ctx context.Context, jobName, image string, annotations, labels map[string]string, executionID int, payload *executionPayload, record *dto.DatasetItemWithID) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)

		jobEnvVars, err := getAlgoJobEnvVars(executionID, payload, &payload.algorithmVersion, record)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get job environment variables")
			return err
		}

		envVarMap := make(map[string]string, len(jobEnvVars))
		for _, envVar := range jobEnvVars {
			envVarMap[envVar.Name] = envVar.Value
		}

		outputPath := envVarMap["OUTPUT_PATH"]
		labels["timestamp"] = envVarMap["TIMESTAMP"]

		initContainers := []corev1.Container{
			{
				Name:    "create-output-dir",
				Image:   "busybox:1.35",
				Command: []string{"sh", "-c"},
				Args: []string{
					fmt.Sprintf(`
                        mkdir -p "%s"
                        chmod 755 "%s"
                    `, outputPath, outputPath),
				},
			},
		}

		return k8s.CreateJob(ctx, &k8s.JobConfig{
			JobName:        jobName,
			Image:          image,
			Command:        strings.Split(payload.algorithmVersion.Command, " "),
			Annotations:    annotations,
			Labels:         labels,
			EnvVars:        jobEnvVars,
			InitContainers: initContainers,
		})
	})
}

func getAlgoJobEnvVars(executionID int, payload *executionPayload, containerVersion *database.ContainerVersion, record *dto.DatasetItemWithID) ([]corev1.EnvVar, error) {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	preDuration, ok := record.Param["pre_duration"].(int)
	if !ok || preDuration == 0 {
		return nil, fmt.Errorf("failed to get the preduration")
	}

	now := time.Now()
	timestamp := now.Format("20060102_150405")

	outputPath := fmt.Sprintf("/data/%s", payload.dataset)
	if containerVersion.Container.Name != config.GetString("algo.detector") {
		outputPath = fmt.Sprintf("/data/%s/%s/%s", payload.dataset, containerVersion.Container.Name, timestamp)
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "ENV_MODE", Value: config.GetString("system.env_mode")},
		{Name: "TIMEZONE", Value: tz},
		{Name: "TIMESTAMP", Value: timestamp},
		{Name: "NORMAL_START", Value: strconv.FormatInt(record.StartTime.Add(-time.Duration(preDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(record.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.dataset)},
		{Name: "OUTPUT_PATH", Value: outputPath},
		{Name: "RCABENCH_USERNAME", Value: "admin"},
		{Name: "RCABENCH_PASSWORD", Value: "admin123"},
		{Name: "ALGORITHM_ID", Value: strconv.Itoa(containerVersion.ContainerID)},
		{Name: "ALGORITHM_VERSION_ID", Value: strconv.Itoa(containerVersion.ID)},
		{Name: "EXECUTION_ID", Value: strconv.Itoa(executionID)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	if payload.envVars != nil {
		extraEnvVarMap := make(map[string]struct{}, len(payload.envVars))
		for name, value := range payload.envVars {
			if index, exists := envNameIndexMap[name]; exists {
				jobEnvVars[index].Value = value
			} else {
				jobEnvVars = append(jobEnvVars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
				extraEnvVarMap[name] = struct{}{}
			}
		}

		// Check if all required environment variables are provided
		if containerVersion.EnvVars != "" {
			envVarsArray := strings.Split(containerVersion.EnvVars, ",")
			for _, envVar := range envVarsArray {
				if _, exists := extraEnvVarMap[envVar]; !exists {
					return nil, fmt.Errorf("environment variable %s is required but not provided in algorithm exeuciton payload", envVar)
				}
			}
		}
	} else {
		if containerVersion.EnvVars != "" {
			return nil, fmt.Errorf("environment variables %s are required but not provided in algorithm execution payload", containerVersion.EnvVars)
		}
	}

	return jobEnvVars, nil
}
