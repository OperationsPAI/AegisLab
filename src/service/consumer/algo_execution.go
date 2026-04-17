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

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	k8sinfra "aegis/infra/k8s"
	redisinfra "aegis/infra/redis"
	"aegis/model"
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
	datasetVersionID *int
	labels           []dto.LabelItem
}

type algoJobCreationParams struct {
	jobName     string
	image       string
	annotations map[string]string
	labels      map[string]string
	datapackID  int
	executionID int
	payload     *executionPayload
}

func (p *algoJobCreationParams) toK8sJobConfig(envVars []corev1.EnvVar, initContainers []corev1.Container, volumeMountconfigs []k8sinfra.VolumeMountConfig) *k8sinfra.JobConfig {
	return &k8sinfra.JobConfig{
		JobName:            p.jobName,
		Image:              p.image,
		Command:            strings.Split(p.payload.algorithm.Command, " "),
		EnvVars:            envVars,
		Annotations:        p.annotations,
		Labels:             p.labels,
		InitContainers:     initContainers,
		VolumeMountConfigs: volumeMountconfigs,
		ServiceAccountName: config.GetString("k8s.job.service_account.name"),
	}
}

// executeAlgorithm handles the execution of an algorithm task
func executeAlgorithm(ctx context.Context, task *dto.UnifiedTask, deps RuntimeDeps) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting algorithm execution attempt %d", task.ReStartNum+1))
		logEntry := logrus.WithFields(logrus.Fields{
			"task_id":  task.TaskID,
			"trace_id": task.TraceID,
		})
		k8sGateway := deps.K8sGateway
		if k8sGateway == nil {
			return handleExecutionError(span, logEntry, "k8s gateway not initialized", fmt.Errorf("k8s gateway not initialized"))
		}
		redisGateway := deps.RedisGateway
		if redisGateway == nil {
			return handleExecutionError(span, logEntry, "redis gateway not initialized", fmt.Errorf("redis gateway not initialized"))
		}

		rateLimiter := deps.AlgorithmRateLimiter
		if rateLimiter == nil {
			return handleExecutionError(span, logEntry, "algorithm execution rate limiter not initialized", fmt.Errorf("algorithm execution rate limiter not initialized"))
		}
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
				if err := rescheduleAlgoExecutionTask(childCtx, deps.DB, redisGateway, task, "failed to acquire algorithm execution token within timeout, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		payload, err := parseExecutionPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse execution payload", err)
		}

		executionID, err := createExecution(deps.DB, task.TaskID, payload.algorithm.ID, payload.datapack.ID, payload.datasetVersionID, payload.labels)
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

		datapackJson, err := json.Marshal(payload.datapack)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal datapack item", err)
		}
		annotations[consts.JobAnnotationDatapack] = string(datapackJson)

		jobLabels := utils.MergeSimpleMaps(
			task.GetLabels(),
			map[string]string{
				consts.K8sLabelAppID:       consts.AppID,
				consts.JobLabelDatapack:    payload.datapack.Name,
				consts.JobLabelExecutionID: strconv.Itoa(executionID),
			},
		)

		params := &algoJobCreationParams{
			jobName:     task.TaskID,
			image:       payload.algorithm.ImageRef,
			annotations: annotations,
			labels:      jobLabels,
			datapackID:  payload.datapack.ID,
			executionID: executionID,
			payload:     payload,
		}
		if err := createAlgoJob(childCtx, k8sGateway, params); err != nil {
			return err
		}

		return nil
	})
}

// rescheduleAlgoExecutionTask reschedules a algorithm execution task with a random delay between 1 to 5 minutes
func rescheduleAlgoExecutionTask(ctx context.Context, db *gorm.DB, redisGateway *redisinfra.Gateway, task *dto.UnifiedTask, reason string) error {
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

		updateTaskState(childCtx,
			newTaskStateUpdate(
				task.TraceID,
				task.TaskID,
				consts.TaskTypeRunAlgorithm,
				consts.TaskRescheduled,
				reason,
			).withEvent(consts.EventNoTokenAvailable, executeTime.String()).withDB(db).withRedis(redisGateway),
		)

		task.Reschedule(executeTime)
		if err := common.SubmitTaskWithDB(childCtx, db, redisGateway, task); err != nil {
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

	// dataset_version_id is optional (can be 0 or missing for datapack-based executions)
	var datasetVersionID *int
	if datasetVersionIDFloat, ok := payload[consts.ExecuteDatasetVersionID].(float64); ok && datasetVersionIDFloat > consts.DefaultInvalidID {
		id := int(datasetVersionIDFloat)
		datasetVersionID = &id
	}

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
func createAlgoJob(ctx context.Context, gateway *k8sinfra.Gateway, params *algoJobCreationParams) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		logEntry := logrus.WithFields(logrus.Fields{
			"job_name":     params.jobName,
			"execution_id": params.executionID,
		})

		volumeMountConfigs, err := getRequiredVolumeMountConfigs(gateway, []consts.VolumeMountName{
			consts.VolumeMountDataset,
			consts.VolumeMountExperimentStorage,
		})
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get volume mount configurations", err)
		}

		datapackPathPrefix := volumeMountConfigs[0].MountPath
		expPathPrefix := volumeMountConfigs[1].MountPath

		jobEnvVars, err := getAlgoJobEnvVars(params.jobName, params.executionID, datapackPathPrefix, expPathPrefix, params.payload)
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
				Image:   config.GetString("k8s.init_container.busybox_image"),
				Command: []string{"sh", "-c"},
				Args: []string{
					fmt.Sprintf(`
                        mkdir -p "%s"
                        chmod 755 "%s"
                    `, outputPath, outputPath),
				},
			},
		}

		return gateway.CreateJob(childCtx, params.toK8sJobConfig(jobEnvVars, initContainers, volumeMountConfigs))
	})
}

// getAlgoJobEnvVars constructs the environment variables for the algorithm job
func getAlgoJobEnvVars(taskID string, executionID int, datapackPathPrefix, expPathPrefix string, payload *executionPayload) ([]corev1.EnvVar, error) {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = time.Local.String()
	}

	now := time.Now()
	timestamp := now.Format(customTimeFormat)

	var outputPath string
	if payload.algorithm.ContainerName == config.GetDetectorName() {
		outputPath = filepath.Join(datapackPathPrefix, payload.datapack.Name)
	} else {
		outputPath = filepath.Join(expPathPrefix, payload.algorithm.ContainerName, payload.algorithm.Name, timestamp)
	}

	// Generate service token for job authentication
	serviceToken, _, err := utils.GenerateServiceToken(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service token: %w", err)
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "TIMEZONE", Value: tz},
		{Name: "TIMESTAMP", Value: timestamp},
		{Name: "NORMAL_START", Value: strconv.FormatInt(payload.datapack.StartTime.Add(-time.Duration(payload.datapack.PreDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(payload.datapack.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(payload.datapack.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(payload.datapack.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: filepath.Join(datapackPathPrefix, payload.datapack.Name)},
		{Name: "OUTPUT_PATH", Value: outputPath},
		{Name: "RCABENCH_BASE_URL", Value: config.GetString("k8s.service.internal_url")},
		{Name: "RCABENCH_TOKEN", Value: serviceToken},
		{Name: "DATAPACK_ID", Value: strconv.Itoa(payload.datapack.ID)},
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
func createExecution(db *gorm.DB, taskID string, algorithmVersionID, datapackID int, datasetVersionID *int, labelItems []dto.LabelItem) (int, error) {
	var createdExecutionID int
	if db == nil {
		return 0, fmt.Errorf("consumer runtime db is nil")
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		execution := &model.Execution{
			TaskID:             &taskID,
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
