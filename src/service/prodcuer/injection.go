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
	"gorm.io/gorm"
)

type injectionProcessItem struct {
	index         int
	faultDuration int
	node          *chaos.Node
	executeTime   time.Time
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

	labelCondtions := make([]map[string]string, 0, len(labelItems))
	for _, item := range labelItems {
		labelCondtions = append(labelCondtions, map[string]string{
			"key":   item.Key,
			"value": item.Value,
		})
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		injectionIDs, err := repository.ListInjectionIDsByLabels(database.DB, labelCondtions)
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
	injection, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	labels, err := repository.ListLabelsByInjectionID(database.DB, injection.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get injection labels: %w", err)
	}

	injection.Task.Labels = labels
	resp := dto.NewInjectionDetailResp(injection)

	groundTruths, err := getInjectionGroundTruths([]string{injection.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to get injection ground truths: %w", err)
	}

	if gt, exists := groundTruths[injection.Name]; exists {
		resp.GroundTruth = &gt
	}

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
			injection.Task.Labels = labels
		}
		injectionResps = append(injectionResps, *dto.NewInjectionResp(&injection))
	}

	resp := dto.ListResp[dto.InjectionResp]{
		Items:      injectionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// ListInjectionsNoissues handles the request to list fault injections without issues
func ListInjectionsNoIssues(req *dto.ListInjectionNoIssuesReq) ([]dto.InjectionNoIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelCondtions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelCondtions = append(labelCondtions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := repository.ListInjectionsNoIssues(database.DB, labelCondtions, &opts.CustomStartTime, &opts.CustomEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	var items []dto.InjectionNoIssuesResp
	for _, record := range records {
		resp, err := dto.NewInjectionNoIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp: %w", err)
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

	labelCondtions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelCondtions = append(labelCondtions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := repository.ListInjectionsWithIssues(database.DB, labelCondtions, &opts.CustomStartTime, &opts.CustomEndTime)
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
			if err := repository.AddInjectionLabels(tx, injectionID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndInjectionID(tx, injection.ID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) == 0 {
				return fmt.Errorf("no labels found for the given keys")
			}

			if err := repository.ClearInjectionLabels(tx, []int{injectionID}, labelIDs); err != nil {
				return fmt.Errorf("failed to clear injection labels: %w", err)
			}

			if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
				return fmt.Errorf("failed to decrease label usage counts: %w", err)
			}
		}

		labels, err := repository.ListLabelsByInjectionID(database.DB, injectionID)
		if err != nil {
			return fmt.Errorf("failed to get injection labels: %w", err)
		}

		injection.Task.Labels = labels
		managedInjection = injection
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewInjectionResp(managedInjection), nil
}

// ProduceRestartPedestalTasks produces pedestal restart tasks into Redis based on the request specifications
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
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	pedestalVersion, exists := pedestalVersionResults[&req.Pedestal.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("pedestal version not found for %v", req.Pedestal)
	}

	helmConfig, err := repository.GetHelmConfigByContainerVersionID(database.DB, pedestalVersion.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: helm config not found for pedestal version id %d", consts.ErrNotFound, pedestalVersion.ID)
		}
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}

	pedestalItem := dto.NewContainerVersionItem(&pedestalVersion)
	envVars, err := common.ListContainerVersionEnvVars(req.Pedestal.EnvVars, &pedestalVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to list pedestal env vars: %w", err)
	}

	pedestalItem.EnvVars = envVars

	helmValues, err := common.ListHelmConfigValues(nil, helmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to list pedestal helm values: %w", err)
	}

	pedestalInfo := dto.NewPedestalInfo(&pedestalVersion, helmConfig)
	pedestalInfo.HelmConfig.Values = helmValues
	pedestalItem.Extra = pedestalInfo

	benchmarkVersionResults, err := common.MapRefsToContainerVersions([]*dto.ContainerRef{&req.Benchmark.ContainerRef}, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	benchmarkVersion, exists := benchmarkVersionResults[&req.Benchmark.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("benchmark version not found for %v", req.Benchmark)
	}

	benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
	envVars, err = common.ListContainerVersionEnvVars(req.Benchmark.EnvVars, &benchmarkVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
	}

	benchmarkVersionItem.EnvVars = envVars

	processedItems, err := parseInjectionSpecs(req.Specs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse injection specs: %w", err)
	}

	injectionItems := make([]dto.SubmitInjectionItem, 0, len(processedItems))
	for _, item := range processedItems {
		payload := map[string]any{
			consts.RestartPedestal:      pedestalItem,
			consts.RestartHelmConfig:    helmConfig,
			consts.RestartIntarval:      req.Interval,
			consts.RestartFaultDuration: item.faultDuration,
			consts.RestartInjectPayload: map[string]any{
				consts.InjectBenchmark:   benchmarkVersionItem,
				consts.InjectPreDuration: req.PreDuration,
				consts.InjectNode:        item.node,
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

	return &dto.SubmitInjectionResp{
		GroupID:         groupID,
		Items:           injectionItems,
		DuplicatedCount: len(req.Specs) - len(processedItems),
		OriginalCount:   len(req.Specs),
	}, nil
}

// ProduceFaultInjectionTasks produces fault injection tasks into Redis based on the request specifications
func ProduceFaultInjectionTasks(ctx context.Context, task *dto.UnifiedTask, injectTime time.Time, payload map[string]any) error {
	newTask := &dto.UnifiedTask{
		Type:         consts.TaskTypeFaultInjection,
		Immediate:    false,
		ExecuteTime:  injectTime.Unix(),
		Payload:      payload,
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		State:        consts.TaskPending,
		TraceCarrier: task.TraceCarrier,
		GroupCarrier: task.GroupCarrier,
	}
	err := common.SubmitTask(ctx, newTask)
	if err != nil {
		return fmt.Errorf("failed to submit fault injection task: %w", err)
	}
	return nil
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
		datapacks, datasetVersionID, err := extractDatapacks(database.DB, spec.Datapack, spec.Dataset, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to extract datapacks: %w", err)
		}

		var buildingItems []dto.SubmitBuildingItem
		for _, datapack := range datapacks {
			if datapack.StartTime == nil || datapack.EndTime == nil {
				return nil, fmt.Errorf("datapack %s does not have valid start_time and end_time", datapack.Name)
			}

			benchmarkVersion, exists := benchmarkVersionResults[&spec.Benchmark.ContainerRef]
			if !exists {
				return nil, fmt.Errorf("benchmark version not found for %v", spec.Benchmark)
			}

			benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
			envVars, err := common.ListContainerVersionEnvVars(spec.Benchmark.EnvVars, &benchmarkVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
			}

			benchmarkVersionItem.EnvVars = envVars

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

func getInjectionGroundTruths(names []string) (map[string]chaos.Groundtruth, error) {
	engineConfigMap, err := repository.ListEngineConfigByNames(database.DB, names)
	if err != nil {
		return nil, err
	}

	groundtruthMap := make(map[string]chaos.Groundtruth, len(engineConfigMap))
	for injection, engineConf := range engineConfigMap {
		var node chaos.Node
		if err := json.Unmarshal([]byte(engineConf), &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chaos-experiment node for injection %s: %w", injection, err)
		}

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert chaos-experiment node to InjectionConf for injection %s: %w", injection, err)
		}

		groundtruth, err := conf.GetGroundtruth()
		if err != nil {
			return nil, fmt.Errorf("failed to get ground truth for injection %s: %w", injection, err)
		}

		groundtruthMap[injection] = groundtruth
	}

	return groundtruthMap, nil
}

func parseInjectionSpecs(specs []chaos.Node) ([]injectionProcessItem, error) {
	items := make([]injectionProcessItem, 0, len(specs))
	for idx, spec := range specs {
		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, fmt.Errorf("failed to find key %d in the children", spec.Value)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value

		items = append(items, injectionProcessItem{
			index:         idx,
			faultDuration: faultDuration,
			node:          &spec,
		})
	}

	if len(items) != 0 {
		newItems, err := removeDuplicated(items)
		if err != nil {
			return nil, fmt.Errorf("failed to remove duplicated injection specs: %w", err)
		}
		return newItems, nil
	}

	return items, nil
}

// Filter out items that already exist in DB, using engine_config as uniqueness key,
// and drop duplicates within the incoming request while preserving order.
func removeDuplicated(items []injectionProcessItem) ([]injectionProcessItem, error) {
	engineConfigStrs := make([]string, len(items))
	for i, item := range items {
		if item.node == nil {
			engineConfigStrs[i] = ""
			continue
		}

		b, err := json.Marshal(item.node)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal engine config at index %d: %w", i, err)
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
