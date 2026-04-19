package gateway

import (
	"context"

	"aegis/dto"
	rbac "aegis/module/rbac"
)

type rbacIAMClient interface {
	Enabled() bool
	CreateRole(context.Context, *rbac.CreateRoleReq) (*rbac.RoleResp, error)
	DeleteRole(context.Context, int) error
	GetRole(context.Context, int) (*rbac.RoleDetailResp, error)
	ListRoles(context.Context, *rbac.ListRoleReq) (*dto.ListResp[rbac.RoleResp], error)
	UpdateRole(context.Context, *rbac.UpdateRoleReq, int) (*rbac.RoleResp, error)
	AssignRolePermissions(context.Context, int, []int) error
	RemoveRolePermissions(context.Context, int, []int) error
	ListUsersFromRole(context.Context, int) ([]rbac.UserListItem, error)
	GetPermission(context.Context, int) (*rbac.PermissionDetailResp, error)
	ListPermissions(context.Context, *rbac.ListPermissionReq) (*dto.ListResp[rbac.PermissionResp], error)
	ListRolesFromPermission(context.Context, int) ([]rbac.RoleResp, error)
	GetResource(context.Context, int) (*rbac.ResourceResp, error)
	ListResources(context.Context, *rbac.ListResourceReq) (*dto.ListResp[rbac.ResourceResp], error)
	ListResourcePermissions(context.Context, int) ([]rbac.PermissionResp, error)
}

type remoteAwareRBACService struct {
	rbac.HandlerService
	iam rbacIAMClient
}

func (s remoteAwareRBACService) CreateRole(ctx context.Context, req *rbac.CreateRoleReq) (*rbac.RoleResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CreateRole(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) DeleteRole(ctx context.Context, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DeleteRole(ctx, roleID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) GetRole(ctx context.Context, roleID int) (*rbac.RoleDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetRole(ctx, roleID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListRoles(ctx context.Context, req *rbac.ListRoleReq) (*dto.ListResp[rbac.RoleResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListRoles(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) UpdateRole(ctx context.Context, req *rbac.UpdateRoleReq, roleID int) (*rbac.RoleResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.UpdateRole(ctx, req, roleID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) AssignRolePermissions(ctx context.Context, permissionIDs []int, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignRolePermissions(ctx, roleID, permissionIDs)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) RemoveRolePermissions(ctx context.Context, permissionIDs []int, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveRolePermissions(ctx, roleID, permissionIDs)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListUsersFromRole(ctx context.Context, roleID int) ([]rbac.UserListItem, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListUsersFromRole(ctx, roleID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) GetPermission(ctx context.Context, permissionID int) (*rbac.PermissionDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetPermission(ctx, permissionID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListPermissions(ctx context.Context, req *rbac.ListPermissionReq) (*dto.ListResp[rbac.PermissionResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListPermissions(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListRolesFromPermission(ctx context.Context, permissionID int) ([]rbac.RoleResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListRolesFromPermission(ctx, permissionID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) GetResource(ctx context.Context, resourceID int) (*rbac.ResourceResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetResource(ctx, resourceID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListResources(ctx context.Context, req *rbac.ListResourceReq) (*dto.ListResp[rbac.ResourceResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListResources(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListResourcePermissions(ctx context.Context, resourceID int) ([]rbac.PermissionResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListResourcePermissions(ctx, resourceID)
	}
	return nil, missingRemoteDependency("iam-service")
}
