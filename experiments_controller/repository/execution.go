package repository

import (
	"fmt"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
)

func CreateExecutionResult(algorithm, taskID string, datasetID int) (int, error) {
	executionResult := database.ExecutionResult{
		TaskID:    taskID,
		Dataset:   datasetID,
		Algorithm: algorithm,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return 0, err
	}

	return executionResult.ID, nil
}

func ListExecutionRecordByDatasetID(datasetID int, sortOrder string) ([]dto.ExecutionRecord, error) {
	query := database.DB.
		Model(&database.ExecutionResult{}).
		Where("dataset = ? and algorithm != 'detector'", datasetID).
		Order(sortOrder)

	var executions []database.ExecutionResult
	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to get executions: %v", err)
	}

	var execIDs []int
	for _, e := range executions {
		execIDs = append(execIDs, e.ID)
	}

	if len(execIDs) == 0 {
		return nil, fmt.Errorf("failed to get executions")
	}

	granularities, err := listGranularityWithFilters(execIDs, []string{}, 5)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[int]dto.ExecutionRecord)
	for _, exec := range executions {
		resultMap[exec.ID] = dto.ExecutionRecord{
			Algorithm:          exec.Algorithm,
			GranularityRecords: []dto.GranularityRecord{},
		}
	}

	for _, gran := range granularities {
		if result, exists := resultMap[gran.ExecutionID]; exists {
			var record dto.GranularityRecord
			record.Convert(gran)
			result.GranularityRecords = append(result.GranularityRecords, record)
		}
	}

	results := make([]dto.ExecutionRecord, 0)
	for _, exec := range executions {
		results = append(results, resultMap[exec.ID])
	}

	return results, nil
}

func ListExecutionRecordByExecID(executionIDs []int,
	algorithms,
	levels []string,
	rank int,
) ([]dto.ExecutionRecordWithDatasetID, error) {
	query := database.DB.
		Model(&database.ExecutionResult{}).
		Select("id, algorithm, dataset")

	if len(executionIDs) > 0 {
		query = query.Where("id IN (?)", executionIDs)
	}

	if len(algorithms) > 0 {
		query = query.Where("algorithm IN (?)", algorithms)
	}

	var executions []database.ExecutionResult
	if err := query.Find(&executions).Error; err != nil {
		return nil, err
	}

	var execIDs []int
	for _, e := range executions {
		execIDs = append(execIDs, e.ID)
	}

	if len(execIDs) == 0 {
		return nil, fmt.Errorf("failed to get executions")
	}

	granularities, err := listGranularityWithFilters(execIDs, levels, rank)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[int]dto.ExecutionRecordWithDatasetID)
	for _, exec := range executions {
		resultMap[exec.ID] = dto.ExecutionRecordWithDatasetID{
			DatasetID: exec.Dataset,
			ExecutionRecord: dto.ExecutionRecord{
				Algorithm:          exec.Algorithm,
				GranularityRecords: []dto.GranularityRecord{},
			},
		}
	}

	for _, gran := range granularities {
		if result, exists := resultMap[gran.ExecutionID]; exists {
			var record dto.GranularityRecord
			record.Convert(gran)
			result.GranularityRecords = append(result.GranularityRecords, record)
		}
	}

	results := make([]dto.ExecutionRecordWithDatasetID, 0)
	for _, exec := range executions {
		results = append(results, resultMap[exec.ID])
	}

	return results, nil
}
