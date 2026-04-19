package chaossystem

import (
	"context"

	"aegis/dto"
)

// HandlerService captures chaos system operations consumed by HTTP and resource gRPC handlers.
type HandlerService interface {
	ListSystems(context.Context, *ListChaosSystemReq) (*dto.ListResp[ChaosSystemResp], error)
	GetSystem(context.Context, int) (*ChaosSystemResp, error)
	CreateSystem(context.Context, *CreateChaosSystemReq) (*ChaosSystemResp, error)
	UpdateSystem(context.Context, int, *UpdateChaosSystemReq) (*ChaosSystemResp, error)
	DeleteSystem(context.Context, int) error
	UpsertMetadata(context.Context, int, *BulkUpsertSystemMetadataReq) error
	ListMetadata(context.Context, int, string) ([]SystemMetadataResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
