package executionmodule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	"aegis/service/common"
	"aegis/utils"

	"gorm.io/gorm"
)

type Service struct {
	repo  *Repository
	redis *redisinfra.Gateway
}

func NewService(repo *Repository, redis *redisinfra.Gateway) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) ListProjectExecutions(_ context.Context, req *ListExecutionReq, projectID int) (*dto.ListResp[ExecutionResp], error) {
	var project model.Project
	if err := s.repo.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	limit, offset := req.ToGormParams()
	executions, total, err := s.repo.listProjectExecutionsView(projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions for project %d: %w", projectID, err)
	}

	items := make([]ExecutionResp, 0, len(executions))
	for i := range executions {
		items = append(items, *NewExecutionResp(&executions[i], executions[i].Labels))
	}

	return &dto.ListResp[ExecutionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) SubmitAlgorithmExecution(ctx context.Context, req *SubmitExecutionReq, groupID string, userID int) (*SubmitExecutionResp, error) {
	db := s.repo.db

	project, err := s.repo.getProjectByName(req.ProjectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	refs := make([]*dto.ContainerRef, 0, len(req.Specs))
	for i := range req.Specs {
		refs = append(refs, &req.Specs[i].Algorithm.ContainerRef)
	}

	algorithmVersionResults, err := common.MapRefsToContainerVersionsWithDB(db, refs, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}
	if len(algorithmVersionResults) == 0 {
		return nil, fmt.Errorf("no valid algorithm versions found for the provided specs")
	}

	var allExecutionItems []SubmitExecutionItem
	for idx, spec := range req.Specs {
		datapacks, datasetID, err := common.ExtractDatapacks(s.repo.db, spec.Datapack, spec.Dataset, userID, consts.TaskTypeRunAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("failed to extract datapacks: %w", err)
		}

		algorithmVersion, exists := algorithmVersionResults[refs[idx]]
		if !exists {
			return nil, fmt.Errorf("algorithm version not found for %v", spec.Algorithm)
		}

		for _, datapack := range datapacks {
			if datapack.StartTime == nil || datapack.EndTime == nil {
				return nil, fmt.Errorf("datapack %s does not have valid start_time and end_time", datapack.Name)
			}

			algorithmItem := dto.NewContainerVersionItem(&algorithmVersion)
			envVars, err := common.ListContainerVersionEnvVarsWithDB(db, spec.Algorithm.EnvVars, &algorithmVersion)
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

			if err := common.SubmitTaskWithDB(ctx, db, s.redis, task); err != nil {
				return nil, fmt.Errorf("failed to submit task: %w", err)
			}

			allExecutionItems = append(allExecutionItems, SubmitExecutionItem{
				Index:              idx,
				TraceID:            task.TraceID,
				TaskID:             task.TaskID,
				AlgorithmID:        algorithmVersion.ContainerID,
				AlgorithmVersionID: algorithmVersion.ID,
				DatapackID:         &datapack.ID,
			})
		}
	}

	return &SubmitExecutionResp{
		GroupID: groupID,
		Items:   allExecutionItems,
	}, nil
}

func (s *Service) ListExecutions(_ context.Context, req *ListExecutionReq) (*dto.ListResp[ExecutionResp], error) {
	limit, offset := req.ToGormParams()
	executions, total, err := s.repo.listExecutionsView(limit, offset, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}

	items := make([]ExecutionResp, 0, len(executions))
	for i := range executions {
		items = append(items, *NewExecutionResp(&executions[i], executions[i].Labels))
	}

	return &dto.ListResp[ExecutionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) GetExecution(_ context.Context, id int) (*ExecutionDetailResp, error) {
	execution, labels, detectorResults, granularityResults, err := s.repo.getExecutionResultView(id)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: execution id: %d", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	resp := NewExecutionDetailResp(execution, labels)
	if len(detectorResults) > 0 {
		items := make([]DetectorResultItem, 0, len(detectorResults))
		for _, result := range detectorResults {
			items = append(items, NewDetectorResultItem(&result))
		}
		resp.DetectorResults = items
	}
	if len(granularityResults) > 0 {
		items := make([]GranularityResultItem, 0, len(granularityResults))
		for _, result := range granularityResults {
			items = append(items, NewGranularityResultItem(&result))
		}
		resp.GranularityResults = items
	}
	return resp, nil
}

func (s *Service) ListAvailableLabels(_ context.Context) ([]dto.LabelItem, error) {
	labels, err := s.repo.listAvailableExecutionLabels()
	if err != nil {
		return nil, err
	}

	items := make([]dto.LabelItem, 0, len(labels))
	for _, label := range labels {
		items = append(items, dto.LabelItem{Key: label.Key, Value: label.Value})
	}
	return items, nil
}

func (s *Service) ManageLabels(_ context.Context, req *ManageExecutionLabelReq, executionID int) (*ExecutionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage execution labels request is nil")
	}

	var managedExecution *model.Execution
	var managedLabels []model.Label
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		execution, _, err := repo.getExecutionView(executionID)
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
			if err := repo.AddExecutionLabels(execution.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add execution labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repo.listExecutionLabelIDsByKeys(execution.ID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) > 0 {
				if err := repo.ClearExecutionLabels([]int{executionID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear execution labels: %w", err)
				}
				if err := repo.BatchDecreaseLabelUsages(labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		reloadedExecution, labels, err := repo.getExecutionView(executionID)
		if err != nil {
			return fmt.Errorf("failed to reload execution labels: %w", err)
		}
		managedExecution = reloadedExecution
		managedLabels = labels
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewExecutionResp(managedExecution, managedLabels), nil
}

func (s *Service) BatchDelete(_ context.Context, req *BatchDeleteExecutionReq) error {
	if len(req.IDs) > 0 {
		return s.batchDeleteByIDs(req.IDs)
	}
	return s.batchDeleteByLabels(req.Labels)
}

func (s *Service) UploadDetectorResults(_ context.Context, req *UploadDetectorResultReq, executionID int) (*UploadExecutionResultResp, error) {
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		if err := repo.UpdateExecutionDuration(executionID, req.Duration); err != nil {
			return err
		}

		results := make([]model.DetectorResult, 0, len(req.Results))
		for _, item := range req.Results {
			results = append(results, *item.ConvertToDetectorResult(executionID))
		}
		if err := repo.SaveDetectorResults(results); err != nil {
			return fmt.Errorf("failed to save detector results for execution %d: %w", executionID, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UploadExecutionResultResp{
		ResultCount:  len(req.Results),
		UploadedAt:   time.Now(),
		HasAnomalies: req.HasAnomalies(),
	}, nil
}

func (s *Service) UploadGranularityResults(_ context.Context, req *UploadGranularityResultReq, executionID int) (*UploadExecutionResultResp, error) {
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.withDB(tx)
		if err := repo.UpdateExecutionDuration(executionID, req.Duration); err != nil {
			return err
		}

		results := make([]model.GranularityResult, 0, len(req.Results))
		for _, item := range req.Results {
			results = append(results, *item.ConvertToGranularityResult(executionID))
		}
		if err := repo.SaveGranularityResults(results); err != nil {
			return fmt.Errorf("failed to save detector results for execution %d: %w", executionID, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UploadExecutionResultResp{
		ResultCount: len(req.Results),
		UploadedAt:  time.Now(),
	}, nil
}

func (s *Service) batchDeleteByIDs(executionIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}
	return s.repo.Transaction(func(tx *gorm.DB) error {
		return s.repo.withDB(tx).BatchDeleteExecutions(executionIDs)
	})
}

func (s *Service) batchDeleteByLabels(labelItems []dto.LabelItem) error {
	if len(labelItems) == 0 {
		return nil
	}
	executionIDs, err := s.repo.ListExecutionIDsByLabelItems(labelItems)
	if err != nil {
		return fmt.Errorf("failed to list execution ids by labels: %w", err)
	}
	return s.batchDeleteByIDs(executionIDs)
}
