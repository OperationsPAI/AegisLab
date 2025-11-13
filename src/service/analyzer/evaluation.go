package analyzer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"
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
	for _, spec := range req.Specs {
		algorithms = append(algorithms, &spec.Algorithm)
	}

	containerVersionResults, err := common.MapRefsToContainerVersions(algorithms, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	successItems := make([]dto.EvaluateDatapackItem, 0, len(req.Specs))
	failedItems := make([]string, 0)

	for i, spec := range req.Specs {
		specIdentifier := fmt.Sprintf("spec[%d]: algorithm=%s, datapack=%s", i, spec.Algorithm.Name, spec.Datapack)

		containerVersion, exists := containerVersionResults[&spec.Algorithm]
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

		executions, err := repository.ListExecutionsByDatapackFilter(database.DB, containerVersion.ID, spec.Datapack, labelConditions)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("%s - failed to query executions: %v", specIdentifier, err))
			continue
		}

		if len(executions) == 0 {
			failedItems = append(failedItems, fmt.Sprintf("%s - no executions found", specIdentifier))
			continue
		}

		refs := make([]dto.ExecutionGranularityRef, 0, len(executions))
		for _, execution := range executions {
			refs = append(refs, dto.NewExecutionGranularityRef(&execution))
		}

		item := dto.EvaluateDatapackItem{
			Algorithm:        containerVersion.Container.Name,
			AlgorithmVersion: containerVersion.Name,
			Datapack:         spec.Datapack,
			ExecutionRefs:    refs,
		}

		datapack := executions[0].Datapack
		if datapack != nil {
			groundTruth, err := getGroundtruth(datapack)
			if err != nil {
				logrus.Warnf("failed to get groundtruth for datapack %s: %v", spec.Datapack, err)
			} else {
				item.Groundtruth = *groundTruth
			}
		}

		successItems = append(successItems, item)
	}

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
	for _, spec := range req.Specs {
		algorithms = append(algorithms, &spec.Algorithm)
		datasets = append(datasets, &spec.Dataset)
	}

	containerVersionResults, err := common.MapRefsToContainerVersions(algorithms, consts.ContainerTypeAlgorithm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	datasetVersionResults, err := common.MapRefsToDatasetVersions(datasets, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map dataset refs to versions: %w", err)
	}

	successItems := make([]dto.EvaluateDatasetItem, 0, len(req.Specs))
	failedItems := make([]string, 0)

	for i, spec := range req.Specs {
		specIdentifier := fmt.Sprintf("spec[%d]: algorithm=%s, dataset=%s", i, spec.Algorithm.Name, spec.Dataset.Name)

		containerVersion, exists := containerVersionResults[&spec.Algorithm]
		if !exists {
			failedItems = append(failedItems, fmt.Sprintf("%s - algorithm version not found", specIdentifier))
			continue
		}

		datasetVersion, exists := datasetVersionResults[&spec.Dataset]
		if !exists {
			failedItems = append(failedItems, fmt.Sprintf("%s - dataset version not found", specIdentifier))
			continue
		}

		labelConditions := dto.ConvertLabelItemsToConditions(spec.FilterLabels)

		executions, err := repository.ListExecutionsByDatasetFilter(database.DB, containerVersion.ID, datasetVersion.ID, labelConditions)
		if err != nil {
			failedItems = append(failedItems, fmt.Sprintf("%s - failed to query executions: %v", specIdentifier, err))
			continue
		}

		if len(executions) == 0 {
			failedItems = append(failedItems, fmt.Sprintf("%s - no executions found", specIdentifier))
			continue
		}

		refs := make([]dto.EvaluateDatapackRef, 0, len(executions))
		for _, execution := range executions {
			refs = append(refs, dto.NewEvaluateDatapackRef(execution.Datapack.Name, &execution))
		}

		executedDatapacks := make([]int, 0, len(executions))
		for _, execution := range executions {
			executedDatapacks = append(executedDatapacks, execution.Datapack.ID)
		}
		executedDatapacks = utils.ToUniqueSlice(executedDatapacks)

		item := dto.EvaluateDatasetItem{
			Algorithm:        containerVersion.Container.Name,
			AlgorithmVersion: containerVersion.Name,
			Dataset:          datasetVersion.Dataset.Name,
			DatasetVersion:   datasetVersion.Name,
			TotalCount:       len(datasetVersion.Injections),
			ExecutedCount:    len(executedDatapacks),
			EvaluateRefs:     refs,
		}

		successItems = append(successItems, item)
	}

	resp := dto.BatchEvaluateDatasetResp{
		SuccessCount: len(successItems),
		SuccessItems: successItems,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}
	return &resp, nil
}

// getGroundtruth extracts the ground truth from a datapack's engine configuration
func getGroundtruth(datapack *database.FaultInjection) (*chaos.Groundtruth, error) {
	var node chaos.Node
	if err := json.Unmarshal([]byte(datapack.EngineConfig), &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chaos-experiment node for datapack %s: %v", datapack.Name, err)
	}

	conf, err := chaos.NodeToStruct[chaos.InjectionConf](&node)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chaos-experiment node to InjectionConf for datapack %s: %v", datapack.Name, err)
	}

	groundtruth, err := conf.GetGroundtruth()
	if err != nil {
		return nil, fmt.Errorf("failed to get ground truth for datapack %s: %v", datapack.Name, err)
	}

	return &groundtruth, nil
}
