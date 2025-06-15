package repository

import (
	"fmt"
	"strings"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
)

func CreateExecutionResult(taskID, algorithm, dataset string) (int, error) {
	executionResult := database.ExecutionResult{
		TaskID:    taskID,
		Algorithm: algorithm,
		Dataset:   dataset,
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
		return []dto.ExecutionRecord{}, nil
	}

	granularities, err := listGranularityWithFilters(execIDs, []string{}, 5)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[int]*dto.ExecutionRecord)
	for _, exec := range executions {
		resultMap[exec.ID] = &dto.ExecutionRecord{
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
		results = append(results, *resultMap[exec.ID])
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

	resultMap := make(map[int]*dto.ExecutionRecordWithDatasetID)
	for _, exec := range executions {
		resultMap[exec.ID] = &dto.ExecutionRecordWithDatasetID{
			// TODO 修改
			DatasetID: 0,
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
		results = append(results, *resultMap[exec.ID])
	}

	return results, nil
}

func ListExecutionRawData(pairs []dto.AlgorithmDatasetPair) ([]dto.RawDataItem, error) {
	if len(pairs) == 0 {
		return nil, fmt.Errorf("no algorithm-dataset pairs provided")
	}

	execIDMap, err := getLatestExecutionMap(pairs)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution IDs: %v", err)
	}

	if len(execIDMap) == 0 {
		return nil, fmt.Errorf("no execution IDs found for the provided pairs")
	}

	var execIDs []int
	for id := range execIDMap {
		execIDs = append(execIDs, id)
	}

	var granularityResults []database.GranularityResult
	if err := database.DB.
		Model(&database.GranularityResult{}).
		Where("execution_id IN (?)", execIDs).
		Find(&granularityResults).Error; err != nil {
		return nil, fmt.Errorf("failed to query granularity results: %v", err)
	}

	var items []dto.RawDataItem
	for id, pairStr := range execIDMap {
		parts := strings.Split(pairStr, "_")
		algorithm := parts[0]
		dataset := parts[1]

		var records []dto.GranularityRecord
		for _, gran := range granularityResults {
			if id == gran.ExecutionID {
				var record dto.GranularityRecord
				record.Convert(gran)
				records = append(records, record)
			}
		}

		items = append(items, dto.RawDataItem{
			Algorithm: algorithm,
			Dataset:   dataset,
			Entries:   records,
		})
	}

	return items, nil
}

func getLatestExecutionMap(pairs []dto.AlgorithmDatasetPair) (map[int]string, error) {
	uniquePairs := make(map[string]dto.AlgorithmDatasetPair)
	for _, pair := range pairs {
		key := fmt.Sprintf("%s_%s", pair.Algorithm, pair.Dataset)
		uniquePairs[key] = pair
	}

	var algorithms []string
	var datasets []string
	for _, pair := range uniquePairs {
		algorithms = append(algorithms, pair.Algorithm)
		datasets = append(datasets, pair.Dataset)
	}

	var executions []database.ExecutionResult
	if err := database.DB.
		Model(&database.ExecutionResult{}).
		Where("algorithm IN (?) AND dataset IN (?)", algorithms, datasets).
		Order("algorithm, dataset, created_at DESC").
		Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query executions: %w", err)
	}

	execIDMap := make(map[int]string)
	seen := make(map[string]bool)

	for _, exec := range executions {
		key := fmt.Sprintf("%s_%s", exec.Algorithm, exec.Dataset)
		if !seen[key] {
			if _, exists := uniquePairs[key]; exists {
				execIDMap[exec.ID] = key
				seen[key] = true
			}
		}
	}

	return execIDMap, nil
}
