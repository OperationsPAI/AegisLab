package producer

import (
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"fmt"
)

// ListEvaluations lists evaluations with pagination
func ListEvaluations(req *dto.ListEvaluationReq) (*dto.ListResp[dto.EvaluationResp], error) {
	limit, offset := req.ToGormParams()

	evaluations, total, err := repository.ListEvaluations(database.DB, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}

	evalResps := make([]dto.EvaluationResp, 0, len(evaluations))
	for _, eval := range evaluations {
		evalResps = append(evalResps, *dto.NewEvaluationResp(&eval))
	}

	resp := dto.ListResp[dto.EvaluationResp]{
		Items:      evalResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// GetEvaluation retrieves a single evaluation by ID
func GetEvaluation(id int) (*dto.EvaluationResp, error) {
	eval, err := repository.GetEvaluationByID(database.DB, id)
	if err != nil {
		return nil, err
	}

	return dto.NewEvaluationResp(eval), nil
}

// DeleteEvaluation soft-deletes an evaluation by ID
func DeleteEvaluation(id int) error {
	return repository.DeleteEvaluation(database.DB, id)
}
