package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/google/uuid"
)

func parseTime(timeStr string) time.Time {
	layout := "2006-01-02 15:04:05-07:00"
	t, _ := time.Parse(layout, timeStr)
	return t
}

func TestCreateAlgoJob(t *testing.T) {
	algo := "e-diagnose"
	datasetName := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	startTime := parseTime("2025-01-14 17:40:20+08:00")
	endTime := parseTime("2025-01-14 17:45:19+08:00")

	jobName := fmt.Sprintf("%s-%s", algo, datasetName)
	image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), algo, "latest")
	labels := map[string]string{
		LabelTaskID:      uuid.New().String(),
		LabelTaskType:    string(TaskTypeRunAlgorithm),
		LabelDataset:     datasetName,
		LabelExecutionID: fmt.Sprint(1),
	}
	jobEnv := &client.JobEnv{
		StartTime: startTime,
		EndTime:   endTime,
	}

	err := createAlgoJob(context.Background(), datasetName, jobName, "experiment", image, []string{"python", "run_exp.py"}, labels, jobEnv)
	time.Sleep(time.Second * 5)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteAlgoJob(t *testing.T) {
	algo := "e-diagnose"
	datasetname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"

	jobName := fmt.Sprintf("%s-%s", algo, datasetname)

	err := client.DeleteK8sJob(context.Background(), "experiment", jobName)
	time.Sleep(time.Second * 5)
	if err != nil {
		t.Error(err)
	}
}
