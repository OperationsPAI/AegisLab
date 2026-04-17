package sdkmodule

import (
	"context"
	"fmt"

	"aegis/dto"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListEvaluations(_ context.Context, req *ListSDKEvaluationReq) (*dto.ListResp[SDKEvaluationSample], error) {
	limit, offset := req.ToGormParams()
	items, total, err := s.repo.ListSDKEvaluations(req.ExpID, req.Stage, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list SDK evaluations: %w", err)
	}
	return &dto.ListResp[SDKEvaluationSample]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) GetEvaluation(_ context.Context, id int) (*SDKEvaluationSample, error) {
	return s.repo.GetSDKEvaluationByID(id)
}

func (s *Service) ListExperiments(_ context.Context) (*SDKExperimentListResp, error) {
	items, err := s.repo.ListSDKExperiments()
	if err != nil {
		return nil, fmt.Errorf("failed to list SDK experiments: %w", err)
	}
	return &SDKExperimentListResp{Experiments: items}, nil
}

func (s *Service) ListDatasetSamples(_ context.Context, req *ListSDKDatasetSampleReq) (*dto.ListResp[SDKDatasetSample], error) {
	limit, offset := req.ToGormParams()
	items, total, err := s.repo.ListSDKDatasetSamples(req.Dataset, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list SDK dataset samples: %w", err)
	}
	return &dto.ListResp[SDKDatasetSample]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}
