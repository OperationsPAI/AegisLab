package executor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/tracing"
	"aegis/utils"
)

type datasetPayload struct {
	containerVersion database.ContainerVersion
	injectionName    string
	preDuration      int
	envVars          map[string]string
	startTime        time.Time
	endTime          time.Time
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

		jobName := task.TaskID
		labels := map[string]string{
			consts.LabelTaskID:    task.TaskID,
			consts.LabelTraceID:   task.TraceID,
			consts.LabelGroupID:   task.GroupID,
			consts.LabelProjectID: getDefaultIDString(task.ProjectID),
			consts.LabelUserID:    getDefaultIDString(task.UserID),
			consts.LabelTaskType:  string(consts.TaskTypeBuildDataset),
			consts.LabelDataset:   payload.injectionName,
		}

		return createDatasetJob(ctx, jobName, payload.containerVersion.ImageRef, annotations, labels, payload)
	})
}

func parseDatasetPayload(payload map[string]any) (*datasetPayload, error) {
	return tracing.WithSpanReturnValue(context.Background(), func(ctx context.Context) (*datasetPayload, error) {
		message := "missing or invalid '%s' key in payload"

		containerVersion, err := utils.ConvertToType[database.ContainerVersion](payload[consts.BuildContainerVersion])
		if err != nil {
			return nil, fmt.Errorf("failed to convert '%s' to ContainerVersion: %v", consts.BuildContainerVersion, err)
		}

		injectionName, ok := payload[consts.BuildDataset].(string)
		if !ok || injectionName == "" {
			return nil, fmt.Errorf(message, consts.BuildDataset)
		}

		preDurationFloat, ok := payload[consts.BuildPreDuration].(float64)
		if !ok || preDurationFloat <= 0 {
			return nil, fmt.Errorf(message, consts.BuildPreDuration)
		}
		preDuration := int(preDurationFloat)

		envVars, err := utils.ConvertToType[map[string]string](payload[consts.ExecuteEnvVars])
		if err != nil {
			return nil, fmt.Errorf("failed to convert '%s' to map[string]string: %v", consts.ExecuteEnvVars, err)
		}

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
		}

		return &datasetPayload{
			containerVersion: containerVersion,
			injectionName:    injectionName,
			preDuration:      preDuration,
			envVars:          envVars,
			startTime:        startTime,
			endTime:          endTime,
		}, nil
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

func createDatasetJob(ctx context.Context, jobName, image string, annotations, labels map[string]string, payload *datasetPayload) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)

		jobEnvVars, err := getDatasetJobEnvVars(payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get job environment variables")
			return err
		}

		return k8s.CreateJob(ctx, &k8s.JobConfig{
			JobName:     jobName,
			Image:       image,
			Command:     strings.Split(payload.containerVersion.Command, " "),
			EnvVars:     jobEnvVars,
			Annotations: annotations,
			Labels:      labels,
		})
	})
}

func getDatasetJobEnvVars(payload *datasetPayload) ([]corev1.EnvVar, error) {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "ENV_MODE", Value: config.GetString("system.env_mode")},
		{Name: "TIMEZONE", Value: tz},
		{Name: "NORMAL_START", Value: strconv.FormatInt(payload.startTime.Add(-time.Duration(payload.preDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(payload.startTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(payload.startTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(payload.endTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.injectionName)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.injectionName)},
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
		if payload.containerVersion.EnvVars != "" {
			envVarsArray := strings.Split(payload.containerVersion.EnvVars, ",")
			for _, envVar := range envVarsArray {
				if _, exists := extraEnvVarMap[envVar]; !exists {
					return nil, fmt.Errorf("environment variable %s is required but not provided in dataset build payload", envVar)
				}
			}
		}
	} else {
		if payload.containerVersion.EnvVars != "" {
			return nil, fmt.Errorf("environment variables %s are required but not provided in dataset build payload", payload.containerVersion.EnvVars)
		}
	}

	return jobEnvVars, nil
}
