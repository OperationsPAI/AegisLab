package repository

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"gorm.io/gorm"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
)

const BATCH_SIZE = 500 // 服务器端分批大小

func CheckExecutionResultExists(id int) (bool, error) {
	var execution database.ExecutionResult
	if err := database.DB.Where("id = ?", id).First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check execution result: %v", err)
	}

	return true, nil
}

func CreateExecutionResult(taskID string, algorithmID, algorithmTagID, datapackID int, duration float64, labels *dto.ExecutionLabels) (int, error) {
	altorithmLabel, err := GetContainerLabel(algorithmID, algorithmTagID)
	if err != nil {
		return 0, fmt.Errorf("failed to get algorithm label: %v", err)
	}

	executionResult := &database.ExecutionResult{
		AlgorithmLabelID: altorithmLabel.ID,
		DatapackID:       datapackID,
		Duration:         duration,
		Status:           consts.ExecutionSuccess,
	}

	// Set TaskID to nil if it's empty, otherwise set the value
	if taskID != "" {
		executionResult.TaskID = &taskID
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(executionResult).Error; err != nil {
			return fmt.Errorf("failed to create execution result: %v", err)
		}

		return nil
	}); err != nil {
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

	// Add user-defined labels if provided
	if labels != nil && labels.Tag != "" {
		if err := AddExecutionResultLabel(executionResult.ID, consts.LabelKeyTag, labels.Tag, "User-defined tag"); err != nil {
			fmt.Printf("Warning: Failed to create tag label: %v\n", err)
		}
	}

	return executionResult.ID, nil
}

func UpdateExecutionResult(id int, updates map[string]any) error {
	result := database.DB.Model(&database.ExecutionResult{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("injection not found or no changes made")
	}
	return nil
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

// GetExecutionLabelsMap gets all labels for multiple execution results in batch (optimized)
func GetExecutionLabelsMap(executionIDs []int) (map[int][]database.Label, error) {
	if len(executionIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.ExecutionResultLabel
	if err := database.DB.Preload("Label").
		Where("execution_id IN ?", executionIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get execution label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.ExecutionID] = append(labelsMap[relation.ExecutionID], *relation.Label)
		}
	}

	for _, id := range executionIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
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

// GetExecutionResultLabels retrieves all labels for an execution result (optimized)
func GetExecutionResultLabels(executionID int) ([]database.Label, error) {
	labelsMap, err := GetExecutionLabelsMap([]int{executionID})
	if err != nil {
		return nil, err
	}
	return labelsMap[executionID], nil
}

// RemoveExecutionResultLabel removes a specific label from an execution result
func RemoveExecutionResultLabel(executionID int, labelKey, labelValue string) error {

	// First get the label ID
	var label database.Label
	if err := database.DB.Where("label_key = ? AND label_value = ?", labelKey, labelValue).First(&label).Error; err != nil {
		return fmt.Errorf("label '%s:%s' not found: %v", labelKey, labelValue, err)
	}

	// Then delete the relationship
	result := database.DB.
		Where("execution_id = ? AND label_id = ?", executionID, label.ID).
		Delete(&database.ExecutionResultLabel{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove execution result label: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("label relationship '%s:%s' not found for execution %d", labelKey, labelValue, executionID)
	}

	return nil
}

// applyLabelFilters applies label filters to a GORM query using native GORM joins
// func applyLabelFilters(query *gorm.DB, tag string) *gorm.DB {
// 	if tag == "" {
// 		return query
// 	}

// 	query = query.Joins("JOIN execution_result_labels erl ON erl.execution_id = execution_results.id").
// 		Joins("JOIN labels l ON l.id = erl.label_id").
// 		Where("l.label_key = ? AND l.label_value = ?", consts.LabelKeyTag, tag)
// 	return query
// }

// GetDatasetEvaluationBatch retrieves evaluation results for multiple algorithm-dataset pairs in batch
// Optimized for datasets with large number of execution records (10000+ executions, 50000+ results)
func GetDatasetEvaluationBatch(req dto.DatasetEvaluationBatchReq) (dto.DatasetEvaluationBatchResp, error) {
	if len(req.Items) == 0 {
		return dto.DatasetEvaluationBatchResp{}, nil
	}

	// Set default dataset versions
	for i := range req.Items {
		if req.Items[i].DatasetVersion == "" {
			req.Items[i].DatasetVersion = "v1.0"
		}
	}

	// 1. Batch query all required datasets
	datasetMap := make(map[string]string) // dataset_name -> version
	for _, item := range req.Items {
		datasetMap[item.Dataset] = item.DatasetVersion
	}

	var datasetConditions []string
	var datasetArgs []any
	for name, version := range datasetMap {
		datasetConditions = append(datasetConditions, "(name = ? AND version = ?)")
		datasetArgs = append(datasetArgs, name, version)
	}

	var datasetRecords []database.Dataset
	if err := database.DB.
		Where(strings.Join(datasetConditions, " OR "), datasetArgs...).
		Where("status = ?", consts.DatasetEnabled).
		Find(&datasetRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to query datasets: %v", err)
	}

	datasetLookup := make(map[string]database.Dataset)
	for _, dataset := range datasetRecords {
		key := fmt.Sprintf("%s:%s", dataset.Name, dataset.Version)
		datasetLookup[key] = dataset
	}

	// 2. Collect all unique algorithms for batch processing
	algorithmSet := make(map[string]bool)
	datasetIDSet := make(map[int]bool)
	tagSet := make(map[string]bool)

	for _, item := range req.Items {
		algorithmSet[item.Algorithm] = true
		if item.Tag != "" {
			tagSet[item.Tag] = true
		}

		datasetKey := fmt.Sprintf("%s:%s", item.Dataset, item.DatasetVersion)
		if dataset, exists := datasetLookup[datasetKey]; exists {
			datasetIDSet[dataset.ID] = true
		}
	}

	algorithms := make([]string, 0, len(algorithmSet))
	datasetIDs := make([]int, 0, len(datasetIDSet))
	tags := make([]string, 0, len(tagSet))

	for algorithm := range algorithmSet {
		algorithms = append(algorithms, algorithm)
	}
	for datasetID := range datasetIDSet {
		datasetIDs = append(datasetIDs, datasetID)
	}
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	// 3. Batch query fault injections for all datasets
	var allFaultInjections []struct {
		DatasetID        int    `json:"dataset_id"`
		FaultInjectionID int    `json:"fault_injection_id"`
		InjectionName    string `json:"injection_name"`
	}

	if err := database.DB.
		Table("dataset_fault_injections dfi").
		Select("dfi.dataset_id, dfi.fault_injection_id, fis.injection_name").
		Joins("JOIN fault_injection_schedules fis ON fis.id = dfi.fault_injection_id").
		Where("dfi.dataset_id IN ?", datasetIDs).
		Find(&allFaultInjections).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query fault injections: %v", err)
	}

	// Group fault injections by dataset ID
	datasetFaultMap := make(map[int][]string)
	for _, fi := range allFaultInjections {
		datasetFaultMap[fi.DatasetID] = append(datasetFaultMap[fi.DatasetID], fi.InjectionName)
	}

	// 4. Batch query all execution results with optimized query
	var allExecResults []struct {
		database.ExecutionResult
		Algorithm    string `json:"algorithm"`
		DatapackName string `json:"datapack_name"`
		DatasetID    int    `json:"dataset_id"`
	}

	baseQuery := database.DB.
		Table("execution_results er").
		Select("er.*, c.name as algorithm, fis.injection_name as datapack_name, dfi.dataset_id").
		Joins("JOIN containers c ON c.id = er.algorithm_id").
		Joins("JOIN fault_injection_schedules fis ON fis.id = er.datapack_id").
		Joins("JOIN dataset_fault_injections dfi ON dfi.fault_injection_id = fis.id").
		Where("c.name IN ? AND dfi.dataset_id IN ? AND er.status = ?",
			algorithms, datasetIDs, consts.ExecutionSuccess)

	if err := baseQuery.Find(&allExecResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query execution results: %v", err)
	}

	// 5. Handle tag filtering if needed
	var filteredExecResults []struct {
		database.ExecutionResult
		Algorithm    string `json:"algorithm"`
		DatapackName string `json:"datapack_name"`
		DatasetID    int    `json:"dataset_id"`
	}

	if len(tags) > 0 {
		// Get execution IDs for tag filtering
		executionIDs := make([]int, len(allExecResults))
		for i, exec := range allExecResults {
			executionIDs[i] = exec.ID
		}

		// Batch query tag labels
		var tagLabels []struct {
			ExecutionID int    `json:"execution_id"`
			LabelValue  string `json:"label_value"`
		}
		if err := database.DB.
			Table("execution_result_labels erl").
			Select("erl.execution_id, l.label_value").
			Joins("JOIN labels l ON l.id = erl.label_id").
			Where("erl.execution_id IN ? AND l.label_key = ? AND l.label_value IN ?",
				executionIDs, consts.LabelKeyTag, tags).
			Find(&tagLabels).Error; err != nil {
			return nil, fmt.Errorf("failed to query tag labels: %v", err)
		}

		execTagMap := make(map[int]string)
		for _, tagLabel := range tagLabels {
			execTagMap[tagLabel.ExecutionID] = tagLabel.LabelValue
		}

		// Filter based on request requirements
		for _, exec := range allExecResults {
			for _, item := range req.Items {
				datasetKey := fmt.Sprintf("%s:%s", item.Dataset, item.DatasetVersion)
				dataset, exists := datasetLookup[datasetKey]
				if !exists {
					continue
				}

				if exec.Algorithm == item.Algorithm && exec.DatasetID == dataset.ID {
					if item.Tag == "" || execTagMap[exec.ID] == item.Tag {
						filteredExecResults = append(filteredExecResults, exec)
						break
					}
				}
			}
		}
	} else {
		filteredExecResults = allExecResults
	}

	// 6. Collect all unique datapack names for ground truth batch query
	datapackSet := make(map[string]bool)
	for _, exec := range filteredExecResults {
		datapackSet[exec.DatapackName] = true
	}

	datapacks := make([]string, 0, len(datapackSet))
	for datapack := range datapackSet {
		datapacks = append(datapacks, datapack)
	}

	// 7. Batch query ground truth for all datapacks
	groundtruthMap, err := GetGroundtruthMap(datapacks)
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth map: %v", err)
	}

	// 8. Batch query granularity results for all executions
	executionIDs := make([]int, len(filteredExecResults))
	for i, exec := range filteredExecResults {
		executionIDs[i] = exec.ID
	}

	var allGranularityResults []database.GranularityResult
	if len(executionIDs) > 0 {
		if err := database.DB.
			Where("execution_id IN ?", executionIDs).
			Find(&allGranularityResults).Error; err != nil {
			return nil, fmt.Errorf("failed to batch query granularity results: %v", err)
		}
	}

	// Group granularity results by execution ID
	granMap := make(map[int][]dto.GranularityRecord)
	for _, gran := range allGranularityResults {
		var record dto.GranularityRecord
		record.Convert(gran)
		granMap[gran.ExecutionID] = append(granMap[gran.ExecutionID], record)
	}

	// 9. Build response for each request item
	response := make(dto.DatasetEvaluationBatchResp, len(req.Items))

	for i, item := range req.Items {
		datasetKey := fmt.Sprintf("%s:%s", item.Dataset, item.DatasetVersion)
		dataset, exists := datasetLookup[datasetKey]

		if !exists {
			return nil, fmt.Errorf("dataset %s with version %s not found at index %d", item.Dataset, item.DatasetVersion, i)
		}

		// Get fault injections count for this dataset
		totalCount := len(datasetFaultMap[dataset.ID])

		// Find matching execution results for this specific request
		var matchingExecs []struct {
			database.ExecutionResult
			Algorithm    string `json:"algorithm"`
			DatapackName string `json:"datapack_name"`
			DatasetID    int    `json:"dataset_id"`
		}

		for _, exec := range filteredExecResults {
			if exec.Algorithm == item.Algorithm && exec.DatasetID == dataset.ID {
				matchingExecs = append(matchingExecs, exec)
			}
		}

		// Find the latest execution for each datapack to avoid duplicates
		latestDatapackExecs := make(map[string]struct {
			database.ExecutionResult
			Algorithm    string `json:"algorithm"`
			DatapackName string `json:"datapack_name"`
			DatasetID    int    `json:"dataset_id"`
		})

		for _, exec := range matchingExecs {
			if existing, exists := latestDatapackExecs[exec.DatapackName]; !exists || exec.CreatedAt.After(existing.CreatedAt) {
				latestDatapackExecs[exec.DatapackName] = exec
			}
		}

		// Build evaluation items from latest executions only
		evaluationItems := make([]dto.DatapackEvaluationItem, 0, len(latestDatapackExecs))
		for _, exec := range latestDatapackExecs {
			predictions := granMap[exec.ID]
			if predictions == nil {
				predictions = []dto.GranularityRecord{}
			}

			evaluationItems = append(evaluationItems, dto.DatapackEvaluationItem{
				DatapackName:      exec.DatapackName,
				ExecutionID:       exec.ID,
				ExecutionDuration: exec.Duration,
				Groundtruth:       groundtruthMap[exec.DatapackName],
				Predictions:       predictions,
				ExecutedAt:        exec.CreatedAt,
			})
		}

		response[i] = dto.AlgorithmDatasetResp{
			Algorithm:      item.Algorithm,
			Dataset:        item.Dataset,
			DatasetVersion: item.DatasetVersion,
			TotalCount:     totalCount,
			ExecutedCount:  len(evaluationItems),
			Items:          evaluationItems,
		}
	}

	return response, nil
}

// GetDatapackEvaluationBatch retrieves the latest execution results for multiple algorithm-datapack pairs in batch
// Optimized for large batch requests (10000+ items) with server-side chunking
// GetDatapackEvaluationBatch retrieves the latest execution results for multiple algorithm-datapack pairs in batch
// Optimized for large batch requests (10000+ items) with server-side chunking
func GetDatapackEvaluationBatch(req dto.DatapackEvaluationBatchReq) (dto.DatapackEvaluationBatchResp, error) {
	if len(req.Items) == 0 {
		return dto.DatapackEvaluationBatchResp{}, nil
	}

	response := make(dto.DatapackEvaluationBatchResp, len(req.Items))

	for i := 0; i < len(req.Items); i += BATCH_SIZE {
		end := min(i+BATCH_SIZE, len(req.Items))

		// Process current batch
		batchReq := dto.DatapackEvaluationBatchReq{
			Items: req.Items[i:end],
		}

		batchResponse, err := processDatapackBatch(batchReq)
		if err != nil {
			return nil, fmt.Errorf("failed to process batch %d-%d: %v", i, end-1, err)
		}

		for j, item := range batchResponse {
			response[i+j] = item
		}
	}

	return response, nil
}

// processDatapackBatch processes a single batch of datapack evaluations
func processDatapackBatch(req dto.DatapackEvaluationBatchReq) (dto.DatapackEvaluationBatchResp, error) {
	// 1. Collect unique algorithms, datapacks and tags
	algorithmSet := make(map[string]bool)
	datapackSet := make(map[string]bool)
	tagSet := make(map[string]bool)

	for _, item := range req.Items {
		algorithmSet[item.Algorithm] = true
		datapackSet[item.Datapack] = true
		if item.Tag != "" {
			tagSet[item.Tag] = true
		}
	}

	algorithms := make([]string, 0, len(algorithmSet))
	datapacks := make([]string, 0, len(datapackSet))
	tags := make([]string, 0, len(tagSet))

	for algorithm := range algorithmSet {
		algorithms = append(algorithms, algorithm)
	}
	for datapack := range datapackSet {
		datapacks = append(datapacks, datapack)
	}
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	// 2. Batch query execution results
	var allExecResults []struct {
		database.ExecutionResult
		Algorithm    string `json:"algorithm"`
		DatapackName string `json:"datapack_name"`
	}

	baseQuery := database.DB.
		Table("execution_results").
		Select("execution_results.*, containers.name as algorithm, fault_injection_schedules.injection_name as datapack_name").
		Joins("JOIN containers ON containers.id = execution_results.algorithm_id").
		Joins("JOIN fault_injection_schedules ON fault_injection_schedules.id = execution_results.datapack_id").
		Where("containers.name IN ? AND fault_injection_schedules.injection_name IN ? AND execution_results.status = ?",
			algorithms, datapacks, consts.ExecutionSuccess)

	if err := baseQuery.Find(&allExecResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query execution results: %v", err)
	}

	// 3. Process tag filtering (if needed)
	if len(tags) > 0 {
		executionIDs := make([]int, len(allExecResults))
		for i, exec := range allExecResults {
			executionIDs[i] = exec.ID
		}

		var tagLabels []struct {
			ExecutionID int    `json:"execution_id"`
			LabelValue  string `json:"label_value"`
		}
		if err := database.DB.
			Table("execution_result_labels erl").
			Select("erl.execution_id, l.label_value").
			Joins("JOIN labels l ON l.id = erl.label_id").
			Where("erl.execution_id IN ? AND l.label_key = ? AND l.label_value IN ?",
				executionIDs, consts.LabelKeyTag, tags).
			Find(&tagLabels).Error; err != nil {
			return nil, fmt.Errorf("failed to query tag labels: %v", err)
		}

		execTagMap := make(map[int]string)
		for _, tagLabel := range tagLabels {
			execTagMap[tagLabel.ExecutionID] = tagLabel.LabelValue
		}

		// Apply tag filtering
		var filteredResults []struct {
			database.ExecutionResult
			Algorithm    string `json:"algorithm"`
			DatapackName string `json:"datapack_name"`
		}

		for _, exec := range allExecResults {
			for _, item := range req.Items {
				if exec.Algorithm == item.Algorithm && exec.DatapackName == item.Datapack {
					if item.Tag == "" || execTagMap[exec.ID] == item.Tag {
						filteredResults = append(filteredResults, exec)
						break
					}
				}
			}
		}
		allExecResults = filteredResults
	}

	// 4. Find the latest execution for each algorithm-datapack pair
	type ExecKey struct {
		Algorithm string
		Datapack  string
		Tag       string
	}

	latestExecMap := make(map[ExecKey]struct {
		database.ExecutionResult
		Algorithm    string `json:"algorithm"`
		DatapackName string `json:"datapack_name"`
	})

	for _, exec := range allExecResults {
		for _, item := range req.Items {
			if exec.Algorithm == item.Algorithm && exec.DatapackName == item.Datapack {
				key := ExecKey{
					Algorithm: item.Algorithm,
					Datapack:  item.Datapack,
					Tag:       item.Tag,
				}

				if existing, exists := latestExecMap[key]; !exists || exec.CreatedAt.After(existing.CreatedAt) {
					latestExecMap[key] = exec
				}
			}
		}
	}

	// 5. Batch query ground truth
	groundtruthMap, err := GetGroundtruthMap(datapacks)
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth map: %v", err)
	}

	// 6. Batch query granularity results
	var latestExecutionIDs []int
	for _, exec := range latestExecMap {
		latestExecutionIDs = append(latestExecutionIDs, exec.ID)
	}

	var allGranularityResults []database.GranularityResult
	if len(latestExecutionIDs) > 0 {
		if err := database.DB.
			Where("execution_id IN ?", latestExecutionIDs).
			Find(&allGranularityResults).Error; err != nil {
			return nil, fmt.Errorf("failed to batch query granularity results: %v", err)
		}
	}

	granMap := make(map[int][]dto.GranularityRecord)
	for _, gran := range allGranularityResults {
		var record dto.GranularityRecord
		record.Convert(gran)
		granMap[gran.ExecutionID] = append(granMap[gran.ExecutionID], record)
	}

	// 7. Build response
	response := make(dto.DatapackEvaluationBatchResp, len(req.Items))

	for i, item := range req.Items {
		key := ExecKey{
			Algorithm: item.Algorithm,
			Datapack:  item.Datapack,
			Tag:       item.Tag,
		}

		if exec, exists := latestExecMap[key]; exists {
			predictions := granMap[exec.ID]
			if predictions == nil {
				predictions = []dto.GranularityRecord{}
			}

			response[i] = dto.AlgorithmDatapackResp{
				Algorithm:         item.Algorithm,
				Datapack:          item.Datapack,
				ExecutionID:       exec.ID,
				ExecutionDuration: exec.Duration,
				Groundtruth:       groundtruthMap[item.Datapack],
				Predictions:       predictions,
				ExecutedAt:        exec.CreatedAt,
				Found:             exists,
			}
		} else {
			response[i] = dto.AlgorithmDatapackResp{
				Algorithm:         item.Algorithm,
				Datapack:          item.Datapack,
				ExecutionID:       0,
				ExecutionDuration: 0,
				Groundtruth:       groundtruthMap[item.Datapack],
				Predictions:       []dto.GranularityRecord{},
				ExecutedAt:        time.Time{},
				Found:             exists,
			}
		}
	}

	return response, nil
}
