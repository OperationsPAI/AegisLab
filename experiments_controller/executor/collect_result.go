package executor

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
)

type ResultPayload struct {
	DatasetName string
	ExecutionID int
}

func parseResultPayload(payload map[string]interface{}) (*ResultPayload, error) {
	datasetName, ok := payload[CollectDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", CollectDataset)
	}
	executionID, err := strconv.Atoi(payload[CollectExecutionID].(string))
	if err != nil || executionID == 0 {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", CollectExecutionID)
	}
	return &ResultPayload{
		DatasetName: datasetName,
		ExecutionID: executionID,
	}, nil
}

// 读取 CSV 内容并转换为结果
func readCSVContent2Result(csvContent []byte, executionID int) ([]database.GranularityResult, error) {
	reader := csv.NewReader(bytes.NewReader(csvContent))

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

func readDetectorCSV(csvContent []byte, executionID int) ([]database.Detector, error) {
	reader := csv.NewReader(bytes.NewReader(csvContent))

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

func executeCollectResult(ctx context.Context, taskID string, payload map[string]interface{}) error {
	return collectResult(taskID, payload)
}

func collectResult(taskID string, payload map[string]interface{}) error {
	resultPayload, err := parseResultPayload(payload)
	if err != nil {
		return err
	}

	path := config.GetString("nfs.path")
	if path == "" {
		path = "/mnt/nfs/rcabench_dataset"
	}

	resultCSV := filepath.Join(path, resultPayload.DatasetName, "result.csv")
	content, err := os.ReadFile(resultCSV)
	if err != nil {
		updateTaskStatus(taskID, "Error", "There is no result.csv file, please check whether it is nomal")
	} else {
		results, err := readCSVContent2Result(content, resultPayload.ExecutionID)
		if err != nil {
			return fmt.Errorf("convert result.csv to database struct failed: %v", err)
		}

		if err = database.DB.Create(&results).Error; err != nil {
			return fmt.Errorf("save result.csv to database failed: %v", err)
		}
	}

	conclusionCSV := filepath.Join(path, resultPayload.DatasetName, "conclusion.csv")
	content, err = os.ReadFile(conclusionCSV)
	if err != nil {
		updateTaskStatus(taskID, "Error", "There is no conclusion.csv file in /app/output, please check whether it is nomal")
	} else {
		results, err := readDetectorCSV(content, resultPayload.ExecutionID)
		if err != nil {
			return fmt.Errorf("convert conclusion.csv to database struct failed: %v", err)
		}

		if err = database.DB.Create(&results).Error; err != nil {
			return fmt.Errorf("save conclusion.csv to database failed: %v", err)
		}
	}

	return nil
}
