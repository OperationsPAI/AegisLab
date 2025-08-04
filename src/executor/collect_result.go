package executor

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type collectionPayload struct {
	algorithm   dto.AlgorithmItem
	dataset     string
	executionID int
	timestamp   string
}

func executeCollectResult(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		collectPayload, err := parseCollectPayload(task.Payload)
		if err != nil {
			return err
		}

		if collectPayload.algorithm.Name == config.GetString("algo.detector") {
			results, err := repository.ListDetectorResultsByExecutionID(collectPayload.executionID)
			if err != nil {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatasetNoDetectorData,
				})

				span.AddEvent("failed to get detector results by execution ID")
				span.RecordError(err)
				return fmt.Errorf("failed to get detector results by execution ID: %v", err)
			}

			if len(results) == 0 {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatasetNoDetectorData,
				})
				span.AddEvent("no detector results found for the execution ID")
				logrus.Info("no detector results found for the execution ID")
				return nil
			}

			hasIssues := false
			for _, v := range results {
				if v.Issues != "{}" {
					hasIssues = true
				}
			}

			if !hasIssues {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatasetNoAnomaly,
					Payload:   results,
				})

				span.AddEvent("the detector result has no issues")
				logrus.Info("the detector result has no issues")
			} else {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatasetResultCollection,
					Payload:   results,
				})
			}

			updateTaskStatus(
				childCtx,
				task.TraceID,
				task.TaskID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				consts.TaskStatusCompleted,
				task.Type,
			)

			if hasIssues && repository.CheckCachedField(childCtx, consts.InjectionAlgorithmsKey, task.GroupID) {
				items, err := repository.GetCachedAlgorithmItemsFromRedis(childCtx, consts.InjectionAlgorithmsKey, task.GroupID)
				if err != nil {
					span.AddEvent("failed to get algorithms from redis")
					span.RecordError(err)
					return fmt.Errorf("failed to get algorithms from redis: %v", err)
				}

				for _, item := range items {
					childTask := &dto.UnifiedTask{
						Type: consts.TaskTypeRunAlgorithm,
						Payload: map[string]any{
							consts.ExecuteAlgorithm: item,
							consts.ExecuteDataset:   collectPayload.dataset,
						},
						Immediate:    true,
						TraceID:      task.TraceID,
						GroupID:      task.GroupID,
						ProjectID:    task.ProjectID,
						TraceCarrier: task.TraceCarrier,
					}

					if _, _, err := SubmitTask(childCtx, childTask); err != nil {
						span.AddEvent("failed to submit algorithm execution task")
						span.RecordError(err)
						return fmt.Errorf("failed to submit algorithm execution task: %v", err)
					}
				}

				logrus.Infof("Algorithm executions submitted successfully")
			}
		} else {
			path := config.GetString("jfs.path")
			algorithmPath := filepath.Join(path, collectPayload.dataset, collectPayload.algorithm.Name, collectPayload.timestamp)
			resultCSV := filepath.Join(algorithmPath, consts.ExecutionResultFile)
			content, err := os.ReadFile(resultCSV)
			if err != nil {
				span.AddEvent("failed to read result.csv file")
				span.RecordError(err)
				return fmt.Errorf("failed to read result.csv file: %v", err)
			}

			results, err := readCSVContent2Result(content, collectPayload.executionID)
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

			repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeCollectResult,
				EventName: consts.EventAlgoResultCollection,
				Payload:   results,
			})

			updateTaskStatus(
				childCtx,
				task.TraceID,
				task.TaskID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				consts.TaskStatusCompleted,
				task.Type,
			)
		}

		logrus.WithField("task_id", task.TaskID).Info("Collect result task completed successfully")
		return nil
	})
}

func parseCollectPayload(payload map[string]any) (*collectionPayload, error) {
	algorithm, err := utils.ConvertToType[dto.AlgorithmItem](payload[consts.CollectAlgorithm])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to AlgorithmItem: %v", consts.CollectAlgorithm, err)
	}

	dataset, ok := payload[consts.CollectDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.CollectDataset)
	}

	executionIDFloat, ok := payload[consts.CollectExecutionID].(float64)
	if !ok || executionIDFloat == 0.0 {
		return nil, fmt.Errorf("missing '%s' key in payload", consts.CollectExecutionID)
	}
	executionID := int(executionIDFloat)

	timestamp, ok := payload[consts.CollectTimestamp].(string)
	if !ok || timestamp == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.CollectTimestamp)
	}

	return &collectionPayload{
		algorithm:   algorithm,
		dataset:     dataset,
		executionID: executionID,
		timestamp:   timestamp,
	}, nil
}

func readDetectorCSV(csvContent []byte, executionID int) ([]database.Detector, error) {
	reader := csv.NewReader(bytes.NewReader(csvContent))

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"SpanName", "Issues", "AbnormalAvgDuration", "NormalAvgDuration", "AbnormalSuccRate", "NormalSuccRate", "AbnormalP90", "NormalP90", "AbnormalP95", "NormalP95", "AbnormalP99", "NormalP99"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	// Read all rows
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

		// Handle empty values
		var abnormalAvgDuration, normalAvgDuration, abnormalSuccRate, normalSuccRate *float64
		var abnormalP90, normalP90, abnormalP95, normalP95, abnormalP99, normalP99 *float64

		// If the field is not empty, convert to float64, otherwise set to nil
		if row[2] != "" {
			val, err := strconv.ParseFloat(row[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AbnormalAvgDuration value in row %d: %v", i+1, err)
			}
			abnormalAvgDuration = &val
		}
		if row[3] != "" {
			val, err := strconv.ParseFloat(row[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NormalAvgDuration value in row %d: %v", i+1, err)
			}
			normalAvgDuration = &val
		}
		if row[4] != "" {
			val, err := strconv.ParseFloat(row[4], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AbnormalSuccRate value in row %d: %v", i+1, err)
			}
			abnormalSuccRate = &val
		}
		if row[5] != "" {
			val, err := strconv.ParseFloat(row[5], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NormalSuccRate value in row %d: %v", i+1, err)
			}
			normalSuccRate = &val
		}
		if row[6] != "" {
			val, err := strconv.ParseFloat(row[6], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AbnormalP90 value in row %d: %v", i+1, err)
			}
			abnormalP90 = &val
		}
		if row[7] != "" {
			val, err := strconv.ParseFloat(row[7], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NormalP90 value in row %d: %v", i+1, err)
			}
			normalP90 = &val
		}
		if row[8] != "" {
			val, err := strconv.ParseFloat(row[8], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AbnormalP95 value in row %d: %v", i+1, err)
			}
			abnormalP95 = &val
		}
		if row[9] != "" {
			val, err := strconv.ParseFloat(row[9], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NormalP95 value in row %d: %v", i+1, err)
			}
			normalP95 = &val
		}
		if row[10] != "" {
			val, err := strconv.ParseFloat(row[10], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AbnormalP99 value in row %d: %v", i+1, err)
			}
			abnormalP99 = &val
		}
		if row[11] != "" {
			val, err := strconv.ParseFloat(row[11], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid NormalP99 value in row %d: %v", i+1, err)
			}
			normalP99 = &val
		}

		// Add data to results
		results = append(results, database.Detector{
			ExecutionID:         executionID,
			SpanName:            spanName,
			Issues:              issues,
			AbnormalAvgDuration: abnormalAvgDuration,
			NormalAvgDuration:   normalAvgDuration,
			AbnormalSuccRate:    abnormalSuccRate,
			NormalSuccRate:      normalSuccRate,
			AbnormalP90:         abnormalP90,
			NormalP90:           normalP90,
			AbnormalP95:         abnormalP95,
			NormalP95:           normalP95,
			AbnormalP99:         abnormalP99,
			NormalP99:           normalP99,
		})
	}

	return results, nil
}

func readCSVContent2Result(csvContent []byte, executionID int) ([]database.GranularityResult, error) {
	reader := csv.NewReader(bytes.NewReader(csvContent))

	// Read header
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
