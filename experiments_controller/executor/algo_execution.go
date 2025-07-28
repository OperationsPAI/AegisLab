package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
)

type executionPayload struct {
	algorithm dto.AlgorithmItem
	dataset   string
	envVars   map[string]string
}

func rescheduleAlgoExecutionTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	span := trace.SpanFromContext(ctx)

	var executeTime time.Time

	// 实现随机 1 到 5 分钟的延迟
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

		record, err := repository.GetDatasetByName(payload.dataset, consts.DatasetBuildSuccess)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to query database for dataset")
			return fmt.Errorf("failed to query database for dataset %s: %v", payload.dataset, err)
		}

		container, err := repository.GetContaineInfo(&dto.GetContainerFilterOptions{
			Type:  consts.ContainerTypeAlgorithm,
			Name:  payload.algorithm.Name,
			Image: payload.algorithm.Image,
			Tag:   payload.algorithm.Tag,
		})
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get container info for algorithm")
			return fmt.Errorf("failed to get container info for algorithm %s: %v", payload.algorithm.Name, err)
		}

		executionID, err := repository.CreateExecutionResult(task.TaskID, container.ID, record.ID)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create execution result")
			return fmt.Errorf("failed to create execution result: %v", err)
		}

		itemJson, err := json.Marshal(payload.algorithm)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to marshal algorithm item")
			return fmt.Errorf("failed to marshal algorithm item: %v", err)
		}

		annotations[consts.AnnotationAlgorithm] = string(itemJson)

		jobName := task.TaskID
		fullImage := fmt.Sprintf("%s:%s", container.Image, container.Tag)
		labels := map[string]string{
			consts.LabelTaskID:      task.TaskID,
			consts.LabelTraceID:     task.TraceID,
			consts.LabelGroupID:     task.GroupID,
			consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
			consts.LabelDataset:     payload.dataset,
			consts.LabelExecutionID: strconv.Itoa(executionID),
		}

		return createAlgoJob(childCtx, config.GetString("k8s.namespace"), jobName, fullImage, annotations, labels, payload, container, record)
	})
}

// 解析算法执行任务的 Payload
func parseExecutionPayload(payload map[string]any) (*executionPayload, error) {
	message := "missing or invalid '%s' key in payload"

	algorithm, err := utils.ConvertToType[dto.AlgorithmItem](payload[consts.ExecuteAlgorithm])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to AlgorithmItem: %v", consts.ExecuteAlgorithm, err)
	}

	dataset, ok := payload[consts.ExecuteDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.ExecuteDataset)
	}

	envVars, err := utils.ConvertToType[map[string]string](payload[consts.ExecuteEnvVars])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to map[string]string: %v", consts.ExecuteEnvVars, err)
	}

	return &executionPayload{
		algorithm: algorithm,
		dataset:   dataset,
		envVars:   envVars,
	}, nil
}

func createAlgoJob(ctx context.Context, jobNamespace, jobName, image string, annotations map[string]string, labels map[string]string, payload *executionPayload, container *database.Container, record *dto.DatasetItemWithID) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		restartPolicy := corev1.RestartPolicyNever
		backoffLimit := int32(2)
		parallelism := int32(1)
		completions := int32(1)

		jobEnvVars, err := getAlgoJobEnvVars(payload, container, record)
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

		return k8s.CreateJob(ctx, k8s.JobConfig{
			Namespace:      jobNamespace,
			JobName:        jobName,
			Image:          image,
			Command:        strings.Split(container.Command, " "),
			RestartPolicy:  restartPolicy,
			BackoffLimit:   backoffLimit,
			Parallelism:    parallelism,
			Completions:    completions,
			Annotations:    annotations,
			Labels:         labels,
			EnvVars:        jobEnvVars,
			InitContainers: initContainers,
		})
	})
}

func getAlgoJobEnvVars(payload *executionPayload, container *database.Container, record *dto.DatasetItemWithID) ([]corev1.EnvVar, error) {
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
	if container.Name != config.GetString("algo.detector") {
		outputPath = fmt.Sprintf("/data/%s/%s/%s", payload.dataset, container.Name, timestamp)
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "TIMEZONE", Value: tz},
		{Name: "TIMESTAMP", Value: timestamp},
		{Name: "NORMAL_START", Value: strconv.FormatInt(record.StartTime.Add(-time.Duration(preDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(record.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.dataset)},
		{Name: "OUTPUT_PATH", Value: outputPath},
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
		if container.EnvVars != "" {
			envVarsArray := strings.Split(container.EnvVars, ",")
			for _, envVar := range envVarsArray {
				if _, exists := extraEnvVarMap[envVar]; !exists {
					return nil, fmt.Errorf("environment variable %s is required but not provided in algorithm exeuciton payload", envVar)
				}
			}
		}
	} else {
		if container.EnvVars != "" {
			return nil, fmt.Errorf("environment variables %s are required but not provided in algorithm execution payload", container.EnvVars)
		}
	}

	return jobEnvVars, nil
}
