package rbacmodule

import (
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRole(_ context.Context, req *CreateRoleReq) (*RoleResp, error) {
	role := req.ConvertToRole()

	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if err := NewRepository(tx).createRoleRecord(role); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: role with name %s already exists", consts.ErrAlreadyExists, role.Name)
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return NewRoleResp(role), nil
}

func (s *Service) DeleteRole(_ context.Context, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		rows, err := NewRepository(tx).deleteRoleCascade(roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: role id %d not found", consts.ErrNotFound, roleID)
		}
		return nil
	})
}

func (s *Service) GetRole(_ context.Context, roleID int) (*RoleDetailResp, error) {
	role, userCount, permissions, err := s.repo.loadRoleDetail(roleID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role with ID %d not found", consts.ErrNotFound, roleID)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	resp := NewRoleDetailResp(role)
	resp.UserCount = userCount

	resp.Permissions = make([]PermissionResp, 0, len(permissions))
	for _, permission := range permissions {
		resp.Permissions = append(resp.Permissions, *NewPermissionResp(&permission))
	}

	return resp, nil
}

func (s *Service) ListRoles(_ context.Context, req *ListRoleReq) (*dto.ListResp[RoleResp], error) {
	limit, offset := req.ToGormParams()
	roles, total, err := s.repo.listRoleViews(limit, offset, req.IsSystem, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	items := make([]RoleResp, len(roles))
	for i, role := range roles {
		items[i] = *NewRoleResp(&role)
	}

	return &dto.ListResp[RoleResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateRole(_ context.Context, req *UpdateRoleReq, roleID int) (*RoleResp, error) {
	var updatedRole *model.Role

	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		role, err := repo.updateMutableRole(roleID, func(existingRole *model.Role) {
			req.PatchRoleModel(existingRole)
		})
		if err != nil {
			return err
		}
		updatedRole = role
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewRoleResp(updatedRole), nil
}

func (s *Service) AssignRolePermissions(_ context.Context, permissionIDs []int, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.assignRolePermissions(roleID, permissionIDs); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: role already has one or more of these permissions", consts.ErrAlreadyExists)
			}
			return fmt.Errorf("failed to assign permissions to role: %w", err)
		}
		return nil
	})
}

func (s *Service) RemoveRolePermissions(_ context.Context, permissionIDs []int, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.removeRolePermissions(roleID, permissionIDs); err != nil {
			return fmt.Errorf("failed to remove permissions from role: %w", err)
		}
		return nil
	})
}

func (s *Service) ListUsersFromRole(_ context.Context, roleID int) ([]UserListItem, error) {
	_, users, err := s.repo.listUsersFromRole(roleID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role not found", consts.ErrNotFound)
		}
		return nil, err
	}

	userResps := make([]UserListItem, 0, len(users))
	for _, user := range users {
		userResps = append(userResps, *NewUserListItem(&user))
	}
	return userResps, nil
}

func (s *Service) GetPermission(_ context.Context, permissionID int) (*PermissionDetailResp, error) {
	permission, err := s.repo.getPermissionDetail(permissionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: permission not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return NewPermissionDetailResp(permission), nil
}

func (s *Service) ListPermissions(_ context.Context, req *ListPermissionReq) (*dto.ListResp[PermissionResp], error) {
	limit, offset := req.ToGormParams()
	permissions, total, err := s.repo.listPermissionViews(limit, offset, req.Action, req.IsSystem, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	items := make([]PermissionResp, len(permissions))
	for i, permission := range permissions {
		items[i] = *NewPermissionResp(&permission)
	}

	return &dto.ListResp[PermissionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) ListRolesFromPermission(_ context.Context, permissionID int) ([]RoleResp, error) {
	_, roles, err := s.repo.listRolesFromPermission(permissionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: permission not found", consts.ErrNotFound)
		}
		return nil, err
	}

	items := make([]RoleResp, 0, len(roles))
	for _, role := range roles {
		items = append(items, *NewRoleResp(&role))
	}
	return items, nil
}

func (s *Service) GetResource(_ context.Context, resourceID int) (*ResourceResp, error) {
	resource, err := s.repo.getResourceDetail(resourceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: resource with ID %d not found", consts.ErrNotFound, resourceID)
		}
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}
	return NewResourceResp(resource), nil
}

func (s *Service) ListResources(_ context.Context, req *ListResourceReq) (*dto.ListResp[ResourceResp], error) {
	limit, offset := req.ToGormParams()
	resources, total, err := s.repo.listResourceViews(limit, offset, req.Type, req.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	items := make([]ResourceResp, 0, len(resources))
	for i := range resources {
		items = append(items, *NewResourceResp(&resources[i]))
	}

	return &dto.ListResp[ResourceResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) ListResourcePermissions(_ context.Context, resourceID int) ([]PermissionResp, error) {
	_, permissions, err := s.repo.listResourcePermissions(resourceID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: resource with ID %d not found", consts.ErrNotFound, resourceID)
		}
		return nil, err
	}

	items := make([]PermissionResp, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, *NewPermissionResp(&permission))
	}
	return items, nil
}
