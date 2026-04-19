package container

import (
	"context"

	"aegis/dto"

	"mime/multipart"
)

// HandlerService captures the container operations consumed by the HTTP handler.
type HandlerService interface {
	CreateContainer(context.Context, *CreateContainerReq, int) (*ContainerResp, error)
	DeleteContainer(context.Context, int) error
	GetContainer(context.Context, int) (*ContainerDetailResp, error)
	ListContainers(context.Context, *ListContainerReq) (*dto.ListResp[ContainerResp], error)
	UpdateContainer(context.Context, *UpdateContainerReq, int) (*ContainerResp, error)
	ManageContainerLabels(context.Context, *ManageContainerLabelReq, int) (*ContainerResp, error)
	CreateContainerVersion(context.Context, *CreateContainerVersionReq, int, int) (*ContainerVersionResp, error)
	DeleteContainerVersion(context.Context, int) error
	GetContainerVersion(context.Context, int, int) (*ContainerVersionDetailResp, error)
	ListContainerVersions(context.Context, *ListContainerVersionReq, int) (*dto.ListResp[ContainerVersionResp], error)
	UpdateContainerVersion(context.Context, *UpdateContainerVersionReq, int, int) (*ContainerVersionResp, error)
	SetContainerVersionImage(context.Context, *SetContainerVersionImageReq, int) (*SetContainerVersionImageResp, error)
	SubmitContainerBuilding(context.Context, *SubmitBuildContainerReq, string, int) (*SubmitContainerBuildResp, error)
	UploadHelmChart(context.Context, *multipart.FileHeader, int, int, int) (*UploadHelmChartResp, error)
	UploadHelmValueFile(context.Context, *multipart.FileHeader, int, int, int) (*UploadHelmValueFileResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
