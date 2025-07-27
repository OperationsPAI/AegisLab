package repository

import (
	"encoding/json"
	"errors"
	"fmt"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"gorm.io/gorm"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
)

func CreateExecutionResult(taskID string, algorithmID, datasetID int) (int, error) {
	executionResult := database.ExecutionResult{
		TaskID:      taskID,
		AlgorithmID: algorithmID,
		DatasetID:   datasetID,
		Status:      consts.ExecutionInitial,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return 0, err
	}

	return executionResult.ID, nil
}

func ListExecutionRawDataByIds(params dto.RawDataReq) ([]dto.RawDataItem, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	query := database.DB.
		Where("id IN (?) and status = (?)", params.ExecutionIDs, consts.ExecutionSuccess)
	query = opts.AddTimeFilter(query, "created_at")

	var execResults []database.ExecutionResultProject
	if err := query.Find(&execResults).Error; err != nil {
		return nil, fmt.Errorf("failed to query execution results: %v", err)
	}

	execResultMap := make(map[int]database.ExecutionResultProject, len(execResults))
	for _, execResult := range execResults {
		execResultMap[execResult.ID] = execResult
	}

	for _, id := range params.ExecutionIDs {
		if _, exists := execResultMap[id]; !exists {
			return nil, fmt.Errorf("execution ID %d not found in the database", id)
		}
	}

	datasets := make([]string, 0, len(execResults))
	for _, execResult := range execResults {
		datasets = append(datasets, execResult.Dataset)
	}

	var granularityResults []database.GranularityResult
	if err := database.DB.
		Model(&database.GranularityResult{}).
		Where("execution_id IN (?)", params.ExecutionIDs).
		Find(&granularityResults).Error; err != nil {
		return nil, fmt.Errorf("failed to query granularity results: %v", err)
	}

	granMap := make(map[int][]dto.GranularityRecord, len(execResultMap))
	for _, gran := range granularityResults {
		var record dto.GranularityRecord
		record.Convert(gran)

		if _, exists := granMap[gran.ExecutionID]; !exists {
			granMap[gran.ExecutionID] = []dto.GranularityRecord{record}
		} else {
			granMap[gran.ExecutionID] = append(granMap[gran.ExecutionID], record)
		}
	}

	groundtruthMap, err := GetGroundtruthMap(datasets)
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth map: %v", err)
	}

	var items []dto.RawDataItem
	for id, execResult := range execResultMap {
		if granRecords, exists := granMap[execResult.ID]; exists {
			items = append(items, dto.RawDataItem{
				Algorithm:   execResult.Algorithm,
				Dataset:     execResult.Dataset,
				ExecutionID: id,
				Entries:     granRecords,
				Groundtruth: groundtruthMap[execResult.Dataset],
			})
		}
	}

	return items, nil
}

func ListExecutionRawDatasByPairs(params dto.RawDataReq) ([]dto.RawDataItem, error) {
	datasets := make([]string, 0, len(params.Pairs))
	for _, pair := range params.Pairs {
		datasets = append(datasets, pair.Dataset)
	}

	execIDMap, err := getLatestExecutionMapByPair(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution IDs: %v", err)
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

	granMap := make(map[int][]dto.GranularityRecord, len(execIDs))
	for _, gran := range granularityResults {
		var record dto.GranularityRecord
		record.Convert(gran)

		if _, exists := granMap[gran.ExecutionID]; !exists {
			granMap[gran.ExecutionID] = []dto.GranularityRecord{record}
		} else {
			granMap[gran.ExecutionID] = append(granMap[gran.ExecutionID], record)
		}
	}

	groundtruthMap, err := GetGroundtruthMap(datasets)
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth map: %v", err)
	}

	pairKeyIDMap := make(map[string]int, len(execIDMap))
	for id, storedPairKey := range execIDMap {
		pairKeyIDMap[storedPairKey] = id
	}

	var items []dto.RawDataItem
	for _, pair := range params.Pairs {
		item := &dto.RawDataItem{
			Algorithm:   pair.Algorithm,
			Dataset:     pair.Dataset,
			Groundtruth: groundtruthMap[pair.Dataset],
		}

		pairKey := fmt.Sprintf("%s_%s", pair.Algorithm, pair.Dataset)
		id, exists := pairKeyIDMap[pairKey]
		if exists {
			if granRecords, exists := granMap[id]; exists {
				item.ExecutionID = id
				item.Entries = granRecords
			}
		}

		items = append(items, *item)
	}

	return items, nil
}

func UpdateStatusByExecID(executionID int, status int) error {
	var record database.ExecutionResult
	err := database.DB.
		Where("id = ?", executionID).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("record with id %d not found", executionID)
		}

		return fmt.Errorf("failed to query record: %v", err)
	}

	result := database.DB.
		Model(&record).
		Updates(map[string]any{"status": status})

	if result.Error != nil {
		return fmt.Errorf("failed to update record: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("record found but no fields were updated, possibly because values are unchanged")
	}

	return nil
}

func getLatestExecutionMapByPair(params dto.RawDataReq) (map[int]string, error) {
	uniquePairs := make(map[string]dto.AlgorithmDatasetPair)
	for _, pair := range params.Pairs {
		key := fmt.Sprintf("%s_%s", pair.Algorithm, pair.Dataset)
		uniquePairs[key] = pair
	}

	var algorithms []string
	var datasets []string
	for _, pair := range uniquePairs {
		algorithms = append(algorithms, pair.Algorithm)
		datasets = append(datasets, pair.Dataset)
	}

	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	query := database.DB.Model(&database.ExecutionResultProject{}).
		Where("algorithm IN (?) AND dataset IN (?) AND status = (?)", algorithms, datasets, consts.ExecutionSuccess)
	query = opts.AddTimeFilter(query, "created_at")

	var executions []database.ExecutionResultProject
	if err := query.Order("algorithm, dataset, created_at DESC").
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

func GetGroundtruthMap(datasets []string) (map[string]chaos.Groundtruth, error) {
	engineConfMap, err := ListEngineConfigsByNames(datasets)
	if err != nil {
		return nil, err
	}

	groundtruthMap := make(map[string]chaos.Groundtruth, len(engineConfMap))
	for dataset, engineConf := range engineConfMap {
		var node chaos.Node
		if err := json.Unmarshal([]byte(engineConf), &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chaos-experiment node for dataset %s: %v", dataset, err)
		}

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert chaos-experiment node to InjectionConf for dataset %s: %v", dataset, err)
		}

		groundtruth, err := conf.GetGroundtruth()
		if err != nil {
			return nil, fmt.Errorf("failed to get ground truth for dataset %s: %v", dataset, err)
		}

		groundtruthMap[dataset] = groundtruth
	}

	return groundtruthMap, nil
}

// ListSuccessfulExecutions 获取所有成功执行的算法记录
func ListSuccessfulExecutions() ([]dto.SuccessfulExecutionItem, error) {
	return ListSuccessfulExecutionsWithFilter(dto.SuccessfulExecutionsReq{})
}

// ListSuccessfulExecutionsWithFilter 根据筛选条件获取成功执行的算法记录
func ListSuccessfulExecutionsWithFilter(req dto.SuccessfulExecutionsReq) ([]dto.SuccessfulExecutionItem, error) {
	var executions []database.ExecutionResultProject
	query := database.DB.Where("status = ?", consts.ExecutionSuccess)

	if req.StartTime != nil {
		query = query.Where("created_at >= ?", *req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", *req.EndTime)
	}

	query = query.Order("created_at DESC")

	if req.Offset != nil && *req.Offset > 0 {
		query = query.Offset(*req.Offset)
	}
	if req.Limit != nil && *req.Limit > 0 {
		query = query.Limit(*req.Limit)
	}

	err := query.Find(&executions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query successful executions: %v", err)
	}

	result := make([]dto.SuccessfulExecutionItem, len(executions))
	for i, exec := range executions {
		result[i] = dto.SuccessfulExecutionItem{
			ID:        exec.ID,
			Algorithm: exec.Algorithm,
			Dataset:   exec.Dataset,
			CreatedAt: exec.CreatedAt,
		}
	}

	return result, nil
}
