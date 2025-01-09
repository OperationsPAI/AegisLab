package executor

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	corev1 "k8s.io/api/core/v1"

	"gorm.io/gorm"
)

type AlgorithmExecutionPayload struct {
	Benchmark   string
	Algorithm   string
	DatasetName string
}

func executeAlgorithm(ctx context.Context, taskID string, payload map[string]interface{}) error {
	algPayload, err := parseAlgorithmExecutionPayload(payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", algPayload.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", algPayload.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", algPayload.DatasetName, err)
	}

	executionResult := database.ExecutionResult{
		Dataset: faultRecord.ID,
		TaskID:  taskID,
		Algo:    algPayload.Algorithm,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	updateTaskStatus(taskID, "Running", fmt.Sprintf("Running algorithm for task %s", taskID))
	createAlgoJob(ctx, algPayload.DatasetName, fmt.Sprintf("%s|%s", algPayload.Algorithm, algPayload.DatasetName), "experiment", algPayload.Algorithm, []string{"python", "run_exp.py"})
	return nil
}
func createAlgoJob(ctx context.Context, datasetname, jobname, namespace, image string, command []string) error {
	fc := client.NewK8sClient()
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(1)
	completions := int32(1)

	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	envVars := []corev1.EnvVar{
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetname)},
		{Name: "OUTPUT_PATH", Value: "/app/output"},
		{Name: "TIMEZONE", Value: tz},
		{Name: "WORKSPACE", Value: "/app"},
	}
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "nfs-volume",
			MountPath: "/data",
		},
	}
	pvc := config.GetString("nfs.pvc_name")
	if config.GetString("nfs.pvc_name") == "" {
		pvc = "nfs-shared-pvc"
	}
	volumes := []corev1.Volume{
		{
			Name: "nfs-volume",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc,
				},
			},
		},
	}

	err := client.CreateK8sJob(ctx, fc, namespace, jobname, image, command, restartPolicy,
		backoffLimit, parallelism, completions, envVars, volumeMounts, volumes)
	return err
}

// 解析算法执行任务的 Payload
func parseAlgorithmExecutionPayload(payload map[string]interface{}) (*AlgorithmExecutionPayload, error) {
	benchmark, ok := payload[EvalPayloadBench].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadBench)
	}
	algorithm, ok := payload[EvalPayloadAlgo].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadAlgo)
	}
	datasetName, ok := payload[EvalPayloadDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadDataset)
	}
	return &AlgorithmExecutionPayload{
		Benchmark:   benchmark,
		Algorithm:   algorithm,
		DatasetName: datasetName,
	}, nil
}

// 读取 CSV 内容并转换为结果
func readCSVContent2Result(csvContent string, executionID int) ([]database.GranularityResult, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"level", "result", "rank", "confidence"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %v", err)
	}

	var results []database.GranularityResult
	for i, row := range rows {
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has incorrect number of columns: got %d, expected %d", i+1, len(row), len(expectedHeader))
		}

		level := row[0]
		result := row[1]
		rank, err := strconv.Atoi(row[2])
		if err != nil {
			return nil, fmt.Errorf("invalid rank value in row %d: %v", i+1, err)
		}
		confidence, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid confidence value in row %d: %v", i+1, err)
		}

		results = append(results, database.GranularityResult{
			ExecutionID: executionID,
			Level:       level,
			Result:      result,
			Rank:        rank,
			Confidence:  confidence,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return results, nil
}

func readDetectorCSV(csvContent string, executionID int) ([]database.Detector, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"SpanName", "Issues", "AvgDuration", "SuccRate", "P90", "P95", "P99"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %v", err)
	}

	var results []database.Detector
	for i, row := range rows {
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has incorrect number of columns: got %d, expected %d", i+1, len(row), len(expectedHeader))
		}

		spanName := row[0]
		issues := row[1]

		// 处理空值
		var avgDuration, succRate, p90, p95, p99 *float64

		// 如果字段非空，转换为 float64，否则设置为 nil
		if row[2] != "" {
			val, err := strconv.ParseFloat(row[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AvgDuration value in row %d: %v", i+1, err)
			}
			avgDuration = &val
		}
		if row[3] != "" {
			val, err := strconv.ParseFloat(row[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid SuccRate value in row %d: %v", i+1, err)
			}
			succRate = &val
		}
		if row[4] != "" {
			val, err := strconv.ParseFloat(row[4], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P90 value in row %d: %v", i+1, err)
			}
			p90 = &val
		}
		if row[5] != "" {
			val, err := strconv.ParseFloat(row[5], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P95 value in row %d: %v", i+1, err)
			}
			p95 = &val
		}
		if row[6] != "" {
			val, err := strconv.ParseFloat(row[6], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P99 value in row %d: %v", i+1, err)
			}
			p99 = &val
		}

		// 将数据添加到结果
		results = append(results, database.Detector{
			ExecutionID: executionID,
			SpanName:    spanName,
			Issues:      issues,
			AvgDuration: avgDuration,
			SuccRate:    succRate,
			P90:         p90,
			P95:         p95,
			P99:         p99,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return results, nil
}
