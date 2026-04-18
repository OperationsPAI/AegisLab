package gatewayapp

import (
	"context"

	"aegis/dto"
	rbacmodule "aegis/module/rbac"
)

type rbacIAMClient interface {
	Enabled() bool
	CreateRole(context.Context, *rbacmodule.CreateRoleReq) (*rbacmodule.RoleResp, error)
	DeleteRole(context.Context, int) error
	GetRole(context.Context, int) (*rbacmodule.RoleDetailResp, error)
	ListRoles(context.Context, *rbacmodule.ListRoleReq) (*dto.ListResp[rbacmodule.RoleResp], error)
	UpdateRole(context.Context, *rbacmodule.UpdateRoleReq, int) (*rbacmodule.RoleResp, error)
	AssignRolePermissions(context.Context, int, []int) error
	RemoveRolePermissions(context.Context, int, []int) error
	ListUsersFromRole(context.Context, int) ([]rbacmodule.UserListItem, error)
	GetPermission(context.Context, int) (*rbacmodule.PermissionDetailResp, error)
	ListPermissions(context.Context, *rbacmodule.ListPermissionReq) (*dto.ListResp[rbacmodule.PermissionResp], error)
	ListRolesFromPermission(context.Context, int) ([]rbacmodule.RoleResp, error)
	GetResource(context.Context, int) (*rbacmodule.ResourceResp, error)
	ListResources(context.Context, *rbacmodule.ListResourceReq) (*dto.ListResp[rbacmodule.ResourceResp], error)
	ListResourcePermissions(context.Context, int) ([]rbacmodule.PermissionResp, error)
}

type remoteAwareRBACService struct {
	rbacmodule.HandlerService
	iam rbacIAMClient
}

func (s remoteAwareRBACService) CreateRole(ctx context.Context, req *rbacmodule.CreateRoleReq) (*rbacmodule.RoleResp, error) {
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

func (s remoteAwareRBACService) GetRole(ctx context.Context, roleID int) (*rbacmodule.RoleDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetRole(ctx, roleID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListRoles(ctx context.Context, req *rbacmodule.ListRoleReq) (*dto.ListResp[rbacmodule.RoleResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListRoles(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) UpdateRole(ctx context.Context, req *rbacmodule.UpdateRoleReq, roleID int) (*rbacmodule.RoleResp, error) {
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

func (s remoteAwareRBACService) ListUsersFromRole(ctx context.Context, roleID int) ([]rbacmodule.UserListItem, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListUsersFromRole(ctx, roleID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) GetPermission(ctx context.Context, permissionID int) (*rbacmodule.PermissionDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetPermission(ctx, permissionID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListPermissions(ctx context.Context, req *rbacmodule.ListPermissionReq) (*dto.ListResp[rbacmodule.PermissionResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListPermissions(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListRolesFromPermission(ctx context.Context, permissionID int) ([]rbacmodule.RoleResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListRolesFromPermission(ctx, permissionID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) GetResource(ctx context.Context, resourceID int) (*rbacmodule.ResourceResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetResource(ctx, resourceID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListResources(ctx context.Context, req *rbacmodule.ListResourceReq) (*dto.ListResp[rbacmodule.ResourceResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListResources(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareRBACService) ListResourcePermissions(ctx context.Context, resourceID int) ([]rbacmodule.PermissionResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListResourcePermissions(ctx, resourceID)
	}
	return nil, missingRemoteDependency("iam-service")
}
