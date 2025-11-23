package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/tracing"
	"aegis/utils"
)

type datapackPayload struct {
	benchmark        dto.ContainerVersionItem
	datapack         dto.InjectionItem
	datasetVersionID *int
	labels           []dto.LabelItem
	namespace        string
}

type datapackJobCreationParams struct {
	jobName     string
	image       string
	annotations map[string]string
	labels      map[string]string
	payload     *datapackPayload
	logEntry    *logrus.Entry
}

func (p *datapackJobCreationParams) toK8sJobConfig(envVars []corev1.EnvVar) *k8s.JobConfig {
	return &k8s.JobConfig{
		JobName:     p.jobName,
		Image:       p.image,
		Command:     strings.Split(p.payload.benchmark.Command, " "),
		EnvVars:     envVars,
		Annotations: p.annotations,
		Labels:      p.labels,
	}
}

// executeBuildDatapack handles the execution of a datapack building task
func executeBuildDatapack(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		logEntry := logrus.WithFields(logrus.Fields{"task_id": task.TaskID, "trace_id": task.TraceID})

		payload, err := parseDatapackPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse datapack payload", err)
		}

		annotations, err := task.GetAnnotations(childCtx)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get annotations", err)
		}

		itemJson, err := json.Marshal(payload.datapack)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal datapack item", err)
		}
		annotations[consts.JobAnnotationDatapack] = string(itemJson)

		jobName := task.TaskID
		jobLabels := utils.MergeSimpleMaps(
			task.GetLabels(),
			map[string]string{
				consts.JobLabelDatapack:  payload.datapack.Name,
				consts.JobLabelDatasetID: strconv.Itoa(utils.GetIntValue(payload.datasetVersionID, 0)),
			},
		)

		params := &datapackJobCreationParams{
			jobName:     jobName,
			image:       payload.benchmark.ImageRef,
			annotations: annotations,
			labels:      jobLabels,
			payload:     payload,
			logEntry:    logEntry,
		}
		return createDatapackJob(childCtx, params)
	})
}

// parseDatapackPayload extracts and validates the datapack payload from the task
func parseDatapackPayload(payload map[string]any) (*datapackPayload, error) {
	benchmark, err := utils.ConvertToType[dto.ContainerVersionItem](payload[consts.BuildBenchmark])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to ContainerVersion: %w", consts.BuildBenchmark, err)
	}

	datapack, err := utils.ConvertToType[dto.InjectionItem](payload[consts.BuildDatapack])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to InjectionItem: %w", consts.BuildDatapack, err)
	}

	datasetID, err := utils.GetPointerIntFromMap(payload, consts.BuildDatasetVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s' from payload: %w", consts.BuildDatasetVersionID, err)
	}

	labels, err := utils.ConvertToType[[]dto.LabelItem](payload[consts.BuildLabels])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to []LabelItem: %w", consts.BuildLabels, err)
	}

	namespace, ok := payload[consts.InjectNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("missing or invalid '%s' in payload", consts.InjectNamespace)
	}

	return &datapackPayload{
		benchmark:        benchmark,
		datapack:         datapack,
		datasetVersionID: datasetID,
		labels:           labels,
		namespace:        namespace,
	}, nil
}

func createDatapackJob(ctx context.Context, params *datapackJobCreationParams) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		jobEnvVars, err := getDatapackJobEnvVars(params.payload)
		if err != nil {
			return handleExecutionError(span, params.logEntry, "failed to get job environment variables", err)
		}

		return k8s.CreateJob(childCtx, params.toK8sJobConfig(jobEnvVars))
	})
}

func getDatapackJobEnvVars(payload *datapackPayload) ([]corev1.EnvVar, error) {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	now := time.Now()
	timestamp := now.Format(customTimeFormat)

	jobEnvVars := []corev1.EnvVar{
		{Name: "ENV_MODE", Value: config.GetString("system.env_mode")},
		{Name: consts.BuildEnvVarNamespace, Value: payload.namespace},
		{Name: "TIMEZONE", Value: tz},
		{Name: "TIMESTAMP", Value: timestamp},
		{Name: "NORMAL_START", Value: strconv.FormatInt(payload.datapack.StartTime.Add(-time.Duration(payload.datapack.PreDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(payload.datapack.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(payload.datapack.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(payload.datapack.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.datapack.Name)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.datapack.Name)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	for _, envVar := range payload.benchmark.EnvVars {
		if _, exists := envNameIndexMap[envVar.Key]; !exists {
			if envVar.TemplateString != "" {
				continue
			}

			if envVar.TemplateString != "" {
				logrus.Warnf("Skipping templated env var %s in benchmark version %d", envVar.Key, payload.benchmark.ID)
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
