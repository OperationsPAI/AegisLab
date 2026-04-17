package injectionmodule

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	lokiinfra "aegis/infra/loki"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	"aegis/service/common"
	"aegis/utils"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"gorm.io/gorm"
)

type Service struct {
	repo       *Repository
	store      *DatapackStore
	lokiClient *lokiinfra.Client
	redis      *redisinfra.Gateway
}

func NewService(repo *Repository, store *DatapackStore, lokiClient *lokiinfra.Client, redis *redisinfra.Gateway) *Service {
	return &Service{repo: repo, store: store, lokiClient: lokiClient, redis: redis}
}

func (s *Service) ListProjectInjections(ctx context.Context, req *ListInjectionReq, projectID int) (*dto.ListResp[InjectionResp], error) {
	var project model.Project
	if err := s.repo.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	limit, offset := req.ToGormParams()
	injections, total, err := s.repo.ListProjectInjectionsView(projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list injections for project %d: %w", projectID, err)
	}

	items := make([]InjectionResp, 0, len(injections))
	for _, injection := range injections {
		items = append(items, *NewInjectionResp(&injection))
	}

	return &dto.ListResp[InjectionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) Search(ctx context.Context, req *SearchInjectionReq, projectID *int) (*dto.SearchResp[InjectionDetailResp], error) {
	if req == nil {
		return nil, fmt.Errorf("search injection request is nil")
	}
	injections, total, err := s.repo.SearchInjections(req, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to search injections: %w", err)
	}
	items := make([]InjectionDetailResp, 0, len(injections))
	for _, injection := range injections {
		items = append(items, *NewInjectionDetailResp(&injection))
	}

	resp := &dto.SearchResp[InjectionDetailResp]{
		Pagination: req.ConvertToPaginationInfo(total),
	}
	if len(req.GroupBy) > 0 {
		resp.Groups = dto.BuildGroupTree(items, req.GroupBy)
	} else {
		resp.Items = items
	}
	return resp, nil
}

func (s *Service) ListNoIssues(ctx context.Context, req *ListInjectionNoIssuesReq, projectID *int) ([]InjectionNoIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := splitLabelCondition(item)
		labelConditions = append(labelConditions, map[string]string{"key": parts[0], "value": parts[1]})
	}

	opts, err := req.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := s.repo.ListIssuesFreeInjections(labelConditions, &opts.CustomStartTime, &opts.CustomEndTime, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	items := make([]InjectionNoIssuesResp, 0, len(records))
	for i, record := range records {
		resp, err := NewInjectionNoIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp at index %d: %w", i, err)
		}
		items = append(items, *resp)
	}
	return items, nil
}

func (s *Service) ListWithIssues(ctx context.Context, req *ListInjectionWithIssuesReq, projectID *int) ([]InjectionWithIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := splitLabelCondition(item)
		labelConditions = append(labelConditions, map[string]string{"key": parts[0], "value": parts[1]})
	}

	opts, err := req.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := s.repo.ListIssueInjections(labelConditions, &opts.CustomStartTime, &opts.CustomEndTime, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	items := make([]InjectionWithIssuesResp, 0, len(records))
	for _, record := range records {
		resp, err := NewInjectionWithIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp: %w", err)
		}
		items = append(items, *resp)
	}
	return items, nil
}

func (s *Service) SubmitFaultInjection(ctx context.Context, req *SubmitInjectionReq, groupID string, userID int, projectID *int) (*SubmitInjectionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("submit injection request is nil")
	}
	db := s.repo.db

	if projectID == nil {
		project, err := s.repo.ResolveProject(req.ProjectName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
			}
			return nil, fmt.Errorf("failed to get project: %w", err)
		}
		projectID = &project.ID
	}

	pedestalVersionResults, err := common.MapRefsToContainerVersionsWithDB(db, []*dto.ContainerRef{&req.Pedestal.ContainerRef}, consts.ContainerTypePedestal, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map pedestal container ref to version: %w", err)
	}
	pedestalVersion, exists := pedestalVersionResults[&req.Pedestal.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("pedestal version not found for container: %s (version: %s)", req.Pedestal.Name, req.Pedestal.Version)
	}

	helmConfig, err := s.repo.LoadPedestalHelmConfig(pedestalVersion.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: helm config not found for pedestal version id %d", consts.ErrNotFound, pedestalVersion.ID)
		}
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}

	params := flattenYAMLToParameters(req.Pedestal.Payload, "")
	helmValues, err := common.ListHelmConfigValuesWithDB(db, params, helmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to render pedestal helm values: %w", err)
	}

	helmConfigItem := dto.NewHelmConfigItem(helmConfig)
	helmConfigItem.DynamicValues = helmValues

	pedestalItem := dto.NewContainerVersionItem(&pedestalVersion)
	pedestalItem.Extra = helmConfigItem

	benchmarkVersionResults, err := common.MapRefsToContainerVersionsWithDB(db, []*dto.ContainerRef{&req.Benchmark.ContainerRef}, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map benchmark container ref to version: %w", err)
	}
	benchmarkVersion, exists := benchmarkVersionResults[&req.Benchmark.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("benchmark version not found for container: %s (version: %s)", req.Benchmark.Name, req.Benchmark.Version)
	}

	benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
	envVars, err := common.ListContainerVersionEnvVarsWithDB(db, req.Benchmark.EnvVars, &benchmarkVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
	}
	benchmarkVersionItem.EnvVars = envVars

	processedItems := make([]injectionProcessItem, 0, len(req.Specs))
	var parseWarnings []string
	for i := range req.Specs {
		item, warning, err := parseBatchInjectionSpecs(pedestalItem.ContainerName, i, req.Specs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse injection spec batch %d: %w", i, err)
		}
		if warning != "" {
			parseWarnings = append(parseWarnings, warning)
		} else {
			processedItems = append(processedItems, *item)
		}
	}

	uniqueItems, duplicatedInRequest, alreadyExisted, err := s.removeDuplicated(processedItems)
	if err != nil {
		return nil, fmt.Errorf("failed to remove duplicated batches: %w", err)
	}

	var warnings *InjectionWarnings
	if len(parseWarnings) > 0 || len(duplicatedInRequest) > 0 || len(alreadyExisted) > 0 {
		warnings = &InjectionWarnings{
			DuplicateServicesInBatch:  parseWarnings,
			DuplicateBatchesInRequest: duplicatedInRequest,
			BatchesExistInDatabase:    alreadyExisted,
		}
	}

	if len(req.Algorithms) > 0 {
		refs := make([]*dto.ContainerRef, 0, len(req.Algorithms))
		for i := range req.Algorithms {
			refs = append(refs, &req.Algorithms[i].ContainerRef)
		}

		algorithmVersionsResults, err := common.MapRefsToContainerVersionsWithDB(db, refs, consts.ContainerTypeAlgorithm, userID)
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
			envVars, err := common.ListContainerVersionEnvVarsWithDB(db, spec.EnvVars, &algorithmVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to list algorithm env vars: %w", err)
			}

			algorithmVersionItem.EnvVars = envVars
			algorithmVersionItems = append(algorithmVersionItems, algorithmVersionItem)
		}

		if len(algorithmVersionItems) > 0 {
			if err := s.redis.SetHashField(ctx, consts.InjectionAlgorithmsKey, groupID, algorithmVersionItems); err != nil {
				return nil, fmt.Errorf("failed to store injection algorithms: %w", err)
			}
		}
	}

	injectionItems := make([]SubmitInjectionItem, 0, len(uniqueItems))
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
				consts.InjectSystem:      chaos.SystemType(pedestalItem.ContainerName),
			},
		}

		task := &dto.UnifiedTask{
			Type:        consts.TaskTypeRestartPedestal,
			Immediate:   false,
			ExecuteTime: item.executeTime.Unix(),
			Payload:     payload,
			GroupID:     groupID,
			ProjectID:   *projectID,
			UserID:      userID,
			State:       consts.TaskPending,
			Extra: map[consts.TaskExtra]any{
				consts.TaskExtraInjectionAlgorithms: len(req.Algorithms),
			},
		}
		task.SetGroupCtx(ctx)

		if err := common.SubmitTaskWithDB(ctx, db, s.redis, task); err != nil {
			return nil, fmt.Errorf("failed to submit fault injection task: %w", err)
		}

		injectionItems = append(injectionItems, SubmitInjectionItem{
			Index:   item.index,
			TraceID: task.TraceID,
			TaskID:  task.TaskID,
		})
	}

	sort.Slice(injectionItems, func(i, j int) bool { return injectionItems[i].Index < injectionItems[j].Index })
	return &SubmitInjectionResp{
		GroupID:       groupID,
		Items:         injectionItems,
		OriginalCount: len(processedItems),
		Warnings:      warnings,
	}, nil
}

func (s *Service) SubmitDatapackBuilding(ctx context.Context, req *SubmitDatapackBuildingReq, groupID string, userID int, projectID *int) (*SubmitDatapackBuildingResp, error) {
	if req == nil {
		return nil, fmt.Errorf("submit datapack building request is nil")
	}
	db := s.repo.db

	if projectID == nil {
		project, err := s.repo.ResolveProject(req.ProjectName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
			}
			return nil, fmt.Errorf("failed to get project: %w", err)
		}
		projectID = &project.ID
	}

	refs := make([]*dto.ContainerRef, 0, len(req.Specs))
	for i := range req.Specs {
		refs = append(refs, &req.Specs[i].Benchmark.ContainerRef)
	}

	benchmarkVersionResults, err := common.MapRefsToContainerVersionsWithDB(db, refs, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	var allBuildingItems []SubmitBuildingItem
	for idx, spec := range req.Specs {
		datapacks, datasetVersionID, err := common.ExtractDatapacks(s.repo.db, spec.Datapack, spec.Dataset, userID, consts.TaskTypeBuildDatapack)
		if err != nil {
			return nil, fmt.Errorf("failed to extract datapacks: %w", err)
		}

		benchmarkVersion, exists := benchmarkVersionResults[refs[idx]]
		if !exists {
			return nil, fmt.Errorf("benchmark version not found for %v", spec.Benchmark)
		}

		benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
		envVars, err := common.ListContainerVersionEnvVarsWithDB(db, spec.Benchmark.EnvVars, &benchmarkVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
		}
		benchmarkVersionItem.EnvVars = envVars

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
				ProjectID: *projectID,
				UserID:    userID,
				State:     consts.TaskPending,
			}
			task.SetGroupCtx(ctx)

			if err := common.SubmitTaskWithDB(ctx, db, s.redis, task); err != nil {
				return nil, fmt.Errorf("failed to submit datapack building task: %w", err)
			}

			allBuildingItems = append(allBuildingItems, SubmitBuildingItem{
				Index:   idx,
				TraceID: task.TraceID,
				TaskID:  task.TaskID,
			})
		}
	}

	return &SubmitDatapackBuildingResp{
		GroupID: groupID,
		Items:   allBuildingItems,
	}, nil
}

func (s *Service) ListInjections(_ context.Context, req *ListInjectionReq) (*dto.ListResp[InjectionResp], error) {
	limit, offset := req.ToGormParams()
	injections, total, err := s.repo.ListInjectionsView(limit, offset, req.ToFilterOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to list injections: %w", err)
	}

	items := make([]InjectionResp, 0, len(injections))
	for _, injection := range injections {
		items = append(items, *NewInjectionResp(&injection))
	}

	return &dto.ListResp[InjectionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) GetInjection(_ context.Context, id int) (*InjectionDetailResp, error) {
	injection, err := s.repo.GetInjectionWithLabels(id)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}
	return NewInjectionDetailResp(injection), nil
}

func (s *Service) GetMetadata(_ context.Context) (*InjectionMetadataResp, error) {
	return nil, nil
}

func (s *Service) ManageLabels(_ context.Context, req *ManageInjectionLabelReq, id int) (*InjectionResp, error) {
	var managedInjection *model.FaultInjection
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		injection, err := repo.LoadInjection(id)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, id)
			}
			return fmt.Errorf("failed to get injection: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}
			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}
			if err := repo.AddInjectionLabels(injection.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repo.ListInjectionLabelIDsByKeys(injection.ID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}
			if len(labelIDs) > 0 {
				if err := repo.ClearInjectionLabels([]int{id}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear injection labels: %w", err)
				}
				if err := repo.BatchDecreaseLabelUsages(labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		managedInjection, err = repo.GetInjectionWithLabels(id)
		if err != nil {
			return fmt.Errorf("failed to reload injection labels: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewInjectionResp(managedInjection), nil
}

func (s *Service) BatchManageLabels(_ context.Context, req *BatchManageInjectionLabelReq) (*BatchManageInjectionLabelResp, error) {
	resp := &BatchManageInjectionLabelResp{
		FailedItems:  []string{},
		SuccessItems: []InjectionResp{},
	}
	if len(req.Items) == 0 {
		return resp, nil
	}

	return resp, s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		allInjectionIDs := make([]int, 0, len(req.Items))
		operationMap := make(map[int]*InjectionLabelOperation, len(req.Items))
		for i := range req.Items {
			item := &req.Items[i]
			allInjectionIDs = append(allInjectionIDs, item.InjectionID)
			operationMap[item.InjectionID] = item
		}

		foundIDMap, err := repo.LoadExistingInjectionsByID(allInjectionIDs)
		if err != nil {
			return fmt.Errorf("failed to list injections: %w", err)
		}

		validIDs := make([]int, 0, len(foundIDMap))
		for _, id := range allInjectionIDs {
			if _, found := foundIDMap[id]; !found {
				resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d not found", id))
				resp.FailedCount++
				delete(operationMap, id)
			} else {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) == 0 {
			return fmt.Errorf("no valid injection IDs found")
		}

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

		var labelMap map[string]int
		if len(allAddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, allAddLabels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}
			labelMap = make(map[string]int, len(labels))
			for _, label := range labels {
				labelMap[label.Key+":"+label.Value] = label.ID
			}
		}

		var removeLabelMap map[string]int
		if len(allRemoveLabels) > 0 {
			labelConditions := make([]map[string]string, 0, len(allRemoveLabels))
			for _, item := range allRemoveLabels {
				labelConditions = append(labelConditions, map[string]string{"key": item.Key, "value": item.Value})
			}
			removeLabelMap, err = repo.LoadInjectionLabelIDsByItems(labelConditions, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to find labels to remove: %w", err)
			}
		}

		for _, injectionID := range validIDs {
			op := operationMap[injectionID]
			if len(op.AddLabels) > 0 {
				labelIDsToAdd := make([]int, 0, len(op.AddLabels))
				for _, label := range op.AddLabels {
					if labelID, exists := labelMap[label.Key+":"+label.Value]; exists {
						labelIDsToAdd = append(labelIDsToAdd, labelID)
					}
				}
				if len(labelIDsToAdd) > 0 {
					if err := repo.AddInjectionLabels(injectionID, labelIDsToAdd); err != nil {
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
					if labelID, exists := removeLabelMap[label.Key+":"+label.Value]; exists {
						labelIDsToRemove = append(labelIDsToRemove, labelID)
					}
				}
				if len(labelIDsToRemove) > 0 {
					if err := repo.ClearInjectionLabels([]int{injectionID}, labelIDsToRemove); err != nil {
						resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d: failed to remove labels - %s", injectionID, err.Error()))
						resp.FailedCount++
						delete(foundIDMap, injectionID)
						continue
					}
				}
			}
		}

		if len(foundIDMap) > 0 {
			successIDs := make([]int, 0, len(foundIDMap))
			for id := range foundIDMap {
				successIDs = append(successIDs, id)
			}
			updatedInjections, err := repo.ListFaultInjectionsByIDWithLabels(successIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch updated injections: %w", err)
			}
			for i := range updatedInjections {
				injection := &updatedInjections[i]
				resp.SuccessItems = append(resp.SuccessItems, *NewInjectionResp(injection))
				resp.SuccessCount++
			}
		}

		return nil
	})
}

func (s *Service) BatchDelete(ctx context.Context, req *BatchDeleteInjectionReq) error {
	if len(req.IDs) > 0 {
		return s.batchDeleteByIDs(req.IDs)
	}
	return s.batchDeleteByLabels(req.Labels)
}

func (s *Service) Clone(_ context.Context, id int, req *CloneInjectionReq) (*InjectionDetailResp, error) {
	original, err := s.repo.LoadInjection(id)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	cloned := &model.FaultInjection{
		Name:          req.Name,
		FaultType:     original.FaultType,
		Category:      original.Category,
		Description:   original.Description,
		DisplayConfig: original.DisplayConfig,
		EngineConfig:  original.EngineConfig,
		Groundtruths:  original.Groundtruths,
		PreDuration:   original.PreDuration,
		StartTime:     original.StartTime,
		EndTime:       original.EndTime,
		BenchmarkID:   original.BenchmarkID,
		PedestalID:    original.PedestalID,
		State:         consts.DatapackInitial,
		Status:        consts.CommonEnabled,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		if err := repo.CreateInjectionRecord(cloned); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: injection with name %s already exists", consts.ErrAlreadyExists, cloned.Name)
			}
			return fmt.Errorf("failed to create injection: %w", err)
		}
		if len(req.Labels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.Labels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}
			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}
			if err := repo.AddInjectionLabels(cloned.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cloned, err = s.repo.GetInjectionWithLabels(cloned.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloned injection labels: %w", err)
	}
	return NewInjectionDetailResp(cloned), nil
}

func (s *Service) GetLogs(ctx context.Context, id int) (*InjectionLogsResp, error) {
	injection, err := s.repo.LoadInjection(id)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	resp := &InjectionLogsResp{InjectionID: id, Logs: []string{}}
	if injection.TaskID == nil {
		return resp, nil
	}

	resp.TaskID = *injection.TaskID
	task, taskErr := s.repo.LoadTask(*injection.TaskID)
	if taskErr != nil {
		return resp, nil
	}

	lokiCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	logEntries, lokiErr := s.lokiClient.QueryJobLogs(lokiCtx, *injection.TaskID, lokiinfra.QueryOpts{
		Start:     task.CreatedAt,
		Direction: "forward",
	})
	if lokiErr != nil {
		return resp, nil
	}
	for _, entry := range logEntries {
		resp.Logs = append(resp.Logs, entry.Line)
	}
	return resp, nil
}

func (s *Service) GetDatapackFilename(_ context.Context, id int) (string, error) {
	injection, err := s.repo.LoadInjection(id)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, id)
		}
		return "", fmt.Errorf("failed to get injection: %w", err)
	}
	if injection.State < consts.DatapackBuildSuccess {
		return "", fmt.Errorf("datapack for injection id %d is not ready for download", id)
	}
	return injection.Name, nil
}

func (s *Service) DownloadDatapack(_ context.Context, zipWriter *zip.Writer, excludeRules []utils.ExculdeRule, id int) error {
	if zipWriter == nil {
		return fmt.Errorf("zip writer cannot be nil")
	}
	injection, err := s.getReadyDatapack(id)
	if err != nil {
		return err
	}
	if err := s.store.Package(zipWriter, injection.Name, excludeRules); err != nil {
		return fmt.Errorf("failed to package injection to zip: %w", err)
	}
	return nil
}

func (s *Service) GetDatapackFiles(_ context.Context, id int, baseURL string) (*DatapackFilesResp, error) {
	injection, err := s.getReadyDatapack(id)
	if err != nil {
		return nil, err
	}
	resp, err := s.store.BuildFileTree(injection.Name, baseURL, id)
	if err != nil {
		return nil, fmt.Errorf("failed to build file tree: %w", err)
	}
	return resp, nil
}

func (s *Service) DownloadDatapackFile(_ context.Context, id int, filePath string) (string, string, int64, io.ReadSeekCloser, error) {
	injection, err := s.getReadyDatapack(id)
	if err != nil {
		return "", "", 0, nil, err
	}
	return s.store.OpenFile(injection.Name, filePath)
}

func (s *Service) QueryDatapackFile(ctx context.Context, id int, filePath string) (string, int64, io.ReadCloser, error) {
	return s.queryDatapackFileContent(ctx, id, filePath)
}

func (s *Service) UpdateGroundtruth(_ context.Context, id int, req *UpdateGroundtruthReq) error {
	if _, err := s.repo.LoadInjection(id); err != nil {
		return err
	}
	return s.repo.UpdateGroundtruth(id, req.Groundtruths, consts.GroundtruthSourceManual)
}

func (s *Service) UploadDatapack(_ context.Context, req *UploadDatapackReq, file io.Reader, fileSize int64) (*UploadDatapackResp, error) {
	_ = fileSize

	labels, err := req.ParseLabels()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", consts.ErrBadRequest, err.Error())
	}

	groundtruths, err := req.ParseGroundtruths()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", consts.ErrBadRequest, err.Error())
	}

	existing, _ := s.repo.FindInjectionByName(req.Name, false)
	if existing != nil {
		return nil, fmt.Errorf("%w: injection with name %s already exists", consts.ErrAlreadyExists, req.Name)
	}

	tmpFile, err := s.store.CreateUploadTempFile()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = s.store.Remove(tmpPath) }()

	if _, err := io.Copy(tmpFile, file); err != nil {
		_ = tmpFile.Close()
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close uploaded file: %w", err)
	}

	if err := s.store.ValidateArchive(tmpPath); err != nil {
		return nil, fmt.Errorf("%w: %s", consts.ErrBadRequest, err.Error())
	}

	targetDir, err := s.store.EnsureDatapackDirAvailable(req.Name)
	if err != nil {
		return nil, err
	}
	if err := s.store.ExtractArchive(tmpPath, targetDir); err != nil {
		_ = s.store.RemoveAll(targetDir)
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	groundtruthSource := ""
	if len(groundtruths) > 0 {
		groundtruthSource = consts.GroundtruthSourceManual
	} else {
		groundtruths = s.store.ExtractGroundtruths(targetDir)
		if len(groundtruths) > 0 {
			groundtruthSource = consts.GroundtruthSourceImported
		}
	}

	category := chaos.SystemType("")
	if req.Category != "" {
		category = chaos.SystemType(req.Category)
	}

	injection := &model.FaultInjection{
		Name:              req.Name,
		Source:            consts.DatapackSourceManual,
		FaultType:         chaos.ChaosType(0),
		Category:          category,
		Description:       req.Description,
		EngineConfig:      "",
		Groundtruths:      groundtruths,
		GroundtruthSource: groundtruthSource,
		PreDuration:       0,
		BenchmarkID:       nil,
		PedestalID:        nil,
		State:             consts.DatapackBuildSuccess,
		Status:            consts.CommonEnabled,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		if err := repo.CreateInjectionRecord(injection); err != nil {
			return err
		}

		if len(labels) > 0 {
			createdLabels, err := common.CreateOrUpdateLabelsFromItems(tx, labels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			labelIDs := make([]int, 0, len(createdLabels))
			for _, label := range createdLabels {
				labelIDs = append(labelIDs, label.ID)
			}

			if err := repo.AddInjectionLabels(injection.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		_ = s.store.RemoveAll(targetDir)
		return nil, err
	}

	return &UploadDatapackResp{
		ID:   injection.ID,
		Name: injection.Name,
	}, nil
}

func (s *Service) getReadyDatapack(id int) (*model.FaultInjection, error) {
	injection, err := s.repo.LoadInjection(id)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}
	if injection.State < consts.DatapackBuildSuccess {
		return nil, fmt.Errorf("datapack %d is not ready", id)
	}
	return injection, nil
}

func (s *Service) batchDeleteByIDs(injectionIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}
	return s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		return repo.DeleteInjectionsCascade(injectionIDs)
	})
}

func (s *Service) batchDeleteByLabels(labelItems []dto.LabelItem) error {
	if len(labelItems) == 0 {
		return nil
	}
	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, item := range labelItems {
		labelConditions = append(labelConditions, map[string]string{"key": item.Key, "value": item.Value})
	}
	injectionIDs, err := s.repo.ListInjectionIDsByLabelConditions(labelConditions)
	if err != nil {
		return fmt.Errorf("failed to list injection ids by labels: %w", err)
	}
	return s.batchDeleteByIDs(injectionIDs)
}

func splitLabelCondition(item string) [2]string {
	parts := strings.SplitN(item, ":", 2)
	if len(parts) == 1 {
		return [2]string{parts[0], ""}
	}
	return [2]string{parts[0], parts[1]}
}
