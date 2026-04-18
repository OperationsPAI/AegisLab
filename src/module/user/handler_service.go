package usermodule

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the user operations consumed by the HTTP handler.
type HandlerService interface {
	CreateUser(context.Context, *CreateUserReq) (*UserResp, error)
	DeleteUser(context.Context, int) error
	GetUserDetail(context.Context, int) (*UserDetailResp, error)
	ListUsers(context.Context, *ListUserReq) (*dto.ListResp[UserResp], error)
	UpdateUser(context.Context, *UpdateUserReq, int) (*UserResp, error)
	AssignRole(context.Context, int, int) error
	RemoveRole(context.Context, int, int) error
	AssignPermissions(context.Context, *AssignUserPermissionReq, int) error
	RemovePermissions(context.Context, *RemoveUserPermissionReq, int) error
	AssignContainer(context.Context, int, int, int) error
	RemoveContainer(context.Context, int, int) error
	AssignDataset(context.Context, int, int, int) error
	RemoveDataset(context.Context, int, int) error
	AssignProject(context.Context, int, int, int) error
	RemoveProject(context.Context, int, int) error
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
