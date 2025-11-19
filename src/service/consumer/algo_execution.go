package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/tracing"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
)

type executionPayload struct {
	algorithm        dto.ContainerVersionItem
	datapack         dto.InjectionItem
	datasetVersionID int
	labels           []dto.LabelItem
}

type algoJobCreationParams struct {
	jobName     string
	image       string
	annotations map[string]string
	labels      map[string]string
	executionID int
	payload     *executionPayload
}

func (p *algoJobCreationParams) toK8sJobConfig(envVars []corev1.EnvVar, initContainers []corev1.Container) *k8s.JobConfig {
	return &k8s.JobConfig{
		JobName:        p.jobName,
		Image:          p.image,
		Command:        strings.Split(p.payload.algorithm.Command, " "),
		EnvVars:        envVars,
		Annotations:    p.annotations,
		Labels:         p.labels,
		InitContainers: initContainers,
	}
}

// executeAlgorithm handles the execution of an algorithm task
func executeAlgorithm(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting algorithm execution attempt %d", task.ReStartNum+1))
		logEntry := logrus.WithFields(logrus.Fields{
			"task_id":  task.TaskID,
			"trace_id": task.TraceID,
		})

		rateLimiter := GetAlgoExecutionRateLimiter()
		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to acquire rate limit token", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logEntry.Info("No algorithm execution token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				return handleExecutionError(span, logEntry, "failed to wait for token", err)
			}

			if !acquired {
				if err := rescheduleAlgoExecutionTask(childCtx, task, "failed to acquire algorithm execution token within timeout, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		payload, err := parseExecutionPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse execution payload", err)
		}

		executionID, err := createExecution(payload.algorithm.ID, payload.datapack.ID, payload.datasetVersionID, payload.labels)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to create execution result", err)
		}

		annotations, err := task.GetAnnotations(childCtx)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get annotations", err)
		}

		itemJson, err := json.Marshal(payload.algorithm)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal algorithm item", err)
		}
		annotations[consts.JobAnnotationAlgorithm] = string(itemJson)

		jobName := task.TaskID
		jobLabels := utils.MergeSimpleMaps(
			task.GetLabels(),
			map[string]string{
				consts.JobLabelDatapack:    payload.datapack.Name,
				consts.JobLabelExecutionID: strconv.Itoa(executionID),
			},
		)

		params := &algoJobCreationParams{
			jobName:     jobName,
			image:       payload.algorithm.ImageRef,
			annotations: annotations,
			labels:      jobLabels,
			executionID: executionID,
			payload:     payload,
		}
		if err := createAlgoJob(childCtx, params); err != nil {
			return err
		}

		return nil
	})
}

// rescheduleAlgoExecutionTask reschedules a algorithm execution task with a random delay between 1 to 5 minutes
func rescheduleAlgoExecutionTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		randomDelayMinutes := minDelayMinutes + rand.Intn(maxDelayMinutes-minDelayMinutes+1)
		executeTime := time.Now().Add(time.Duration(randomDelayMinutes) * time.Minute)

		span.AddEvent(fmt.Sprintf("rescheduling algorithm execution task: %s", reason))
		logrus.WithFields(logrus.Fields{
			"task_id":     task.TaskID,
			"trace_id":    task.TraceID,
			"delay_mins":  randomDelayMinutes,
			"retry_count": task.ReStartNum + 1,
		}).Warnf("%s: scheduled for %s", reason, executeTime.Format(time.DateTime))

		tracing.SetSpanAttribute(childCtx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskPending))

		publishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRunAlgorithm,
			EventName: consts.EventNoTokenAvailable,
			Payload:   executeTime.String(),
		})

		updateTaskState(
			childCtx,
			task.TraceID,
			task.TaskID,
			reason,
			consts.TaskRescheduled,
			consts.TaskTypeRunAlgorithm,
		)

		task.Reschedule(executeTime)
		if err := common.SubmitTask(childCtx, task); err != nil {
			span.RecordError(err)
			span.AddEvent("failed to submit rescheduled task")
			return fmt.Errorf("failed to submit rescheduled algorithm execution task: %w", err)
		}

		return nil
	})
}

// parseExecutionPayload extracts and validates the execution payload from the task
func parseExecutionPayload(payload map[string]any) (*executionPayload, error) {
	algorithmVersion, err := utils.ConvertToType[dto.ContainerVersionItem](payload[consts.ExecuteAlgorithm])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to ContainerVersionItem: %w", consts.ExecuteAlgorithm, err)
	}

	datapack, err := utils.ConvertToType[dto.InjectionItem](payload[consts.ExecuteDatapack])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to InjectionItem: %w", consts.ExecuteDatapack, err)
	}

	datasetVersionIDFloat, ok := payload[consts.ExecuteDatasetVersionID].(float64)
	if !ok || datasetVersionIDFloat < consts.DefaultInvalidID {
		return nil, fmt.Errorf("missing or invalid '%s' in execution payload: %w", consts.ExecuteDatasetVersionID, err)
	}
	datasetVersionID := int(datasetVersionIDFloat)

	labels, err := utils.ConvertToType[[]dto.LabelItem](payload[consts.ExecuteLabels])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to []LabelItem: %w", consts.ExecuteLabels, err)
	}

	return &executionPayload{
		algorithm:        algorithmVersion,
		datapack:         datapack,
		datasetVersionID: datasetVersionID,
		labels:           labels,
	}, nil
}

// createAlgoJob creates and submits a Kubernetes job for algorithm execution
func createAlgoJob(ctx context.Context, params *algoJobCreationParams) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		logEntry := logrus.WithFields(logrus.Fields{
			"job_name":     params.jobName,
			"execution_id": params.executionID,
		})

		jobEnvVars, err := getAlgoJobEnvVars(params.executionID, params.payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get job environment variables", err)
		}

		envVarMap := make(map[string]string, len(jobEnvVars))
		for _, envVar := range jobEnvVars {
			envVarMap[envVar.Name] = envVar.Value
		}

		outputPath := envVarMap["OUTPUT_PATH"]
		params.labels["timestamp"] = envVarMap["TIMESTAMP"]

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

		return k8s.CreateJob(childCtx, params.toK8sJobConfig(jobEnvVars, initContainers))
	})
}

// getAlgoJobEnvVars constructs the environment variables for the algorithm job
func getAlgoJobEnvVars(executionID int, payload *executionPayload) ([]corev1.EnvVar, error) {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	now := time.Now()
	timestamp := now.Format(customTimeFormat)

	outputPath := filepath.Join("/data", payload.datapack.Name)
	if payload.algorithm.ContainerName != config.GetString("algo.detector") {
		outputPath = filepath.Join(outputPath, payload.algorithm.Name, timestamp)
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "ENV_MODE", Value: config.GetString("system.env_mode")},
		{Name: "TIMEZONE", Value: tz},
		{Name: "TIMESTAMP", Value: timestamp},
		{Name: "NORMAL_START", Value: strconv.FormatInt(payload.datapack.StartTime.Add(-time.Duration(payload.datapack.PreDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(payload.datapack.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(payload.datapack.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(payload.datapack.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: filepath.Join("/data", payload.datapack.Name)},
		{Name: "OUTPUT_PATH", Value: outputPath},
		{Name: "RCABENCH_USERNAME", Value: "admin"},
		{Name: "RCABENCH_PASSWORD", Value: "admin123"},
		{Name: "ALGORITHM_ID", Value: strconv.Itoa(payload.algorithm.ContainerID)},
		{Name: "ALGORITHM_VERSION_ID", Value: strconv.Itoa(payload.algorithm.ID)},
		{Name: "EXECUTION_ID", Value: strconv.Itoa(executionID)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	for _, envVar := range payload.algorithm.EnvVars {
		if _, exists := envNameIndexMap[envVar.Key]; !exists {
			if envVar.TemplateString != "" {
				logrus.Warnf("Skipping templated env var %s in algorithm version %d", envVar.Key, payload.algorithm.ID)
				continue
			}

			valueStr, ok := envVar.Value.(string)
			if !ok {
				logrus.Warnf("Skipping non-string env var %s", envVar.Key)
				continue
			}

			jobEnvVars = append(jobEnvVars, corev1.EnvVar{Name: envVar.Key, Value: valueStr})
		}
	}

	return jobEnvVars, nil
}

// createExecution creates a new execution record with associated labels
func createExecution(algorithmVersionID, datapackID int, datasetVersionID int, labelItems []dto.LabelItem) (int, error) {
	var createdExecutionID int

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		execution := &database.Execution{
			AlgorithmVersionID: algorithmVersionID,
			DatapackID:         datapackID,
			DatasetVersionID:   datasetVersionID,
			State:              consts.ExecutionInitial,
			Status:             consts.CommonEnabled,
		}

		if err := repository.CreateExecution(tx, execution); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: execution with algorithm_version_id %d and datapack_id %d already exists", consts.ErrAlreadyExists, algorithmVersionID, datapackID)
			}
			return fmt.Errorf("failed to create execution: %w", err)
		}

		if len(labelItems) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, labelItems, consts.ExecutionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}

			if err := repository.AddExecutionLabels(tx, execution.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add execution labels: %w", err)
			}
		}

		createdExecutionID = execution.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return createdExecutionID, nil
}
