package label

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the label operations consumed by HTTP and resource gRPC handlers.
type HandlerService interface {
	BatchDelete(context.Context, []int) error
	Create(context.Context, *CreateLabelReq) (*LabelResp, error)
	Delete(context.Context, int) error
	GetDetail(context.Context, int) (*LabelDetailResp, error)
	List(context.Context, *ListLabelReq) (*dto.ListResp[LabelResp], error)
	Update(context.Context, *UpdateLabelReq, int) (*LabelResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
