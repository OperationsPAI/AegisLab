package projectmodule

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the project operations consumed by the HTTP handler.
type HandlerService interface {
	CreateProject(context.Context, *CreateProjectReq, int) (*ProjectResp, error)
	DeleteProject(context.Context, int) error
	GetProjectDetail(context.Context, int) (*ProjectDetailResp, error)
	ListProjects(context.Context, *ListProjectReq) (*dto.ListResp[ProjectResp], error)
	UpdateProject(context.Context, *UpdateProjectReq, int) (*ProjectResp, error)
	ManageProjectLabels(context.Context, *ManageProjectLabelReq, int) (*ProjectResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
