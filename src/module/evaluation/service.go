package evaluationmodule

import (
	"context"
	"encoding/json"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	executionmodule "aegis/module/execution"
	"aegis/repository"
	"aegis/service/common"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListDatapackEvaluationResults(_ context.Context, req *BatchEvaluateDatapackReq, userID int) (*BatchEvaluateDatapackResp, error) {
	if req == nil {
		return nil, fmt.Errorf("batch evaluate datapack request is nil")
	}

	algorithms := make([]*dto.ContainerRef, 0, len(req.Specs))
	for i := range req.Specs {
		algorithms = append(algorithms, &req.Specs[i].Algorithm)
	}

	algorithmVersionResults, err := common.MapRefsToContainerVersionsWithDB(s.repo.db, algorithms, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	successItems := make([]EvaluateDatapackItem, 0, len(req.Specs))
	failedItems := make([]string, 0)

	for i := range req.Specs {
		spec := &req.Specs[i]
		specIdentifier := fmt.Sprintf("spec[%d]: algorithm=%s, datapack=%s", i, spec.Algorithm.Name, spec.Datapack)

		algorithmVersion, exists := algorithmVersionResults[algorithms[i]]
		if !exists {
			failedItems = append(failedItems, fmt.Sprintf("%s - algorithm version not found", specIdentifier))
			continue
		}

		labelConditions := dto.ConvertLabelItemsToConditions(spec.FilterLabels)
		executions, err := repository.ListExecutionsByDatapackFilter(s.repo.db, algorithmVersion.ID, spec.Datapack, labelConditions)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("%s - failed to query executions: %v", specIdentifier, err))
			continue
		}
		if len(executions) == 0 {
			failedItems = append(failedItems, fmt.Sprintf("%s - no executions found", specIdentifier))
			continue
		}

		refs := make([]executionmodule.ExecutionRef, 0, len(executions))
		for _, execution := range executions {
			refs = append(refs, executionmodule.NewExecutionGranularityRef(&execution))
		}

		evaluateRef := EvaluateDatapackRef{
			Datapack:      spec.Datapack,
			ExecutionRefs: refs,
		}

		datapack := executions[0].Datapack
		if datapack != nil {
			groundtruths, err := getGroundtruths(datapack)
			if err != nil {
				logrus.Warnf("failed to get groundtruth for datapack %s: %v", spec.Datapack, err)
			} else {
				evaluateRef.Groundtruths = groundtruths
			}
		}

		successItems = append(successItems, EvaluateDatapackItem{
			Algorithm:           algorithmVersion.Container.Name,
			AlgorithmVersion:    algorithmVersion.Name,
			EvaluateDatapackRef: evaluateRef,
		})
	}

	persistEvaluations(s.repo.db, "datapack", successItems, func(item *EvaluateDatapackItem) *model.Evaluation {
		return &model.Evaluation{
			AlgorithmName:    item.Algorithm,
			AlgorithmVersion: item.AlgorithmVersion,
			DatapackName:     item.Datapack,
			EvalType:         consts.EvalTypeDatapack,
			Status:           consts.CommonEnabled,
		}
	})

	return &BatchEvaluateDatapackResp{
		SuccessCount: len(successItems),
		SuccessItems: successItems,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}

func (s *Service) ListDatasetEvaluationResults(_ context.Context, req *BatchEvaluateDatasetReq, userID int) (*BatchEvaluateDatasetResp, error) {
	if req == nil {
		return nil, fmt.Errorf("batch evaluate datapack request is nil")
	}

	algorithms := make([]*dto.ContainerRef, 0, len(req.Specs))
	datasets := make([]*dto.DatasetRef, 0, len(req.Specs))
	for i := range req.Specs {
		algorithms = append(algorithms, &req.Specs[i].Algorithm)
		datasets = append(datasets, &req.Specs[i].Dataset)
	}

	algorithmVersionResults, err := common.MapRefsToContainerVersionsWithDB(s.repo.db, algorithms, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	datasetVersionResults, err := common.MapRefsToDatasetVersionsWithDB(s.repo.db, datasets, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map dataset refs to versions: %w", err)
	}

	successItems := make([]EvaluateDatasetItem, 0, len(req.Specs))
	failedItems := make([]string, 0)

	for i := range req.Specs {
		spec := &req.Specs[i]
		specIdentifier := fmt.Sprintf("spec[%d]: algorithm=%s, dataset=%s", i, spec.Algorithm.Name, spec.Dataset.Name)

		algorithmVersion, exists := algorithmVersionResults[algorithms[i]]
		if !exists {
			failedItems = append(failedItems, fmt.Sprintf("%s - algorithm version not found", specIdentifier))
			continue
		}

		datasetVersion, exists := datasetVersionResults[datasets[i]]
		if !exists {
			failedItems = append(failedItems, fmt.Sprintf("%s - dataset version not found", specIdentifier))
			continue
		}

		labelConditions := dto.ConvertLabelItemsToConditions(spec.FilterLabels)
		executions, err := repository.ListExecutionsByDatasetFilter(s.repo.db, algorithmVersion.ID, datasetVersion.ID, labelConditions)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("%s - failed to query executions: %v", specIdentifier, err))
			continue
		}
		if len(executions) == 0 {
			failedItems = append(failedItems, fmt.Sprintf("%s - no executions found", specIdentifier))
			continue
		}

		executionMap := make(map[string][]model.Execution)
		for _, execution := range executions {
			name := execution.Datapack.Name
			if _, exists := executionMap[name]; !exists {
				executionMap[name] = make([]model.Execution, 0)
			}
			executionMap[name] = append(executionMap[name], execution)
		}

		notExecutedDatapacks := make([]string, 0)
		for _, datapack := range datasetVersion.Datapacks {
			if _, exists := executionMap[datapack.Name]; !exists {
				notExecutedDatapacks = append(notExecutedDatapacks, datapack.Name)
			}
		}

		evaluateRefs := make([]EvaluateDatapackRef, 0, len(executionMap))
		for datapackName, groupedExecutions := range executionMap {
			refs := make([]executionmodule.ExecutionRef, 0, len(groupedExecutions))
			for _, execution := range groupedExecutions {
				refs = append(refs, executionmodule.NewExecutionGranularityRef(&execution))
			}

			evaluateRef := EvaluateDatapackRef{
				Datapack:      datapackName,
				ExecutionRefs: refs,
			}

			datapack := groupedExecutions[0].Datapack
			if datapack != nil {
				groundtruths, err := getGroundtruths(datapack)
				if err != nil {
					logrus.Warnf("failed to get groundtruth for datapack %s: %v", datapackName, err)
				} else {
					evaluateRef.Groundtruths = groundtruths
				}
			}

			evaluateRefs = append(evaluateRefs, evaluateRef)
		}

		successItems = append(successItems, EvaluateDatasetItem{
			Algorithm:            algorithmVersion.Container.Name,
			AlgorithmVersion:     algorithmVersion.Name,
			Dataset:              datasetVersion.Dataset.Name,
			DatasetVersion:       datasetVersion.Name,
			TotalCount:           len(datasetVersion.Datapacks),
			EvaluateRefs:         evaluateRefs,
			NotExecutedDatapacks: notExecutedDatapacks,
		})
	}

	persistEvaluations(s.repo.db, "dataset", successItems, func(item *EvaluateDatasetItem) *model.Evaluation {
		return &model.Evaluation{
			AlgorithmName:    item.Algorithm,
			AlgorithmVersion: item.AlgorithmVersion,
			DatasetName:      item.Dataset,
			DatasetVersion:   item.DatasetVersion,
			EvalType:         consts.EvalTypeDataset,
			Status:           consts.CommonEnabled,
		}
	})

	return &BatchEvaluateDatasetResp{
		SuccessCount: len(successItems),
		SuccessItems: successItems,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}

func (s *Service) ListEvaluations(_ context.Context, req *ListEvaluationReq) (*dto.ListResp[EvaluationResp], error) {
	limit, offset := req.ToGormParams()
	evaluations, total, err := s.repo.ListEvaluations(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}

	items := make([]EvaluationResp, 0, len(evaluations))
	for _, evaluation := range evaluations {
		items = append(items, *NewEvaluationResp(&evaluation))
	}

	return &dto.ListResp[EvaluationResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) GetEvaluation(_ context.Context, id int) (*EvaluationResp, error) {
	evaluation, err := s.repo.GetEvaluationByID(id)
	if err != nil {
		return nil, err
	}
	return NewEvaluationResp(evaluation), nil
}

func (s *Service) DeleteEvaluation(_ context.Context, id int) error {
	return s.repo.DeleteEvaluation(id)
}

func persistEvaluations[T any](db *gorm.DB, evalType string, items []T, toEval func(*T) *model.Evaluation) {
	if len(items) == 0 {
		return
	}

	evals := make([]model.Evaluation, 0, len(items))
	for i := range items {
		eval := toEval(&items[i])
		resultJSON, err := json.Marshal(&items[i])
		if err != nil {
			logrus.Warnf("failed to marshal %s evaluation result: %v", evalType, err)
			eval.ResultJSON = "{}"
		} else {
			eval.ResultJSON = string(resultJSON)
		}
		evals = append(evals, *eval)
	}

	if err := db.Create(&evals).Error; err != nil {
		logrus.Warnf("failed to batch persist %d %s evaluations: %v", len(evals), evalType, err)
	}
}

func getGroundtruths(datapack *model.FaultInjection) ([]chaos.Groundtruth, error) {
	chaosGroundtruths := make([]chaos.Groundtruth, 0, len(datapack.Groundtruths))
	for _, gt := range datapack.Groundtruths {
		chaosGroundtruths = append(chaosGroundtruths, *gt.ConvertToChaosGroundtruth())
	}
	return chaosGroundtruths, nil
}
