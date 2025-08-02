package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"gorm.io/gorm"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
)

func CreateExecutionResult(taskID string, algorithmID, datapackID int) (int, error) {
	executionResult := database.ExecutionResult{
		TaskID:      taskID,
		AlgorithmID: algorithmID,
		DatapackID:  datapackID,
		Status:      consts.ExecutionInitial,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return 0, err
	}

	// Add label to indicate the source of this execution result
	var labelValue, labelDescription string
	if taskID != "" {
		// TaskID is provided, this is system-managed
		labelValue = consts.ExecutionSourceSystem
		labelDescription = consts.ExecutionSystemDescription
	} else {
		// TaskID is empty, this is manual upload
		labelValue = consts.ExecutionSourceManual
		labelDescription = consts.ExecutionManualDescription
	}

	if err := AddExecutionResultLabel(executionResult.ID, consts.ExecutionLabelSource, labelValue, labelDescription); err != nil {
		fmt.Printf("Warning: Failed to create execution result label: %v\n", err)
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

func GetGroundtruthMap(datapacks []string) (map[string]chaos.Groundtruth, error) {
	engineConfMap, err := ListEngineConfigsByNames(datapacks)
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

// ListSuccessfulExecutions gets all successfully executed algorithm records
func ListSuccessfulExecutions() ([]dto.SuccessfulExecutionItem, error) {
	return ListSuccessfulExecutionsWithFilter(dto.SuccessfulExecutionsReq{})
}

// ListSuccessfulExecutionsWithFilter gets successfully executed algorithm records based on filter conditions
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

// GetExecutionStatistics returns statistics about executions
func GetExecutionStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total executions
	var total int64
	if err := database.DB.Model(&database.ExecutionResult{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total executions: %v", err)
	}
	stats["total"] = total

	// Executions by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	err := database.DB.Model(&database.ExecutionResult{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count executions by status: %v", err)
	}

	// Set status counts
	for _, sc := range statusCounts {
		switch sc.Status {
		case "pending":
			stats["pending"] = sc.Count
		case "running":
			stats["running"] = sc.Count
		case "completed":
			stats["completed"] = sc.Count
		case "failed":
			stats["failed"] = sc.Count
		case "cancelled":
			stats["cancelled"] = sc.Count
		default:
			stats[sc.Status] = sc.Count
		}
	}

	// Initialize missing statuses with 0
	statuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	for _, status := range statuses {
		if _, exists := stats[status]; !exists {
			stats[status] = 0
		}
	}

	return stats, nil
}

// GetExecutionCountByAlgorithm returns count of executions grouped by algorithm
func GetExecutionCountByAlgorithm() (map[string]int64, error) {
	type AlgorithmCount struct {
		Algorithm string `json:"algorithm"`
		Count     int64  `json:"count"`
	}

	var results []AlgorithmCount
	err := database.DB.Model(&database.ExecutionResult{}).
		Select("algorithm, COUNT(*) as count").
		Group("algorithm").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count executions by algorithm: %v", err)
	}

	algorithmCounts := make(map[string]int64)
	for _, result := range results {
		algorithmCounts[result.Algorithm] = result.Count
	}

	return algorithmCounts, nil
}

// GetRecentExecutionActivity returns execution activity for the last N days
func GetRecentExecutionActivity(days int) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Last N days activity
	startDate := time.Now().AddDate(0, 0, -days)
	var recentCount int64
	if err := database.DB.Model(&database.ExecutionResult{}).Where("created_at >= ?", startDate).Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent executions: %v", err)
	}
	stats[fmt.Sprintf("last_%d_days", days)] = recentCount

	// Today's executions
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	if err := database.DB.Model(&database.ExecutionResult{}).Where("created_at >= ?", today).Count(&todayCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's executions: %v", err)
	}
	stats["today"] = todayCount

	return stats, nil
}

// AddExecutionResultLabel adds a label to an execution result
func AddExecutionResultLabel(executionID int, labelKey, labelValue, description string) error {
	// Create or get the label
	label, err := CreateOrGetLabel(labelKey, labelValue, consts.LabelExecution, description)
	if err != nil {
		return fmt.Errorf("failed to create or get label: %v", err)
	}

	relation := database.ExecutionResultLabel{
		ExecutionID: executionID,
		LabelID:     label.ID,
	}

	return database.DB.Where("execution_id = ? AND label_id = ?", executionID, label.ID).
		FirstOrCreate(&relation).Error
}

// GetExecutionResultLabels retrieves all labels for an execution result
func GetExecutionResultLabels(executionID int) ([]database.Label, error) {
	var labels []database.Label
	err := database.DB.
		Joins("JOIN execution_result_labels ON execution_result_labels.label_id = labels.id").
		Where("execution_result_labels.execution_id = ?", executionID).
		Find(&labels).Error
	return labels, err
}

// RemoveExecutionResultLabel removes a specific label from an execution result
func RemoveExecutionResultLabel(executionID int, labelKey, labelValue string) error {

	result := database.DB.
		Where("execution_id = ? AND label_id IN (SELECT id FROM labels WHERE key = ? AND value = ?)",
			executionID, labelKey, labelValue).
		Delete(&database.ExecutionResultLabel{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove execution result label: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("label relationship '%s:%s' not found for execution %d", labelKey, labelValue, executionID)
	}

	return nil
}

// InitializeExecutionLabels initializes system labels for execution results
func InitializeExecutionLabels() error {
	// Initialize source labels
	sourceLabels := []struct {
		value       string
		description string
	}{
		{consts.ExecutionSourceManual, consts.ExecutionManualDescription},
		{consts.ExecutionSourceSystem, consts.ExecutionSystemDescription},
	}

	for _, labelInfo := range sourceLabels {
		_, err := CreateOrGetLabel(consts.ExecutionLabelSource, labelInfo.value, consts.LabelExecution, labelInfo.description)
		if err != nil {
			return fmt.Errorf("failed to initialize execution label %s=%s: %v",
				consts.ExecutionLabelSource, labelInfo.value, err)
		}
	}

	return nil
}

// GetAlgorithmDatasetEvaluation retrieves all execution results for a specific algorithm on a specific dataset
func GetAlgorithmDatasetEvaluation(req dto.AlgorithmDatasetEvaluationReq) (*dto.AlgorithmDatasetEvaluationResp, error) {
	// First, get all fault injection schedules (datapacks) for the specified dataset
	var datasetRecord database.Dataset
	err := database.DB.Where("name = ? AND status = 1", req.Dataset).First(&datasetRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset '%s' not found", req.Dataset)
		}
		return nil, fmt.Errorf("failed to query dataset: %v", err)
	}

	// Get all fault injection schedules (datapacks) in this dataset
	var faultInjections []database.FaultInjectionSchedule
	err = database.DB.
		Joins("JOIN dataset_fault_injections ON dataset_fault_injections.fault_injection_id = fault_injection_schedules.id").
		Where("dataset_fault_injections.dataset_id = ?", datasetRecord.ID).
		Find(&faultInjections).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query fault injections for dataset: %v", err)
	}

	if len(faultInjections) == 0 {
		return &dto.AlgorithmDatasetEvaluationResp{
			Algorithm:     req.Algorithm,
			Dataset:       req.Dataset,
			TotalCount:    0,
			ExecutedCount: 0,
			Items:         []dto.DatapackEvaluationItem{},
		}, nil
	}

	// Get fault injection IDs
	faultInjectionIDs := make([]int, len(faultInjections))
	for i, fi := range faultInjections {
		faultInjectionIDs[i] = fi.ID
	}

	// Query execution results for the specified algorithm and these datapacks
	query := database.DB.
		Table("execution_results").
		Select("execution_results.*, containers.name as algorithm, fault_injection_schedules.injection_name as datapack_name").
		Joins("JOIN containers ON containers.id = execution_results.algorithm_id").
		Joins("JOIN fault_injection_schedules ON fault_injection_schedules.id = execution_results.datapack_id").
		Where("containers.name = ? AND execution_results.datapack_id IN (?) AND execution_results.status = ?",
			req.Algorithm, faultInjectionIDs, consts.ExecutionSuccess)

	// Apply label filters if provided
	if len(req.LabelFilters) > 0 {
		for labelKey, labelValue := range req.LabelFilters {
			query = query.Where("execution_results.id IN (SELECT execution_id FROM execution_result_labels erl JOIN labels l ON erl.label_id = l.id WHERE l.key = ? AND l.value = ?)", labelKey, labelValue)
		}
	}

	var execResults []struct {
		database.ExecutionResult
		Algorithm    string `json:"algorithm"`
		DatapackName string `json:"datapack_name"`
	}

	if err := query.Find(&execResults).Error; err != nil {
		return nil, fmt.Errorf("failed to query execution results: %v", err)
	}

	// If no executions found, return empty result
	if len(execResults) == 0 {
		return &dto.AlgorithmDatasetEvaluationResp{
			Algorithm:     req.Algorithm,
			Dataset:       req.Dataset,
			TotalCount:    len(faultInjections),
			ExecutedCount: 0,
			Items:         []dto.DatapackEvaluationItem{},
		}, nil
	}

	// Get execution IDs for granularity results query
	executionIDs := make([]int, len(execResults))
	for i, exec := range execResults {
		executionIDs[i] = exec.ID
	}

	// Query granularity results for these executions
	var granularityResults []database.GranularityResult
	if err := database.DB.
		Where("execution_id IN (?)", executionIDs).
		Find(&granularityResults).Error; err != nil {
		return nil, fmt.Errorf("failed to query granularity results: %v", err)
	}

	// Group granularity results by execution ID
	granMap := make(map[int][]dto.GranularityRecord, len(executionIDs))
	for _, gran := range granularityResults {
		var record dto.GranularityRecord
		record.Convert(gran)

		if _, exists := granMap[gran.ExecutionID]; !exists {
			granMap[gran.ExecutionID] = []dto.GranularityRecord{record}
		} else {
			granMap[gran.ExecutionID] = append(granMap[gran.ExecutionID], record)
		}
	}

	// Get ground truth for all datapacks
	datapackNames := make([]string, len(faultInjections))
	for i, fi := range faultInjections {
		datapackNames[i] = fi.InjectionName
	}

	groundtruthMap, err := GetGroundtruthMap(datapackNames)
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth map: %v", err)
	}

	// Build response items
	items := make([]dto.DatapackEvaluationItem, 0, len(execResults))
	for _, exec := range execResults {
		item := dto.DatapackEvaluationItem{
			DatapackName: exec.DatapackName,
			ExecutionID:  exec.ID,
			Groundtruth:  groundtruthMap[exec.DatapackName],
			Predictions:  granMap[exec.ID],
			ExecutedAt:   exec.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// If no predictions found, set empty slice
		if item.Predictions == nil {
			item.Predictions = []dto.GranularityRecord{}
		}

		items = append(items, item)
	}

	return &dto.AlgorithmDatasetEvaluationResp{
		Algorithm:     req.Algorithm,
		Dataset:       req.Dataset,
		TotalCount:    len(faultInjections),
		ExecutedCount: len(items),
		Items:         items,
	}, nil
}

// GetAlgorithmDatapackEvaluation retrieves execution result for a specific algorithm on a specific datapack
func GetAlgorithmDatapackEvaluation(req dto.AlgorithmDatapackEvaluationReq) (*dto.AlgorithmDatapackEvaluationResp, error) {
	// Query execution results for the specified algorithm and datapack
	query := database.DB.
		Table("execution_results").
		Select("execution_results.*, containers.name as algorithm, fault_injection_schedules.injection_name as datapack_name").
		Joins("JOIN containers ON containers.id = execution_results.algorithm_id").
		Joins("JOIN fault_injection_schedules ON fault_injection_schedules.id = execution_results.datapack_id").
		Where("containers.name = ? AND fault_injection_schedules.injection_name = ? AND execution_results.status = ?",
			req.Algorithm, req.Datapack, consts.ExecutionSuccess)

	// Apply label filters if provided
	if len(req.LabelFilters) > 0 {
		for labelKey, labelValue := range req.LabelFilters {
			query = query.Where("execution_results.id IN (SELECT execution_id FROM execution_result_labels erl JOIN labels l ON erl.label_id = l.id WHERE l.key = ? AND l.value = ?)", labelKey, labelValue)
		}
	}

	// Order by created_at DESC to get the latest execution
	query = query.Order("execution_results.created_at DESC").Limit(1)

	var execResult struct {
		database.ExecutionResult
		Algorithm    string `json:"algorithm"`
		DatapackName string `json:"datapack_name"`
	}

	err := query.First(&execResult).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No execution found, but we can still get ground truth
			groundtruthMap, err := GetGroundtruthMap([]string{req.Datapack})
			if err != nil {
				return nil, fmt.Errorf("failed to get ground truth for datapack '%s': %v", req.Datapack, err)
			}

			groundtruth, exists := groundtruthMap[req.Datapack]
			if !exists {
				return nil, fmt.Errorf("datapack '%s' not found", req.Datapack)
			}

			return &dto.AlgorithmDatapackEvaluationResp{
				Algorithm:   req.Algorithm,
				Datapack:    req.Datapack,
				ExecutionID: 0,
				Groundtruth: groundtruth,
				Predictions: []dto.GranularityRecord{},
				ExecutedAt:  "",
				Found:       false,
			}, nil
		}
		return nil, fmt.Errorf("failed to query execution result: %v", err)
	}

	// Query granularity results for this execution
	var granularityResults []database.GranularityResult
	if err := database.DB.
		Where("execution_id = ?", execResult.ID).
		Find(&granularityResults).Error; err != nil {
		return nil, fmt.Errorf("failed to query granularity results: %v", err)
	}

	// Convert granularity results
	predictions := make([]dto.GranularityRecord, len(granularityResults))
	for i, gran := range granularityResults {
		predictions[i].Convert(gran)
	}

	// Get ground truth for this datapack
	groundtruthMap, err := GetGroundtruthMap([]string{req.Datapack})
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth map: %v", err)
	}

	groundtruth, exists := groundtruthMap[req.Datapack]
	if !exists {
		return nil, fmt.Errorf("ground truth not found for datapack '%s'", req.Datapack)
	}

	return &dto.AlgorithmDatapackEvaluationResp{
		Algorithm:   req.Algorithm,
		Datapack:    req.Datapack,
		ExecutionID: execResult.ID,
		Groundtruth: groundtruth,
		Predictions: predictions,
		ExecutedAt:  execResult.CreatedAt.Format("2006-01-02 15:04:05"),
		Found:       true,
	}, nil
}
