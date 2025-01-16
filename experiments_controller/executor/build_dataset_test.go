package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/CUHK-SE-Group/rcabench/client"
)

func TestCreateDatasetJob(t *testing.T) {
	datasetname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	startTime := parseTime("2025-01-14 17:40:20+08:00")
	endTime := parseTime("2025-01-14 17:45:19+08:00")
	err := createDatasetJob(context.Background(), datasetname, fmt.Sprintf("dataset-%s", datasetname), "experiment", fmt.Sprintf("10.10.10.240/library/clickhouse_dataset:latest"), []string{"python", "prepare_inputs.py"}, startTime, endTime)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteDatasetJob(t *testing.T) {
	jobname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	client.DeleteK8sJob(context.Background(), "experiment", jobname)
}
