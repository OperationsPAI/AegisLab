package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CreateRole handles the business logic for creating a new role
func CreateRole(req *dto.CreateRoleReq) (*dto.RoleResp, error) {
	role := req.ConvertToRole()

	var createdRole *database.Role
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.CreateRole(tx, role); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: role with name %s already exists", consts.ErrAlreadyExists, role.Name)
			}
			return err
		}

		createdRole = role
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewRoleResp(createdRole), nil
}

// DeleteRole deletes an existing role by marking its status as deleted
func DeleteRole(roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get role: %w", err)
		}

		if role.IsSystem {
			return fmt.Errorf("%w: cannot delete system role", consts.ErrPermissionDenied)
		}

		if _, err := repository.RemoveContainersFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove containers with role: %w", err)
		}
		if _, err := repository.RemoveDatasetsFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove datasets with role: %w", err)
		}
		if _, err := repository.RemoveProjectsFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove projects with role: %w", err)
		}

		if err := repository.RemovePermissionsFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove permissions with role: %w", err)
		}
		if err := repository.RemoveUsersFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove users with role: %w", err)
		}

		row, err := repository.DeleteRole(tx, role.ID)
		if err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}
		if row == 0 {
			return fmt.Errorf("%w: role id %d not found", consts.ErrNotFound, roleID)
		}

		return nil
	})
}

// GetRoleDetail retrieves detailed information about a role by its ID
func GetRoleDetail(roleID int) (*dto.RoleDetailResp, error) {
	role, err := repository.GetRoleByID(database.DB, roleID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role with ID %d not found", consts.ErrNotFound, roleID)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	resp := dto.NewRoleDetailResp(role)

	userCount, err := repository.GetRoleUserCount(database.DB, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role user count: %w", err)
	}
	resp.UserCount = userCount

	permissions, err := repository.GetRolePermissions(database.DB, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	resp.Permissions = make([]dto.PermissionResp, len(permissions))
	for _, permission := range permissions {
		resp.Permissions = append(resp.Permissions, *dto.NewPermissionResp(&permission))
	}

	return resp, nil
}

// ListRoles lists roles based on the provided filters
func ListRoles(req *dto.ListRoleReq) (*dto.ListResp[dto.RoleResp], error) {
	limit, offset := req.ToGormParams()

	roles, total, err := repository.ListRoles(database.DB, limit, offset, req.IsSystem, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	roleResps := make([]dto.RoleResp, len(roles))
	for i, role := range roles {
		roleResps[i] = *dto.NewRoleResp(&role)
	}

	resp := dto.ListResp[dto.RoleResp]{
		Items:      roleResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// SearchRoles searches for roles based on the provided search request
func SearchRoles(req *dto.SearchReq) (*dto.SearchResp[dto.RoleResp], error) {
	searchResult, err := repository.ExecuteSearch(database.DB, req, database.Role{})
	if err != nil {
		return nil, fmt.Errorf("failed to search roles: %w", err)
	}

	var roleResponses []dto.RoleResp
	for _, role := range searchResult.Items {
		roleResponses = append(roleResponses, *dto.NewRoleResp(&role))
	}

	resp := dto.SearchResp[dto.RoleResp]{
		Items:      roleResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}
	return &resp, nil
}

// UpdateRole updates an existing role
func UpdateRole(req *dto.UpdateRoleReq, roleID int) (*dto.RoleResp, error) {
	var updatedRole *database.Role

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingRole, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get role: %w", err)
		}

		if existingRole.IsSystem {
			return fmt.Errorf("%w: cannot update system role", consts.ErrPermissionDenied)
		}

		req.PatchRoleModel(existingRole)

		if err := repository.UpdateRole(tx, existingRole); err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}

		updatedRole = existingRole
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewRoleResp(updatedRole), nil
}
