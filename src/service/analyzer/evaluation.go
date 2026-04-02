package analyzer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"encoding/json"
	"fmt"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
)

// ListDatapackEvaluationResults retrieves evaluation data for multiple algorithm-datapack pairs
func ListDatapackEvaluationResults(req *dto.BatchEvaluateDatapackReq, userID int) (*dto.BatchEvaluateDatapackResp, error) {
	if req == nil {
		return nil, fmt.Errorf("batch evaluate datapack request is nil")
	}

	algorithms := make([]*dto.ContainerRef, 0, len(req.Specs))
	for i := range req.Specs {
		algorithms = append(algorithms, &req.Specs[i].Algorithm)
	}

	algorithmVersionResults, err := common.MapRefsToContainerVersions(algorithms, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	successItems := make([]dto.EvaluateDatapackItem, 0, len(req.Specs))
	failedItems := make([]string, 0)

	for i := range req.Specs {
		spec := &req.Specs[i]
		specIdentifier := fmt.Sprintf("spec[%d]: algorithm=%s, datapack=%s", i, spec.Algorithm.Name, spec.Datapack)

		algorithmVersion, exists := algorithmVersionResults[algorithms[i]]
		if !exists {
			failedItems = append(failedItems, fmt.Sprintf("%s - algorithm version not found", specIdentifier))
			continue
		}

		labelConditions := make([]map[string]string, 0, len(spec.FilterLabels))
		for _, label := range spec.FilterLabels {
			labelConditions = append(labelConditions, map[string]string{
				"key":   label.Key,
				"value": label.Value,
			})
		}

		executions, err := repository.ListExecutionsByDatapackFilter(database.DB, algorithmVersion.ID, spec.Datapack, labelConditions)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("%s - failed to query executions: %v", specIdentifier, err))
			continue
		}

		if len(executions) == 0 {
			failedItems = append(failedItems, fmt.Sprintf("%s - no executions found", specIdentifier))
			continue
		}

		refs := make([]dto.ExecutionRef, 0, len(executions))
		for _, execution := range executions {
			refs = append(refs, dto.NewExecutionGranularityRef(&execution))
		}

		evaluateRef := dto.EvaluateDatapackRef{
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

		item := dto.EvaluateDatapackItem{
			Algorithm:           algorithmVersion.Container.Name,
			AlgorithmVersion:    algorithmVersion.Name,
			EvaluateDatapackRef: evaluateRef,
		}
		successItems = append(successItems, item)
	}

	// Persist successful evaluations to the database
	persistEvaluations("datapack", successItems, func(item *dto.EvaluateDatapackItem) *database.Evaluation {
		return &database.Evaluation{
			AlgorithmName:    item.Algorithm,
			AlgorithmVersion: item.AlgorithmVersion,
			DatapackName:     item.Datapack,
			EvalType:         consts.EvalTypeDatapack,
			Status:           consts.CommonEnabled,
		}
	})

	resp := dto.BatchEvaluateDatapackResp{
		SuccessCount: len(successItems),
		SuccessItems: successItems,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}
	return &resp, nil
}

// ListDatasetEvaluationResults retrieves evaluation results for multiple dataset-algorithm pairs
func ListDatasetEvaluationResults(req *dto.BatchEvaluateDatasetReq, userID int) (*dto.BatchEvaluateDatasetResp, error) {
	if req == nil {
		return nil, fmt.Errorf("batch evaluate datapack request is nil")
	}

	algorithms := make([]*dto.ContainerRef, 0, len(req.Specs))
	datasets := make([]*dto.DatasetRef, 0, len(req.Specs))
	for i := range req.Specs {
		algorithms = append(algorithms, &req.Specs[i].Algorithm)
		datasets = append(datasets, &req.Specs[i].Dataset)
	}

	algorithmVersionResults, err := common.MapRefsToContainerVersions(algorithms, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	datasetVersionResults, err := common.MapRefsToDatasetVersions(datasets, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map dataset refs to versions: %w", err)
	}

	successItems := make([]dto.EvaluateDatasetItem, 0, len(req.Specs))
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

		executions, err := repository.ListExecutionsByDatasetFilter(database.DB, algorithmVersion.ID, datasetVersion.ID, labelConditions)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("%s - failed to query executions: %v", specIdentifier, err))
			continue
		}

		if len(executions) == 0 {
			failedItems = append(failedItems, fmt.Sprintf("%s - no executions found", specIdentifier))
			continue
		}

		executionMap := make(map[string][]database.Execution)
		for _, execution := range executions {
			name := execution.Datapack.Name
			if _, exists := executionMap[name]; !exists {
				executionMap[name] = make([]database.Execution, 0)
			} else {
				executionMap[name] = append(executionMap[name], execution)
			}
		}

		notExecutedDatapacks := []string{}
		for _, datapack := range datasetVersion.Datapacks {
			if _, exists := executionMap[datapack.Name]; !exists {
				notExecutedDatapacks = append(notExecutedDatapacks, datapack.Name)
			}
		}

		evaluateRefs := make([]dto.EvaluateDatapackRef, 0, len(executionMap))
		for datapack_name, groupedExecutions := range executionMap {
			refs := make([]dto.ExecutionRef, 0, len(groupedExecutions))
			for _, execution := range groupedExecutions {
				refs = append(refs, dto.NewExecutionGranularityRef(&execution))
			}

			evaluateRef := dto.EvaluateDatapackRef{
				Datapack:      datapack_name,
				ExecutionRefs: refs,
			}

			datapack := groupedExecutions[0].Datapack
			if datapack != nil {
				groundtruths, err := getGroundtruths(datapack)
				if err != nil {
					logrus.Warnf("failed to get groundtruth for datapack %s: %v", datapack_name, err)
				} else {
					evaluateRef.Groundtruths = groundtruths
				}
			}

			evaluateRefs = append(evaluateRefs, evaluateRef)
		}

		item := dto.EvaluateDatasetItem{
			Algorithm:            algorithmVersion.Container.Name,
			AlgorithmVersion:     algorithmVersion.Name,
			Dataset:              datasetVersion.Dataset.Name,
			DatasetVersion:       datasetVersion.Name,
			TotalCount:           len(datasetVersion.Datapacks),
			EvaluateRefs:         evaluateRefs,
			NotExecutedDatapacks: notExecutedDatapacks,
		}

		successItems = append(successItems, item)
	}

	// Persist successful evaluations to the database
	persistEvaluations("dataset", successItems, func(item *dto.EvaluateDatasetItem) *database.Evaluation {
		return &database.Evaluation{
			AlgorithmName:    item.Algorithm,
			AlgorithmVersion: item.AlgorithmVersion,
			DatasetName:      item.Dataset,
			DatasetVersion:   item.DatasetVersion,
			EvalType:         consts.EvalTypeDataset,
			Status:           consts.CommonEnabled,
		}
	})

	resp := dto.BatchEvaluateDatasetResp{
		SuccessCount: len(successItems),
		SuccessItems: successItems,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}
	return &resp, nil
}

// persistEvaluations batch-persists evaluation results to the database.
// The toEval function maps each item to a database.Evaluation (without ResultJSON).
func persistEvaluations[T any](evalType string, items []T, toEval func(*T) *database.Evaluation) {
	if len(items) == 0 {
		return
	}

	evals := make([]database.Evaluation, 0, len(items))
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

	if err := database.DB.Create(&evals).Error; err != nil {
		logrus.Warnf("failed to batch persist %d %s evaluations: %v", len(evals), evalType, err)
	}
}

// getGroundtruths extracts the ground truth from a datapack's engine configuration
func getGroundtruths(datapack *database.FaultInjection) ([]chaos.Groundtruth, error) {
	chaosGroundtruths := make([]chaos.Groundtruth, 0, len(datapack.Groundtruths))
	for _, gt := range datapack.Groundtruths {
		chaosGroundtruths = append(chaosGroundtruths, *gt.ConvertToChaosGroundtruth())
	}
	return chaosGroundtruths, nil
}
