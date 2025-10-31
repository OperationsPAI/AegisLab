package service

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
func CreateRole(req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
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

	var resp dto.RoleResponse
	resp.ConvertFromRole(createdRole)
	return &resp, nil
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

		role.Status = consts.CommonDeleted
		if err := repository.UpdateRole(tx, role); err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}

		if err := repository.RemoveAllContainersFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove containers with role: %w", err)
		}
		if err := repository.RemoveAllProjectsFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove projects with role: %w", err)
		}
		if err := repository.RemoveAllPermissionsFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove permissions with role: %w", err)
		}
		if err := repository.RemoveAllUsersFromRole(tx, role.ID); err != nil {
			return fmt.Errorf("failed to remove users with role: %w", err)
		}

		return nil
	})
}

// GetRoleDetail retrieves detailed information about a role by its ID
func GetRoleDetail(roleID int) (*dto.RoleDetailResponse, error) {
	role, err := repository.GetRoleByID(database.DB, roleID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role with ID %d not found", consts.ErrNotFound, roleID)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	var resp dto.RoleDetailResponse
	resp.ConvertFromRole(role)

	userCount, err := repository.GetRoleUserCount(database.DB, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role user count: %w", err)
	}
	resp.UserCount = userCount

	permissions, err := repository.GetRolePermissions(database.DB, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	resp.Permissions = make([]dto.PermissionResponse, len(permissions))
	for _, permission := range permissions {
		var permResp dto.PermissionResponse
		permResp.ConvertFromPermission(&permission)
		resp.Permissions = append(resp.Permissions, permResp)
	}

	return &resp, nil
}

// ListRoles lists roles based on the provided filters
func ListRoles(req *dto.ListRoleRequest) (*dto.ListResponse[dto.RoleResponse], error) {
	limit, offset := req.ToGormParams()

	roles, total, err := repository.ListRoles(database.DB, limit, offset, req.IsSystem, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	roleResps := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		var roleResp dto.RoleResponse
		roleResp.ConvertFromRole(&role)
		roleResps[i] = roleResp
	}

	resp := dto.ListResponse[dto.RoleResponse]{
		Items:      roleResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// SearchRoles searches for roles based on the provided search request
func SearchRoles(req *dto.SearchRequest) (*dto.SearchResponse[dto.RoleResponse], error) {
	searchResult, err := repository.ExecuteSearch(database.DB, req, database.Role{})
	if err != nil {
		return nil, fmt.Errorf("failed to search roles: %w", err)
	}

	var roleResponses []dto.RoleResponse
	for _, role := range searchResult.Items {
		var response dto.RoleResponse
		response.ConvertFromRole(&role)
		roleResponses = append(roleResponses, response)
	}

	resp := dto.SearchResponse[dto.RoleResponse]{
		Items:      roleResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}
	return &resp, nil
}

// UpdateRole updates an existing role
func UpdateRole(req *dto.UpdateRoleRequest, roleID int) (*dto.RoleResponse, error) {
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

	var resp dto.RoleResponse
	resp.ConvertFromRole(updatedRole)
	return &resp, nil
}

// AssginPermissionsToRole assigns permissions to a role
func AssginPermissionsToRole(roleID int, permissionIDs []int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if role.IsSystem {
			return fmt.Errorf("%w: cannot assign permissions to system role", consts.ErrPermissionDenied)
		}

		for _, permissionID := range permissionIDs {
			if err := repository.CreateRolePermission(tx, &database.RolePermission{
				RoleID:       role.ID,
				PermissionID: permissionID,
			}); err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					return fmt.Errorf("%w: permission with ID %d already assigned to role", consts.ErrAlreadyExists, permissionID)
				}
				return fmt.Errorf("failed to assign permission to role: %w", err)
			}
		}

		return nil
	})
}

// RemovePermissionsFromRole removes permissions from a role
func RemovePermissionsFromRole(roleID int, permissionIDs []int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if role.IsSystem {
			return fmt.Errorf("%w: cannot remove permissions of system role", consts.ErrPermissionDenied)
		}

		for _, permissionID := range permissionIDs {
			if err := repository.DeleteRolePermission(tx, role.ID, permissionID); err != nil {
				if errors.Is(err, consts.ErrNotFound) {
					return fmt.Errorf("%w: permission with ID %d not assigned to role", consts.ErrNotFound, permissionID)
				}
				return fmt.Errorf("failed to assign permission to role: %w", err)
			}
		}

		return nil
	})
}

// ListUsersFromRole lists users assigned to a specific role
func ListUsersFromRole(roleID int) ([]dto.UserResponse, error) {
	role, err := repository.GetRoleByID(database.DB, roleID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role not found", consts.ErrNotFound)
		}
		return nil, err
	}

	users, err := repository.GetRoleUsers(database.DB, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	var userResps []dto.UserResponse
	for _, user := range users {
		var response dto.UserResponse
		response.ConvertFromUser(&user)
		userResps = append(userResps, response)
	}

	return userResps, nil
}
