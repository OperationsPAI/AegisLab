package rbacmodule

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the RBAC operations consumed by the HTTP handler.
type HandlerService interface {
	CreateRole(context.Context, *CreateRoleReq) (*RoleResp, error)
	DeleteRole(context.Context, int) error
	GetRole(context.Context, int) (*RoleDetailResp, error)
	ListRoles(context.Context, *ListRoleReq) (*dto.ListResp[RoleResp], error)
	UpdateRole(context.Context, *UpdateRoleReq, int) (*RoleResp, error)
	AssignRolePermissions(context.Context, []int, int) error
	RemoveRolePermissions(context.Context, []int, int) error
	ListUsersFromRole(context.Context, int) ([]UserListItem, error)
	GetPermission(context.Context, int) (*PermissionDetailResp, error)
	ListPermissions(context.Context, *ListPermissionReq) (*dto.ListResp[PermissionResp], error)
	ListRolesFromPermission(context.Context, int) ([]RoleResp, error)
	GetResource(context.Context, int) (*ResourceResp, error)
	ListResources(context.Context, *ListResourceReq) (*dto.ListResp[ResourceResp], error)
	ListResourcePermissions(context.Context, int) ([]PermissionResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
