package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
)

func parseTime(timeStr string) time.Time {
	layout := "2006-01-02 15:04:05-07:00"
	t, _ := time.Parse(layout, timeStr)
	return t
}

func TestCreateAlgoJob(t *testing.T) {
	algo := "e-diagnose"
	datasetname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	jobname := fmt.Sprintf("%s-%s", algo, datasetname)
	image := "10.10.10.240/library/e-diagnose:latest"
	startTime := parseTime("2025-01-14 17:40:20+08:00")
	endTime := parseTime("2025-01-14 17:45:19+08:00")
	err := createAlgoJob(context.Background(), datasetname, jobname, "experiment", image, []string{"python", "run_exp.py"}, startTime, endTime)
	time.Sleep(time.Second * 5)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteAlgoJob(t *testing.T) {
	algo := "e-diagnose"
	datasetname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	jobname := fmt.Sprintf("%s-%s", algo, datasetname)
	err := client.DeleteK8sJob(context.Background(), "experiment", jobname)
	time.Sleep(time.Second * 5)
	if err != nil {
		t.Error(err)
	}
}
