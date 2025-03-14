package executor

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/google/uuid"
)

func TestQueryCRDByName(t *testing.T) {
	namespace := "ts"
	datasetName := "ts-ts-preserve-service-request-abort-fhkpjm"
	// datasetName := "ts-ts-preserve-service-cpu-exhaustion-rtlt6h"
	if _, _, err := chaosCli.QueryCRDByName(namespace, datasetName); err != nil {
		t.Error(err)
	}
}

func TestCreateDatasetJob(t *testing.T) {
	datasetName := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	startTime := parseTime("2025-01-14 17:40:20+08:00")
	endTime := parseTime("2025-01-14 17:45:19+08:00")

	jobName := fmt.Sprintf("dataset-%s", datasetName)
	image := fmt.Sprintf("%s/%s_dataset:latest", config.GetString("harbor.repository"), "clickhouse")
	labels := map[string]string{
		LabelTaskID:    uuid.New().String(),
		LabelTaskType:  string(TaskTypeRunAlgorithm),
		LabelDataset:   datasetName,
		LabelStartTime: strconv.FormatInt(startTime.Unix(), 10),
		LabelEndTime:   strconv.FormatInt(endTime.Unix(), 10),
	}
	jobEnv := &k8s.JobEnv{
		Namespace: "ts",
		StartTime: startTime,
		EndTime:   endTime,
	}

	err := createDatasetJob(context.Background(), datasetName, jobName, config.GetString("k8s.namespace"), image, []string{"python", "prepare_inputs.py"}, labels, jobEnv)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteDatasetJob(t *testing.T) {
	jobname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	k8s.DeleteJob(context.Background(), "experiment", jobname)
}
