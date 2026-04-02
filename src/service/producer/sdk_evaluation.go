package producer

import (
	"fmt"

	"aegis/database"
	"aegis/dto"
	"aegis/repository"
)

// ListSDKEvaluations lists SDK evaluation samples with pagination and filtering.
func ListSDKEvaluations(req *dto.ListSDKEvaluationReq) (*dto.ListResp[database.SDKEvaluationSample], error) {
	limit, offset := req.ToGormParams()

	items, total, err := repository.ListSDKEvaluations(database.DB, req.ExpID, req.Stage, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list SDK evaluations: %w", err)
	}

	return &dto.ListResp[database.SDKEvaluationSample]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

// GetSDKEvaluation retrieves a single SDK evaluation sample by ID.
func GetSDKEvaluation(id int) (*database.SDKEvaluationSample, error) {
	item, err := repository.GetSDKEvaluationByID(database.DB, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// ListSDKExperiments returns all distinct experiment IDs.
func ListSDKExperiments() (*dto.SDKExperimentListResp, error) {
	expIDs, err := repository.ListSDKExperiments(database.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to list SDK experiments: %w", err)
	}
	return &dto.SDKExperimentListResp{Experiments: expIDs}, nil
}

// ListSDKDatasetSamples lists SDK dataset samples with pagination and filtering.
func ListSDKDatasetSamples(req *dto.ListSDKDatasetSampleReq) (*dto.ListResp[database.SDKDatasetSample], error) {
	limit, offset := req.ToGormParams()

	items, total, err := repository.ListSDKDatasetSamples(database.DB, req.Dataset, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list SDK dataset samples: %w", err)
	}

	return &dto.ListResp[database.SDKDatasetSample]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}
