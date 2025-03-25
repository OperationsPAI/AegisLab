package repository

import (
	"fmt"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
)

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

	var results []dto.ExecutionRecord
	for _, exec := range executions {
		results = append(results, resultMap[exec.ID])
	}

	return results, nil
}
