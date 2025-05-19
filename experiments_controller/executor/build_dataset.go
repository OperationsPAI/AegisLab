package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
)

type datasetPayload struct {
	Benchmark   string
	Name        string
	PreDuration int
	EnvVars     map[string]string
	StartTime   time.Time
	EndTime     time.Time
}

func executeBuildDataset(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		payload, err := parseDatasetPayload(task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse dataset payload")
			return err
		}

		annotations, err := getAnnotations(ctx, task)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get annotations")
			return err
		}

		imageName := fmt.Sprintf("%s_dataset", payload.Benchmark)
		tag, err := client.GetHarborClient().GetLatestTag(imageName)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get latest tag")
			return fmt.Errorf("failed to get lataest tag of %s: %v", imageName, err)
		}

		jobName := task.TaskID
		image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), imageName, tag)
		labels := map[string]string{
			consts.LabelTaskID:   task.TaskID,
			consts.LabelTraceID:  task.TraceID,
			consts.LabelGroupID:  task.GroupID,
			consts.LabelTaskType: string(consts.TaskTypeBuildDataset),
			consts.LabelDataset:  payload.Name,
		}

		return createDatasetJob(ctx, jobName, image, annotations, labels, payload)
	})
}

func parseDatasetPayload(payload map[string]any) (*datasetPayload, error) {
	return tracing.WithSpanReturnValue(context.Background(), func(ctx context.Context) (*datasetPayload, error) {
		message := "missing or invalid '%s' key in payload"

		benchmark, ok := payload[consts.BuildBenchmark].(string)
		if !ok || benchmark == "" {
			return nil, fmt.Errorf(message, consts.BuildBenchmark)
		}

		name, ok := payload[consts.BuildDataset].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf(message, consts.BuildDataset)
		}

		preDurationFloat, ok := payload[consts.BuildPreDuration].(float64)
		if !ok || preDurationFloat <= 0 {
			return nil, fmt.Errorf(message, consts.BuildPreDuration)
		}
		preDuration := int(preDurationFloat)

		_, startTimeExists := payload[consts.BuildStartTime]
		_, endTimeExists := payload[consts.BuildEndTime]

		var startTime, endTime time.Time
		if startTimeExists && endTimeExists {
			startTimePtr, err := parseTimePtrFromPayload(payload, consts.BuildStartTime)
			if err != nil {
				return nil, fmt.Errorf(message, consts.BuildStartTime)
			}

			endTimePtr, err := parseTimePtrFromPayload(payload, consts.BuildEndTime)
			if err != nil {
				return nil, fmt.Errorf(message, consts.BuildEndTime)
			}

			startTime = *startTimePtr
			endTime = *endTimePtr
		} else {
			datasetItem, err := repository.GetDatasetByName(name, consts.DatasetInjectSuccess, consts.DatasetBuildFailed)
			if err != nil {
				return nil, fmt.Errorf("query database for dataset failed: %v", err)
			}

			startTime = datasetItem.StartTime
			endTime = datasetItem.EndTime
		}

		result := datasetPayload{
			Benchmark:   benchmark,
			Name:        name,
			PreDuration: preDuration,
			StartTime:   startTime,
			EndTime:     endTime,
		}
		if e, exists := payload[consts.BuildEnvVars].(map[string]any); exists {
			envVars := make(map[string]string, len(e))
			for key, value := range e {
				strValue, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf(message, consts.ExecuteEnvVars)
				}
				envVars[key] = strValue
			}

			result.EnvVars = envVars
		}

		return &result, nil
	})
}

func parseTimePtrFromPayload(payload map[string]any, key string) (*time.Time, error) {
	return tracing.WithSpanReturnValue(context.Background(), func(ctx context.Context) (*time.Time, error) {
		timeStr, ok := payload[key].(string)
		if !ok {
			return nil, fmt.Errorf("%s must be a string", key)
		}

		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %s format: %v", key, err)
		}

		return &t, nil
	})
}

func createDatasetJob(ctx context.Context, jobName, image string, annotations map[string]string, labels map[string]string, payload *datasetPayload) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		restartPolicy := corev1.RestartPolicyNever
		backoffLimit := int32(2)
		parallelism := int32(1)
		completions := int32(1)
		jobNamespace := config.GetString("k8s.namespace")
		command := []string{"bash", "/entrypoint.sh"}

		envVars := getDatasetJobEnvVars(payload)
		return k8s.CreateJob(ctx, k8s.JobConfig{
			Namespace:     jobNamespace,
			JobName:       jobName,
			Image:         image,
			Command:       command,
			RestartPolicy: restartPolicy,
			BackoffLimit:  backoffLimit,
			Parallelism:   parallelism,
			Completions:   completions,
			EnvVars:       envVars,
			Annotations:   annotations,
			Labels:        labels,
		})
	})
}

func getDatasetJobEnvVars(payload *datasetPayload) []corev1.EnvVar {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "TIMEZONE", Value: tz},
		{Name: "NORMAL_START", Value: strconv.FormatInt(payload.StartTime.Add(-time.Duration(payload.PreDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(payload.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(payload.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(payload.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.Name)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.Name)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	if payload.EnvVars != nil {
		for name, value := range payload.EnvVars {
			if index, ok := envNameIndexMap[name]; ok {
				jobEnvVars[index].Value = value
			} else {
				jobEnvVars = append(jobEnvVars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
			}
		}
	}

	return jobEnvVars
}
