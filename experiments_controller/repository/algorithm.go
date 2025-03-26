package repository

import (
	"errors"
	"fmt"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"gorm.io/gorm"
)

func GetDetectorRecordByDatasetID(datasetID int) (dto.DetectorRecord, error) {
	var record dto.DetectorRecord

	selectFields := `
        detectors.span_name AS span_name,
        detectors.issues AS issues,
        detectors.avg_duration AS avg_duration,
        detectors.succ_rate AS succ_rate,
        detectors.p90 AS p90,
        detectors.p95 AS p95,
        detectors.p99 AS p99
    `

	query := database.DB.
		Table("detectors").
		Select(selectFields).
		Joins(`
            LEFT JOIN execution_results 
            ON detectors.execution_id = execution_results.id
        `).
		Where("execution_results.dataset = ?", datasetID).
		Order("detectors.created_at DESC").Limit(1)

	if err := query.Find(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DetectorRecord{}, nil
		}

		return dto.DetectorRecord{}, fmt.Errorf("database query error: %w", err)
	}

	return record, nil
}

func GetExecutionRecordsByDatasetID(datasetID int, sortOrder string) ([]dto.ExecutionRecord, error) {
	var executions []database.ExecutionResult
	query := database.DB.
		Where("dataset = ? and algorithm != 'detector'", datasetID).
		Order(sortOrder)

	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to get executions: %v", err)
	}

	var execIDs []int
	for _, e := range executions {
		execIDs = append(execIDs, e.ID)
	}

	var granularities []database.GranularityResult
	if len(execIDs) > 0 {
		if err := database.DB.
			Where("execution_id IN (?)", execIDs).
			Find(&granularities).Error; err != nil {
			return nil, fmt.Errorf("failed to get granularities: %v", err)
		}
	}

	resultMap := make(map[int]dto.ExecutionRecord)
	for _, exec := range executions {
		resultMap[exec.ID] = dto.ExecutionRecord{
			Algorithm:          exec.Algorithm,
			GranularityResults: []dto.GranularityRecord{},
		}
	}

	for _, gran := range granularities {
		if record, exists := resultMap[gran.ExecutionID]; exists {
			record.GranularityResults = append(record.GranularityResults, dto.GranularityRecord{
				Level:      gran.Level,
				Result:     gran.Result,
				Rank:       gran.Rank,
				Confidence: gran.Confidence,
			})
			resultMap[gran.ExecutionID] = record
		}
	}

	results := make([]dto.ExecutionRecord, 0)
	for _, exec := range executions {
		results = append(results, resultMap[exec.ID])
	}

	return results, nil
}
