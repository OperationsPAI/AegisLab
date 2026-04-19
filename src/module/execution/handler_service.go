package execution

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the execution operations consumed by the HTTP handler.
type HandlerService interface {
	ListProjectExecutions(context.Context, *ListExecutionReq, int) (*dto.ListResp[ExecutionResp], error)
	SubmitAlgorithmExecution(context.Context, *SubmitExecutionReq, string, int) (*SubmitExecutionResp, error)
	ListExecutions(context.Context, *ListExecutionReq) (*dto.ListResp[ExecutionResp], error)
	GetExecution(context.Context, int) (*ExecutionDetailResp, error)
	ListAvailableLabels(context.Context) ([]dto.LabelItem, error)
	ManageLabels(context.Context, *ManageExecutionLabelReq, int) (*ExecutionResp, error)
	BatchDelete(context.Context, *BatchDeleteExecutionReq) error
	UploadDetectorResults(context.Context, *UploadDetectorResultReq, int) (*UploadExecutionResultResp, error)
	UploadGranularityResults(context.Context, *UploadGranularityResultReq, int) (*UploadExecutionResultResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
