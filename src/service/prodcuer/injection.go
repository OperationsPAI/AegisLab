package producer

import (
	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// injectionProcessItem represents a batch of parallel fault injections
type injectionProcessItem struct {
	index         int          // Batch index in the original request
	faultDuration int          // Maximum duration among all faults in this batch
	nodes         []chaos.Node // Multiple fault nodes to be injected in parallel
	executeTime   time.Time    // Execution time for this batch
}

// BatchDeleteInjectionsByIDs deletes fault injections based on their IDs
func BatchDeleteInjectionsByIDs(injectionIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return batchDeleteExecutionsCore(tx, injectionIDs)
	})
}

// BatchDeleteInjectionsByLabels deletes fault injections based on label conditions
func BatchDeleteInjectionsByLabels(labelItems []dto.LabelItem) error {
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
		injectionIDs, err := repository.ListInjectionIDsByLabels(database.DB, labelConditions)
		if err != nil {
			return fmt.Errorf("failed to list injection ids by labels: %w", err)
		}

		return batchDeleteInjectionsCore(tx, injectionIDs)
	})
}

// CreateInjection creates a new fault injection along with its associated project-container relationships and labels
func CreateInjection(injection *database.FaultInjection, labelItems []dto.LabelItem) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.CreateInjection(tx, injection); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: injection with name %s already exists", consts.ErrAlreadyExists, injection.Name)
			}
			return fmt.Errorf("failed to create injection: %w", err)
		}

		if len(labelItems) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, labelItems, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			// Collect label IDs
			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}

			// AddInjectionLabels now takes injectionID and labelIDs (stores as TaskLabel internally)
			if err := repository.AddInjectionLabels(tx, injection.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}

		return nil
	})
}

// GetInjectionDetail retrieves detailed information about a specific fault injection
func GetInjectionDetail(injectionID int) (*dto.InjectionDetailResp, error) {
	logrus.WithField("injectionID", injectionID).Info("GetInjectionDetail: starting")

	injection, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"injectionID": injectionID,
		}).Error("failed to get injection from repository: %w", err)

		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"injectionID":   injectionID,
		"injectionName": injection.Name,
	}).Info("GetInjectionDetail: fetched injection from repository")

	labels, err := repository.ListLabelsByInjectionID(database.DB, injection.ID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"injectionID": injectionID,
		}).Error("GetInjectionDetail: failed to get injection labels: %w", err)
		return nil, fmt.Errorf("failed to get injection labels: %w", err)
	}

	injection.Labels = labels
	resp := dto.NewInjectionDetailResp(injection)

	logrus.WithField("injectionID", injectionID).Info("GetInjectionDetail: completed successfully")
	return resp, err
}

// ListInjections lists fault injections based on the provided filters
func ListInjections(req *dto.ListInjectionReq) (*dto.ListResp[dto.InjectionResp], error) {
	limit, offset := req.ToGormParams()
	fitlerOptions := req.ToFilterOptions()

	injections, total, err := repository.ListInjections(database.DB, limit, offset, fitlerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list injections: %w", err)
	}

	injectionIDs := make([]int, 0, len(injections))
	for _, injection := range injections {
		injectionIDs = append(injectionIDs, injection.ID)
	}

	labelsMap, err := repository.ListInjectionLabels(database.DB, injectionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list injection labels: %w", err)
	}

	injectionResps := make([]dto.InjectionResp, 0, len(injections))
	for _, injection := range injections {
		if labels, exists := labelsMap[injection.ID]; exists {
			injection.Labels = labels
		}
		injectionResps = append(injectionResps, *dto.NewInjectionResp(&injection))
	}

	resp := dto.ListResp[dto.InjectionResp]{
		Items:      injectionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// SearchInjections performs advanced search on fault injections
func SearchInjections(req *dto.SearchInjectionReq) (*dto.SearchResp[dto.InjectionDetailResp], error) {
	if req == nil {
		return nil, fmt.Errorf("search injection request is nil")
	}

	searchReq := req.ConvertToSearchReq()

	injections, total, err := repository.ExecuteSearch(database.DB, searchReq, database.FaultInjection{})
	if err != nil {
		return nil, fmt.Errorf("failed to search injections: %w", err)
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		labelConditions = append(labelConditions, map[string]string{
			"key":   item.Key,
			"value": item.Value,
		})
	}

	filteredInjections := []database.FaultInjection{}
	if len(labelConditions) > 0 {
		injectionIDs, err := repository.ListInjectionIDsByLabels(database.DB, labelConditions)
		if err != nil {
			return nil, fmt.Errorf("failed to list injection ids by labels: %w", err)
		}

		injectionIDMap := make(map[int]struct{}, len(injectionIDs))
		for _, id := range injectionIDs {
			injectionIDMap[id] = struct{}{}
		}

		for _, injection := range injections {
			if _, exists := injectionIDMap[injection.ID]; exists {
				filteredInjections = append(filteredInjections, injection)
			}
		}
	} else {
		filteredInjections = injections
	}

	// Convert to response format
	injectionResps := make([]dto.InjectionDetailResp, 0, len(filteredInjections))
	for _, injection := range filteredInjections {
		injectionResps = append(injectionResps, *dto.NewInjectionDetailResp(&injection))
	}

	resp := &dto.SearchResp[dto.InjectionDetailResp]{
		Items:      injectionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}

	return resp, nil
}

// ListInjectionsNoissues handles the request to list fault injections without issues
func ListInjectionsNoIssues(req *dto.ListInjectionNoIssuesReq) ([]dto.InjectionNoIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := repository.ListInjectionsNoIssues(database.DB, labelConditions, &opts.CustomStartTime, &opts.CustomEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	var items []dto.InjectionNoIssuesResp
	for i, record := range records {
		resp, err := dto.NewInjectionNoIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp at index %d: %w", i, err)
		}

		items = append(items, *resp)
	}

	return items, nil
}

// ListInjectionsNoissues handles the request to list fault injections without issues
func ListInjectionsWithIssues(req *dto.ListInjectionWithIssuesReq) ([]dto.InjectionWithIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := repository.ListInjectionsWithIssues(database.DB, labelConditions, &opts.CustomStartTime, &opts.CustomEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	var items []dto.InjectionWithIssuesResp
	for _, record := range records {
		resp, err := dto.NewInjectionWithIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp: %w", err)
		}

		items = append(items, *resp)
	}

	return items, nil
}

// ManageInjectionTags manages labels associated with a fault injection
func ManageInjectionLabels(req *dto.ManageInjectionLabelReq, injectionID int) (*dto.InjectionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage injection labels request is nil")
	}

	var managedInjection *database.FaultInjection

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByID(database.DB, injectionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
			}
			return fmt.Errorf("failed to get injection: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			// Collect label IDs
			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}

			// AddInjectionLabels now takes injectionID and labelIDs (stores as TaskLabel internally)
			if err := repository.AddInjectionLabels(tx, injection.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndInjectionID(tx, injection.ID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) > 0 {
				if err := repository.ClearInjectionLabels(tx, []int{injectionID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear injection labels: %w", err)
				}

				if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		labels, err := repository.ListLabelsByInjectionID(database.DB, injectionID)
		if err != nil {
			return fmt.Errorf("failed to get injection labels: %w", err)
		}

		injection.Labels = labels
		managedInjection = injection
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewInjectionResp(managedInjection), nil
}

// BatchManageInjectionLabels adds or removes labels from multiple injections
// Each injection can have its own set of label operations
func BatchManageInjectionLabels(req *dto.BatchManageInjectionLabelReq) (*dto.BatchManageInjectionLabelResp, error) {
	if req == nil {
		return nil, fmt.Errorf("batch manage injection labels request is nil")
	}

	resp := &dto.BatchManageInjectionLabelResp{
		FailedCount:  0,
		FailedItems:  []string{},
		SuccessCount: 0,
		SuccessItems: []dto.InjectionResp{},
	}

	if len(req.Items) == 0 {
		return resp, nil
	}

	// Process all operations in a single transaction
	return resp, database.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Collect all injection IDs and verify they exist (batch query)
		allInjectionIDs := make([]int, 0, len(req.Items))
		operationMap := make(map[int]*dto.InjectionLabelOperation)

		for i := range req.Items {
			item := &req.Items[i]
			allInjectionIDs = append(allInjectionIDs, item.InjectionID)
			operationMap[item.InjectionID] = item
		}

		injections, err := repository.ListFaultInjectionsByID(tx, allInjectionIDs)
		if err != nil {
			return fmt.Errorf("failed to list injections: %w", err)
		}

		foundIDMap := make(map[int]*database.FaultInjection)
		for i := range injections {
			foundIDMap[injections[i].ID] = &injections[i]
		}

		// Track which IDs were not found
		validIDs := make([]int, 0, len(foundIDMap))
		for _, id := range allInjectionIDs {
			if _, found := foundIDMap[id]; !found {
				resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d not found", id))
				resp.FailedCount++
				delete(operationMap, id) // Remove from operations
			} else {
				validIDs = append(validIDs, id)
			}
		}

		if len(validIDs) == 0 {
			return fmt.Errorf("no valid injection IDs found")
		}

		// Step 2: Collect all unique labels from all operations and create them in batch
		allAddLabels := make([]dto.LabelItem, 0)
		allRemoveLabels := make([]dto.LabelItem, 0)
		labelKeySet := make(map[string]bool)

		for _, op := range operationMap {
			for _, label := range op.AddLabels {
				key := label.Key + ":" + label.Value
				if !labelKeySet[key] {
					labelKeySet[key] = true
					allAddLabels = append(allAddLabels, label)
				}
			}
			for _, label := range op.RemoveLabels {
				key := label.Key + ":" + label.Value
				if !labelKeySet[key] {
					labelKeySet[key] = true
					allRemoveLabels = append(allRemoveLabels, label)
				}
			}
		}

		// Create or update all labels in batch
		var labelMap map[string]int // key:value -> label_id
		if len(allAddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, allAddLabels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			labelMap = make(map[string]int)
			for _, label := range labels {
				key := label.Key + ":" + label.Value
				labelMap[key] = label.ID
			}
		}

		// Get label IDs for removal labels
		var removeLabelMap map[string]int // key:value -> label_id
		if len(allRemoveLabels) > 0 {
			labelConditions := make([]map[string]string, 0, len(allRemoveLabels))
			for _, item := range allRemoveLabels {
				labelConditions = append(labelConditions, map[string]string{
					"key":   item.Key,
					"value": item.Value,
				})
			}

			labelIDs, err := repository.ListLabelIDsByConditions(tx, labelConditions, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to find labels to remove: %w", err)
			}

			// Map them back for quick lookup
			if len(labelIDs) > 0 {
				labels, err := repository.ListLabelsByID(tx, labelIDs)
				if err != nil {
					return fmt.Errorf("failed to list labels by IDs: %w", err)
				}

				removeLabelMap = make(map[string]int)
				for _, label := range labels {
					key := label.Key + ":" + label.Value
					removeLabelMap[key] = label.ID
				}
			}
		}

		// Step 3: Process each injection's operations
		for _, injectionID := range validIDs {
			op := operationMap[injectionID]

			if len(op.AddLabels) > 0 {
				labelIDsToAdd := make([]int, 0, len(op.AddLabels))
				for _, label := range op.AddLabels {
					key := label.Key + ":" + label.Value
					if labelID, exists := labelMap[key]; exists {
						labelIDsToAdd = append(labelIDsToAdd, labelID)
					}
				}

				if len(labelIDsToAdd) > 0 {
					if err := repository.AddInjectionLabels(tx, injectionID, labelIDsToAdd); err != nil {
						resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d: failed to add labels - %s", injectionID, err.Error()))
						resp.FailedCount++
						delete(foundIDMap, injectionID)
						continue
					}
				}
			}

			if len(op.RemoveLabels) > 0 && removeLabelMap != nil {
				labelIDsToRemove := make([]int, 0, len(op.RemoveLabels))
				for _, label := range op.RemoveLabels {
					key := label.Key + ":" + label.Value
					if labelID, exists := removeLabelMap[key]; exists {
						labelIDsToRemove = append(labelIDsToRemove, labelID)
					}
				}

				if len(labelIDsToRemove) > 0 {
					if err := repository.ClearInjectionLabels(tx, []int{injectionID}, labelIDsToRemove); err != nil {
						resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d: failed to remove labels - %s", injectionID, err.Error()))
						resp.FailedCount++
						delete(foundIDMap, injectionID)
						continue
					}
				}
			}
		}

		// Step 4: Fetch updated injection data with labels (batch query)
		if len(foundIDMap) > 0 {
			successIDs := make([]int, 0, len(foundIDMap))
			for id := range foundIDMap {
				successIDs = append(successIDs, id)
			}

			updatedInjections, err := repository.ListFaultInjectionsByID(tx, successIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch updated injections: %w", err)
			}

			labelsMap, err := repository.ListInjectionLabels(tx, successIDs)
			if err != nil {
				return fmt.Errorf("failed to list injection labels: %w", err)
			}

			for i := range updatedInjections {
				injection := &updatedInjections[i]
				if labels, exists := labelsMap[injection.ID]; exists {
					injection.Labels = labels
				}
				injectionResp := dto.NewInjectionResp(injection)
				resp.SuccessItems = append(resp.SuccessItems, *injectionResp)
				resp.SuccessCount++
			}
		}

		return nil
	})
}

// ProduceRestartPedestalTasks produces pedestal restart tasks with support for parallel fault injection
func ProduceRestartPedestalTasks(ctx context.Context, req *dto.SubmitInjectionReq, groupID string, userID int) (*dto.SubmitInjectionResp, error) {
	project, err := repository.GetProjectByName(database.DB, req.ProjectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	pedestalVersionResults, err := common.MapRefsToContainerVersions([]*dto.ContainerRef{&req.Pedestal.ContainerRef}, consts.ContainerTypePedestal, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map pedestal container ref to version: %w", err)
	}

	pedestalVersion, exists := pedestalVersionResults[&req.Pedestal.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("pedestal version not found for container: %s (version: %s)", req.Pedestal.Name, req.Pedestal.Version)
	}

	helmConfig, err := repository.GetHelmConfigByContainerVersionID(database.DB, pedestalVersion.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: helm config not found for pedestal version id %d", consts.ErrNotFound, pedestalVersion.ID)
		}
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}

	params := flattenYAMLToParameters(req.Pedestal.Payload, "")
	helmValues, err := common.ListHelmConfigValues(params, helmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to render pedestal helm values: %w", err)
	}

	pedestalInfo := dto.NewPedestalInfo(helmConfig)
	pedestalInfo.HelmConfig.Values = helmValues

	pedestalItem := dto.NewContainerVersionItem(&pedestalVersion)
	pedestalItem.Extra = pedestalInfo

	benchmarkVersionResults, err := common.MapRefsToContainerVersions([]*dto.ContainerRef{&req.Benchmark.ContainerRef}, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map benchmark container ref to version: %w", err)
	}

	benchmarkVersion, exists := benchmarkVersionResults[&req.Benchmark.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("benchmark version not found for container: %s (version: %s)", req.Benchmark.Name, req.Benchmark.Version)
	}

	benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
	envVars, err := common.ListContainerVersionEnvVars(req.Benchmark.EnvVars, &benchmarkVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
	}

	benchmarkVersionItem.EnvVars = envVars

	// Parse each batch and collect items
	processedItems := make([]injectionProcessItem, 0, len(req.Specs))
	for i := range req.Specs {
		item, err := parseBatchInjectionSpecs(ctx, i, req.Specs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse injection spec batch %d: %w", i, err)
		}
		processedItems = append(processedItems, *item)
	}

	// Remove duplicated batches
	uniqueItems, err := removeDuplicated(processedItems)
	if err != nil {
		return nil, fmt.Errorf("failed to remove duplicated batches: %w", err)
	}

	if len(req.Algorithms) > 0 {
		refs := make([]*dto.ContainerRef, 0, len(req.Algorithms))
		for i := range req.Algorithms {
			refs = append(refs, &req.Algorithms[i].ContainerRef)
		}

		algorithmVersionsResults, err := common.MapRefsToContainerVersions(refs, consts.ContainerTypeAlgorithm, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
		}

		var algorithmVersionItems []dto.ContainerVersionItem
		for i := range req.Algorithms {
			spec := &req.Algorithms[i]
			algorithmVersion, exists := algorithmVersionsResults[&spec.ContainerRef]
			if !exists {
				return nil, fmt.Errorf("algorithm version not found for %v", spec)
			}

			algorithmVersionItem := dto.NewContainerVersionItem(&algorithmVersion)
			envVars, err := common.ListContainerVersionEnvVars(spec.EnvVars, &algorithmVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to list algorithm env vars: %w", err)
			}

			algorithmVersionItem.EnvVars = envVars
			algorithmVersionItems = append(algorithmVersionItems, algorithmVersionItem)
		}

		if len(algorithmVersionItems) > 0 {
			if err := client.SetHashField(ctx, consts.InjectionAlgorithmsKey, groupID, algorithmVersionItems); err != nil {
				return nil, fmt.Errorf("failed to store injection algorithms: %w", err)
			}
		}
	}

	injectionItems := make([]dto.SubmitInjectionItem, 0, len(uniqueItems))
	for _, item := range uniqueItems {
		payload := map[string]any{
			consts.RestartPedestal:      pedestalItem,
			consts.RestartHelmConfig:    helmConfig,
			consts.RestartIntarval:      req.Interval,
			consts.RestartFaultDuration: item.faultDuration,
			consts.RestartInjectPayload: map[string]any{
				consts.InjectBenchmark:   benchmarkVersionItem,
				consts.InjectPreDuration: req.PreDuration,
				consts.InjectNodes:       item.nodes,
				consts.InjectLabels:      req.Labels,
			},
		}

		task := &dto.UnifiedTask{
			Type:        consts.TaskTypeRestartPedestal,
			Immediate:   false,
			ExecuteTime: item.executeTime.Unix(),
			Payload:     payload,
			GroupID:     groupID,
			ProjectID:   project.ID,
			UserID:      userID,
			State:       consts.TaskPending,
		}
		task.SetGroupCtx(ctx)

		err := common.SubmitTask(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("failed to submit fault injection task: %w", err)
		}

		injectionItems = append(injectionItems, dto.SubmitInjectionItem{
			Index:   item.index,
			TraceID: task.TraceID,
			TaskID:  task.TaskID,
		})
	}

	return &dto.SubmitInjectionResp{
		GroupID:         groupID,
		Items:           injectionItems,
		DuplicatedCount: len(processedItems) - len(uniqueItems),
		OriginalCount:   len(processedItems),
	}, nil
}

// ProduceDatapackBuildingTasks produces datapack building tasks into Redis based on the request specifications
func ProduceDatapackBuildingTasks(ctx context.Context, req *dto.SubmitDatapackBuildingReq, groupID string, userID int) (*dto.SubmitDatapackBuildingResp, error) {
	project, err := repository.GetProjectByName(database.DB, req.ProjectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	refs := make([]*dto.ContainerRef, 0, len(req.Specs))
	for _, spec := range req.Specs {
		refs = append(refs, &spec.Benchmark.ContainerRef)
	}

	benchmarkVersionResults, err := common.MapRefsToContainerVersions(refs, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	var allBuildingItems []dto.SubmitBuildingItem
	for idx, spec := range req.Specs {
		datapacks, datasetVersionID, err := extractDatapacks(database.DB, spec.Datapack, spec.Dataset, userID, consts.TaskTypeBuildDatapack)
		if err != nil {
			return nil, fmt.Errorf("failed to extract datapacks: %w", err)
		}

		benchmarkVersion, exists := benchmarkVersionResults[refs[idx]]
		if !exists {
			return nil, fmt.Errorf("benchmark version not found for %v", spec.Benchmark)
		}

		benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
		envVars, err := common.ListContainerVersionEnvVars(spec.Benchmark.EnvVars, &benchmarkVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
		}

		benchmarkVersionItem.EnvVars = envVars

		var buildingItems []dto.SubmitBuildingItem
		for _, datapack := range datapacks {
			if datapack.StartTime == nil || datapack.EndTime == nil {
				return nil, fmt.Errorf("datapack %s does not have valid start_time and end_time", datapack.Name)
			}

			payload := map[string]any{
				consts.BuildBenchmark:        benchmarkVersionItem,
				consts.BuildDatapack:         dto.NewInjectionItem(&datapack),
				consts.BuildDatasetVersionID: datasetVersionID,
				consts.BuildLabels:           req.Labels,
			}

			task := &dto.UnifiedTask{
				Type:      consts.TaskTypeBuildDatapack,
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
				return nil, fmt.Errorf("failed to submit datapack building task: %w", err)
			}

			buildingItems = append(buildingItems, dto.SubmitBuildingItem{
				Index:   idx,
				TraceID: task.TraceID,
				TaskID:  task.TaskID,
			})
		}

		allBuildingItems = append(allBuildingItems, buildingItems...)
	}

	resp := &dto.SubmitDatapackBuildingResp{
		GroupID: groupID,
		Items:   allBuildingItems,
	}
	return resp, nil
}

func batchDeleteInjectionsCore(db *gorm.DB, injectionIDs []int) error {
	executions, err := repository.ListExecutionsByDatapackIDs(db, injectionIDs)
	if err != nil {
		return fmt.Errorf("failed to list executions by datapack ids: %w", err)
	}

	if len(executions) == 0 {
		return fmt.Errorf("no executions found for the given injection ids")
	}

	executionIDs := make([]int, 0, len(executions))
	for _, execution := range executions {
		executionIDs = append(executionIDs, execution.ID)
	}

	if err := batchDeleteExecutionsCore(db, executionIDs); err != nil {
		return fmt.Errorf("failed to batch delete executions: %v", err)
	}

	if err := repository.ClearInjectionLabels(db, injectionIDs, nil); err != nil {
		return fmt.Errorf("failed to clear injection labels: %w", err)
	}

	if err := repository.BatchDeleteInjections(db, injectionIDs); err != nil {
		return fmt.Errorf("failed to delete injections: %w", err)
	}

	return nil
}

// parseBatchInjectionSpecs parses a single batch of fault injection specifications for parallel execution
func parseBatchInjectionSpecs(ctx context.Context, batchIndex int, specs []chaos.Node) (*injectionProcessItem, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("empty fault injection batch at index %d", batchIndex)
	}

	// Extract fault duration - use the maximum duration among all faults in the batch
	maxDuration := 0
	nodes := make([]chaos.Node, 0, len(specs))

	for idx, spec := range specs {
		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, fmt.Errorf("failed to find key %d in the children at batch %d at index %d", spec.Value, batchIndex, idx)
		}

		if len(childNode.Children) < 3 {
			return nil, fmt.Errorf("no child nodes found for fault spec at batch %d at index %d", batchIndex, idx)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value
		if faultDuration > maxDuration {
			maxDuration = faultDuration
		}

		nodes = append(nodes, spec)
	}

	uniqueServices := make(map[string]int, len(nodes))
	for idx, node := range nodes {
		conf, err := chaos.NodeToStruct[chaos.InjectionConf](ctx, &node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to InjectionConf at batch %d at index %d: %w", batchIndex, idx, err)
		}

		groundtruth, err := conf.GetGroundtruth(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get groundtruth from InjectionConf at batch %d at index %d: %w", batchIndex, idx, err)
		}

		for _, service := range groundtruth.Service {
			if service != "" {
				if oldIdx, exists := uniqueServices[service]; exists {
					return nil, fmt.Errorf("duplicated service %s found in batch %d at indexes %d and %d", service, batchIndex, oldIdx, idx)
				}
				uniqueServices[service] = idx
			}
		}
	}

	// Sort nodes to ensure consistent ordering
	nodes = sortNodes(nodes)

	return &injectionProcessItem{
		index:         batchIndex,
		faultDuration: maxDuration,
		nodes:         nodes,
	}, nil
}

// flattenYAMLToParameters converts nested YAML map to flat parameter specs
func flattenYAMLToParameters(data map[string]any, prefix string) []dto.ParameterSpec {
	var params []dto.ParameterSpec

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			// Recursively flatten nested structures
			params = append(params, flattenYAMLToParameters(v, fullKey)...)
		case []any:
			// Convert array to JSON string
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				logrus.Warnf("Failed to marshal array for key %s: %v", fullKey, err)
				continue
			}
			params = append(params, dto.ParameterSpec{
				Key:   fullKey,
				Value: string(jsonBytes),
			})
		default:
			// Primitive values (string, int, bool, etc.)
			params = append(params, dto.ParameterSpec{
				Key:   fullKey,
				Value: v,
			})
		}
	}

	return params
}

// removeDuplicated filters out batches that already exist in DB and removes duplicates within the request
func removeDuplicated(items []injectionProcessItem) ([]injectionProcessItem, error) {
	engineConfigStrs := make([]string, len(items))
	for i, item := range items {
		if len(item.nodes) == 0 {
			engineConfigStrs[i] = ""
			continue
		}

		// Marshal the entire batch of nodes as the engine config
		b, err := json.Marshal(item.nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal engine config at batch index %d: %w", i, err)
		}

		engineConfigStrs[i] = string(b)
	}

	orderedUniqueIdx := make([]int, 0, len(engineConfigStrs))
	seen := make(map[string]struct{}, len(engineConfigStrs))
	for i, key := range engineConfigStrs {
		if key == "" {
			orderedUniqueIdx = append(orderedUniqueIdx, i)
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		orderedUniqueIdx = append(orderedUniqueIdx, i)
	}

	existed := make(map[string]struct{})
	keys := make([]string, 0, len(seen))
	for k := range seen {
		if k != "" {
			keys = append(keys, k)
		}
	}

	batchSize := 100
	for start := 0; start < len(keys); start += batchSize {
		end := min(start+batchSize, len(keys))

		batch := keys[start:end]
		existing, err := repository.ListExistingEngineConfigs(database.DB, batch)
		if err != nil {
			return nil, err
		}

		for _, v := range existing {
			existed[v] = struct{}{}
		}
	}

	out := make([]injectionProcessItem, 0, len(orderedUniqueIdx))
	for _, idx := range orderedUniqueIdx {
		key := engineConfigStrs[idx]
		if key == "" {
			out = append(out, items[idx])
			continue
		}
		if _, ok := existed[key]; ok {
			continue
		}

		items[idx].executeTime = time.Now().Add(time.Duration(idx*2) * time.Second)
		out = append(out, items[idx])
	}

	return out, nil
}

// sortNodes sorts chaos nodes by their Value field and then by their JSON representation for consistency
func sortNodes(nodes []chaos.Node) []chaos.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	// Create a copy to avoid modifying the original slice
	sortedNodes := make([]chaos.Node, len(nodes))
	copy(sortedNodes, nodes)

	// Sort nodes by their Value field first, then by serialized representation for consistency
	// Using a stable sort to maintain relative order for equal elements
	for i := 0; i < len(sortedNodes)-1; i++ {
		for j := i + 1; j < len(sortedNodes); j++ {
			// Primary sort: by Value field
			if sortedNodes[i].Value > sortedNodes[j].Value {
				sortedNodes[i], sortedNodes[j] = sortedNodes[j], sortedNodes[i]
				continue
			}

			// Secondary sort: if Values are equal, sort by JSON representation for consistency
			if sortedNodes[i].Value == sortedNodes[j].Value {
				iJSON, _ := json.Marshal(sortedNodes[i])
				jJSON, _ := json.Marshal(sortedNodes[j])
				if string(iJSON) > string(jJSON) {
					sortedNodes[i], sortedNodes[j] = sortedNodes[j], sortedNodes[i]
				}
			}
		}
	}

	return sortedNodes
}
