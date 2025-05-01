package executor

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/tracing"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type CollectionPayload struct {
	Algorithm   string
	Dataset     string
	ExecutionID int
}

func executeCollectResult(ctx context.Context, task *UnifiedTask) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)

		collectPayload, err := parseCollectionPayload(task.Payload)
		if err != nil {
			return err
		}

		path := config.GetString("nfs.path")

		if collectPayload.Algorithm == "detector" {
			conclusionCSV := filepath.Join(path, collectPayload.Dataset, consts.DetectorConclusionFile)
			content, err := os.ReadFile(conclusionCSV)
			if err != nil {
				span.AddEvent(fmt.Sprintf("there is no conclusion.csv file in %s, please check whether it is nomal", conclusionCSV))
				span.RecordError(err)
				return fmt.Errorf("there is no conclusion.csv file in %s, please check whether it is nomal", conclusionCSV)
			}

			results, err := readDetectorCSV(content, collectPayload.ExecutionID)
			if err != nil {
				span.AddEvent("failed to convert conclusion.csv to database struct")
				span.RecordError(err)
				return fmt.Errorf("failed to convert conclusion.csv to database struct: %v", err)
			}

			if len(results) == 0 {
				span.AddEvent("the detector result is empty")
				logrus.Info("the detector result is empty")
				updateTaskStatus(
					ctx,
					task.TraceID,
					fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
					map[string]any{
						consts.RdbEventTaskID:   task.TaskID,
						consts.RdbEventTaskType: consts.TaskTypeCollectResult,
						consts.RdbEventStatus:   consts.TaskStatusCompleted,
						consts.RdbEventPayload: map[string]any{
							consts.RdbPayloadDetectorResult: results,
						},
					})

				return nil
			}

			if err = database.DB.Create(&results).Error; err != nil {
				span.AddEvent("failed to save conclusion.csv to database")
				span.RecordError(err)
				return fmt.Errorf("failed to save conclusion.csv to database: %v", err)
			}

			updateTaskStatus(
				ctx,
				task.TraceID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				map[string]any{
					consts.RdbEventTaskID:   task.TaskID,
					consts.RdbEventTaskType: consts.TaskTypeCollectResult,
					consts.RdbEventStatus:   consts.TaskStatusCompleted,
					consts.RdbEventPayload: map[string]any{
						consts.RdbPayloadDetectorResult: "",
					},
				})
		} else {
			resultCSV := filepath.Join(path, collectPayload.Dataset, "result.csv")
			content, err := os.ReadFile(resultCSV)
			if err != nil {
				span.AddEvent(fmt.Sprintf("there is no result.csv file in %s, please check whether it is nomal", resultCSV))
				span.RecordError(err)
				return fmt.Errorf("There is no result.csv file, please check whether it is nomal")
			}

			results, err := readCSVContent2Result(content, collectPayload.ExecutionID)
			if err != nil {
				span.AddEvent("failed to convert result.csv to database struct")
				span.RecordError(err)
				return fmt.Errorf("convert result.csv to database struct failed: %v", err)
			}

			if err = database.DB.Create(&results).Error; err != nil {
				span.AddEvent("failed to save result.csv to database")
				span.RecordError(err)
				return fmt.Errorf("save result.csv to database failed: %v", err)
			}

			updateTaskStatus(
				ctx,
				task.TraceID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				map[string]any{
					consts.RdbEventTaskID:   task.TaskID,
					consts.RdbEventTaskType: consts.TaskTypeCollectResult,
					consts.RdbEventStatus:   consts.TaskStatusCompleted,
				})
		}

		return nil
	})
}

func parseCollectionPayload(payload map[string]any) (*CollectionPayload, error) {
	algorithm, ok := payload[consts.CollectAlgorithm].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf("Missing or invalid '%s' key in payload", consts.CollectAlgorithm)
	}

	dataset, ok := payload[consts.CollectDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf("Missing or invalid '%s' key in payload", consts.CollectDataset)
	}

	executionIDFloat, ok := payload[consts.CollectExecutionID].(float64)
	if !ok || executionIDFloat == 0.0 {
		return nil, fmt.Errorf("Missing '%s' key in payload", consts.CollectExecutionID)
	}
	executionID := int(executionIDFloat)

	return &CollectionPayload{
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
	}, nil
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
		})
	}

	return results, nil
}

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
		})
	}

	return results, nil
}
