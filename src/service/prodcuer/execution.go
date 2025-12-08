package producer

import (
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BatchCreateDetectorResults saves multiple detector results for a given execution
func BatchCreateDetectorResults(req *dto.UploadDetectorResultReq, executionID int) (*dto.UploadExecutionResultResp, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := updateExecutionDuration(tx, executionID, req.Duration); err != nil {
			return err
		}

		var detectorResults []database.DetectorResult
		for _, item := range req.Results {
			detectorResults = append(detectorResults, *item.ConvertToDetectorResult(executionID))
		}

		if err := repository.SaveDetectorResults(tx, detectorResults); err != nil {
			return fmt.Errorf("failed to save detector results for execution %d: %w", executionID, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	resp := &dto.UploadExecutionResultResp{
		ResultCount:  len(req.Results),
		UploadedAt:   time.Now(),
		HasAnomalies: req.HasAnomalies(),
	}
	return resp, nil
}

// BatchCreateGranularityResults saves multiple granularity results for a given execution
func BatchCreateGranularityResults(req *dto.UploadGranularityResultReq, executionID int) (*dto.UploadExecutionResultResp, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := updateExecutionDuration(tx, executionID, req.Duration); err != nil {
			return err
		}

		var granularityResults []database.GranularityResult
		for _, item := range req.Results {
			granularityResults = append(granularityResults, *item.ConvertToGranularityResult(executionID))
		}

		if err := repository.SaveGranularityResults(tx, granularityResults); err != nil {
			return fmt.Errorf("failed to save detector results for execution %d: %w", executionID, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	resp := &dto.UploadExecutionResultResp{
		ResultCount: len(req.Results),
		UploadedAt:  time.Now(),
	}
	return resp, nil
}

// BatchDeleteExecutions deletes multiple executions by their IDs
func BatchDeleteExecutionsByIDs(executionIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return batchDeleteExecutionsCore(tx, executionIDs)
	})
}

// BatchDeleteExecutionsByLabels deletes fault executions based on label conditions
func BatchDeleteExecutionsByLabels(labelItems []dto.LabelItem) error {
	if len(labelItems) == 0 {
		return nil
	}

	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, item := range labelItems {
		labelConditions = append(labelConditions, map[string]string{
			"key":   item.Key,
			"value": item.Value,
		})
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		executionIDs, err := repository.ListExecutionIDsByLabels(database.DB, labelConditions)
		if err != nil {
			return fmt.Errorf("failed to list execution ids by labels: %w", err)
		}

		return batchDeleteExecutionsCore(tx, executionIDs)
	})
}

// GetExecutionDetail retrieves detailed information about a specific execution
func GetExecutionDetail(executionID int) (*dto.ExecutionDetailResp, error) {
	execution, err := repository.GetExecutionByID(database.DB, executionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: execution id: %d", consts.ErrNotFound, executionID)
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	labels, err := repository.ListLabelsByExecutionID(database.DB, execution.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution labels: %w", err)
	}

	resp := dto.NewExecutionDetailResp(execution, labels)

	if execution.AlgorithmVersion.Container.Name == config.GetString(consts.DetectorKey) {
		detectorResults, err := repository.ListDetectorResultsByExecutionID(database.DB, execution.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get detector results: %w", err)
		}

		items := make([]dto.DetectorResultItem, 0, len(detectorResults))
		for _, result := range detectorResults {
			items = append(items, dto.NewDetectorResultItem(&result))
		}

		resp.DetectorResults = items
	} else {
		granularityResults, err := repository.ListGranularityResultsByExecutionID(database.DB, execution.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get granularity results: %w", err)
		}

		items := make([]dto.GranularityResultItem, 0, len(granularityResults))
		for _, result := range granularityResults {
			items = append(items, dto.NewGranularityResultItem(&result))
		}

		resp.GranularityResults = items
	}

	return resp, err
}

// ListExecutions lists executions based on the provided request parameters
func ListExecutions(req *dto.ListExecutionReq) (*dto.ListResp[dto.ExecutionResp], error) {
	limit, offset := req.ToGormParams()

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	executions, total, err := repository.ListExecutions(database.DB, limit, offset, req.State, req.Status, labelConditions)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}

	executionIDs := make([]int, 0, len(executions))
	for _, execution := range executions {
		executionIDs = append(executionIDs, execution.ID)
	}

	labelsMap, err := repository.ListExecutionLabels(database.DB, executionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list execution labels: %w", err)
	}

	executionResps := make([]dto.ExecutionResp, 0, len(executions))
	for _, execution := range executions {
		labels := labelsMap[execution.ID]
		executionResps = append(executionResps, *dto.NewExecutionResp(&execution, labels))
	}

	resp := dto.ListResp[dto.ExecutionResp]{
		Items:      executionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// ListAvaliableExecutionLabels lists all available labels for executions
func ListAvaliableExecutionLabels() ([]dto.LabelItem, error) {
	labelsMap, err := repository.ListLabelsGroupByCategory(database.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels grouped by category: %w", err)
	}

	if _, exists := labelsMap[consts.ExecutionCategory]; !exists {
		return []dto.LabelItem{}, nil
	}

	labels := labelsMap[consts.ExecutionCategory]
	labelItems := make([]dto.LabelItem, 0, len(labels))
	for _, label := range labels {
		labelItems = append(labelItems, dto.LabelItem{
			Key:   label.Key,
			Value: label.Value,
		})
	}

	return labelItems, nil
}

// ManageExecutionLabels adds or removes labels for a specific execution
func ManageExecutionLabels(req *dto.ManageExecutionLabelReq, executionID int) (*dto.ExecutionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage execution labels request is nil")
	}

	var managedExecution *database.Execution
	var managedLabels []database.Label
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		execution, err := repository.GetExecutionByID(database.DB, executionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: execution id: %d", consts.ErrNotFound, executionID)
			}
			return fmt.Errorf("failed to get execution: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.ExecutionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}

			if err := repository.AddExecutionLabels(tx, execution.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add execution labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndExecutionID(tx, execution.ID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) == 0 {
				if err := repository.ClearExecutionLabels(tx, []int{executionID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear execution labels: %w", err)
				}

				if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		labels, err := repository.ListLabelsByExecutionID(database.DB, executionID)
		if err != nil {
			return fmt.Errorf("failed to get execution labels: %w", err)
		}

		managedExecution = execution
		managedLabels = labels
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewExecutionResp(managedExecution, managedLabels), nil
}

// ProduceAlgorithmExeuctionTasks produces execution tasks into Redis based on the submission request
func ProduceAlgorithmExeuctionTasks(ctx context.Context, req *dto.SubmitExecutionReq, groupID string, userID int) (*dto.SubmitExecutionResp, error) {
	project, err := repository.GetProjectByName(database.DB, req.ProjectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	refs := make([]*dto.ContainerRef, 0, len(req.Specs))
	for _, spec := range req.Specs {
		refs = append(refs, &spec.Algorithm.ContainerRef)
	}

	algorithmVersionResults, err := common.MapRefsToContainerVersions(refs, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}
	if len(algorithmVersionResults) == 0 {
		return nil, fmt.Errorf("no valid algorithm versions found for the provided specs")
	}

	var allExecutionItems []dto.SubmitExecutionItem
	for idx, spec := range req.Specs {
		datapacks, datasetID, err := extractDatapacks(database.DB, spec.Datapack, spec.Dataset, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to extract datapacks: %w", err)
		}

		algorithmVersion, exists := algorithmVersionResults[refs[idx]]
		if !exists {
			return nil, fmt.Errorf("algorithm version not found for %v", spec.Algorithm)
		}

		var executionItems []dto.SubmitExecutionItem
		for _, datapack := range datapacks {
			if datapack.StartTime == nil || datapack.EndTime == nil {
				return nil, fmt.Errorf("datapack %s does not have valid start_time and end_time", datapack.Name)
			}

			algorithmItem := dto.NewContainerVersionItem(&algorithmVersion)
			envVars, err := common.ListContainerVersionEnvVars(spec.Algorithm.EnvVars, &algorithmVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to list algorithm env vars: %w", err)
			}

			algorithmItem.EnvVars = envVars

			payload := map[string]any{
				consts.ExecuteAlgorithm:        algorithmItem,
				consts.ExecuteDatapack:         dto.NewInjectionItem(&datapack),
				consts.ExecuteDatasetVersionID: utils.GetIntValue(datasetID, consts.DefaultInvalidID),
				consts.ExecuteLabels:           req.Labels,
			}

			task := &dto.UnifiedTask{
				Type:      consts.TaskTypeRunAlgorithm,
				Immediate: true,
				Payload:   payload,
				GroupID:   groupID,
				ProjectID: project.ID,
				UserID:    userID,
				State:     consts.TaskPending,
			}
			task.SetGroupCtx(ctx)

			err = common.SubmitTask(ctx, task)
			if err != nil {
				return nil, fmt.Errorf("failed to submit task: %w", err)
			}

			executionItem := dto.SubmitExecutionItem{
				Index:              idx,
				TraceID:            task.TraceID,
				TaskID:             task.TaskID,
				AlgorithmID:        algorithmVersion.ContainerID,
				AlgorithmVersionID: algorithmVersion.ID,
				DatapackID:         &datapack.ID,
			}
			executionItems = append(executionItems, executionItem)
		}

		allExecutionItems = append(allExecutionItems, executionItems...)
	}

	resp := &dto.SubmitExecutionResp{
		GroupID: groupID,
		Items:   allExecutionItems,
	}
	return resp, nil
}

// extractDatapacks extracts datapacks based on the provided datapack name or dataset ref
func extractDatapacks(db *gorm.DB, datapackName *string, datasetRef *dto.DatasetRef, userID int) ([]database.FaultInjection, *int, error) {
	if datapackName != nil {
		datapack, err := repository.GetInjectionByName(db, *datapackName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get datapack: %w", err)
		}

		if datapack.State != consts.DatapackDetectorSuccess {
			return nil, nil, fmt.Errorf("datapack %s is not ready for detector execution", datapack.Name)
		}

		labels, err := repository.ListLabelsByInjectionID(db, datapack.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get datapack labels: %s", err.Error())
		}

		if exists := checkLabelKeyValue(labels, consts.LabelKeyTag, consts.DetectorNoAnomaly); exists {
			return nil, nil, fmt.Errorf("cannot execute detector algorithm on no_anomaly datapack: %s", datapack.Name)
		}

		return []database.FaultInjection{*datapack}, nil, nil
	}

	if datasetRef != nil {
		datasetVersionResults, err := common.MapRefsToDatasetVersions([]*dto.DatasetRef{datasetRef}, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset versions: %w", err)
		}

		version, exists := datasetVersionResults[datasetRef]
		if !exists {
			return nil, nil, fmt.Errorf("dataset version not found for %v", datasetRef)
		}

		datapacks, err := repository.ListInjectionsByDatasetVersionID(db, version.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset datapacks: %s", err.Error())
		}

		if len(datapacks) == 0 {
			return nil, nil, fmt.Errorf("dataset contains no datapacks")
		}

		datapackIDs := make([]int, 0, len(datapacks))
		for _, dp := range datapacks {
			datapackIDs = append(datapackIDs, dp.ID)
		}

		labelsMap, err := repository.ListInjectionLabels(db, datapackIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get datapack labels map: %s", err.Error())
		}

		for _, dp := range datapacks {
			if dp.State != consts.DatapackDetectorSuccess {
				return nil, nil, fmt.Errorf("datapack %s is not ready for detector execution", dp.Name)
			}
			if _, exists := labelsMap[dp.ID]; !exists {
				return nil, nil, fmt.Errorf("failed to get labels for datapack ID: %d", dp.ID)
			}
			if exists := checkLabelKeyValue(labelsMap[dp.ID], consts.LabelKeyTag, consts.DetectorNoAnomaly); exists {
				return nil, nil, fmt.Errorf("cannot execute detector algorithm on no_anomaly datapack: %s", dp.Name)
			}
		}

		return datapacks, &version.ID, nil
	}

	return nil, nil, fmt.Errorf("either datapack or dataset must be specified")
}

// updateExecutionDuration updates the duration of an execution
func updateExecutionDuration(db *gorm.DB, executionID int, duration float64) error {
	execution, err := repository.GetExecutionByID(db, executionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: execution %d not found", consts.ErrNotFound, executionID)
		}
		return fmt.Errorf("execution %d not found: %w", executionID, err)
	}

	if execution.Status != consts.CommonEnabled {
		return fmt.Errorf("must upload results for an active execution %d", executionID)
	}

	if execution.State == consts.ExecutionSuccess {
		return fmt.Errorf("cannot upload results for a successful execution %d", executionID)
	}

	if err := repository.UpdateExecution(db, executionID, map[string]any{
		"duration": duration,
	}); err != nil {
		return fmt.Errorf("failed to update execution %d duration: %w", executionID, err)
	}

	return nil
}

// batchDeleteExecutionsCore is the core logic for batch deleting executions
func batchDeleteExecutionsCore(db *gorm.DB, executionIDs []int) error {
	if err := repository.RemoveLabelsFromExecutions(db, executionIDs); err != nil {
		return fmt.Errorf("failed to delete execution labels: %w", err)
	}

	if err := repository.BatchDeleteExecutions(db, executionIDs); err != nil {
		return fmt.Errorf("failed to batch delete executions: %w", err)
	}

	return nil
}
