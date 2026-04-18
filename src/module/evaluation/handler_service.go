package evaluationmodule

import (
	"context"

	"aegis/dto"
)

// HandlerService captures evaluation operations consumed by the HTTP handler.
type HandlerService interface {
	ListDatapackEvaluationResults(context.Context, *BatchEvaluateDatapackReq, int) (*BatchEvaluateDatapackResp, error)
	ListDatasetEvaluationResults(context.Context, *BatchEvaluateDatasetReq, int) (*BatchEvaluateDatasetResp, error)
	ListEvaluations(context.Context, *ListEvaluationReq) (*dto.ListResp[EvaluationResp], error)
	GetEvaluation(context.Context, int) (*EvaluationResp, error)
	DeleteEvaluation(context.Context, int) error
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
