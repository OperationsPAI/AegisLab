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

// CreatePermission creates a new permission
func CreatePermission(req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	permission := req.ConvertToPermission()

	var createdPermission *database.Permission
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		resource, err := repository.GetResourceByID(tx, req.ResourceID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: resource not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get resource: %w", err)
		}

		permission.Name = getPermissionName(req.Action, resource.Name)
		permission.ResourceID = resource.ID
		if req.DisplayName != "" {
			permission.DisplayName = req.DisplayName
		} else {
			permission.DisplayName = getPermissionDisplayName(req.Action, resource.DisplayName)
		}

		if err := repository.CreatePermission(tx, permission); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return fmt.Errorf("%w: permission already exists", consts.ErrAlreadyExists)
			}
			return fmt.Errorf("failed to create permission: %w", err)
		}

		permission.Resource = resource
		createdPermission = permission
		return nil
	})
	if err != nil {
		return nil, err
	}

	var resp dto.PermissionResponse
	resp.ConvertFromPermission(createdPermission)
	return &resp, nil
}

// DeletePermission deteles an existing permission by marking its status as deleted
func DeletePermission(permissionID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		permission, err := repository.GetPermissionByID(database.DB, permissionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: permission not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get permission: %w", err)
		}

		if permission.IsSystem {
			return fmt.Errorf("%w: cannot delete system permission", consts.ErrPermissionDenied)
		}

		permission.Status = consts.CommonDeleted
		if err := repository.UpdatePermission(tx, permission); err != nil {
			return fmt.Errorf("failed to update container: %w", err)
		}

		if err := repository.RemoveAllRolesFromPermission(tx, permissionID); err != nil {
			return fmt.Errorf("failed to remove roles from permission: %w", err)
		}
		if err := repository.RemoveAllUsersFromPermission(tx, permissionID); err != nil {
			return fmt.Errorf("failed to remove users from permission: %w", err)
		}

		return nil
	})
}

// GetPermissionDetail retrieves detailed information about a permission by its ID
func GetPermissionDetail(permissionID int) (*dto.PermissionDetailResponse, error) {
	permission, err := repository.GetPermissionByID(database.DB, permissionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: permission not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	var resp dto.PermissionDetailResponse
	resp.ConvertFromPermission(permission)
	return &resp, nil
}

// ListPermissions lists permissions based on the provided request parameters
func ListPermissions(req *dto.ListPermissionRequest) (*dto.ListResponse[dto.PermissionResponse], error) {
	limit, offset := req.ToGormParams()

	permissions, total, err := repository.ListPermissions(database.DB, limit, offset, req.Action.String(), req.IsSystem, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	permissionResps := make([]dto.PermissionResponse, len(permissions))
	for i, permission := range permissions {
		var permResp dto.PermissionResponse
		permResp.ConvertFromPermission(&permission)
		permissionResps[i] = permResp
	}

	resp := dto.ListResponse[dto.PermissionResponse]{
		Items:      permissionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// SearchPermission performs a search for permissions based on the provided search request
func SearchPermissions(req *dto.SearchRequest) (*dto.SearchResponse[dto.PermissionResponse], error) {
	searchResult, err := repository.ExecuteSearch(database.DB, req, database.Permission{})
	if err != nil {
		return nil, fmt.Errorf("failed to search permissions: %w", err)
	}

	var permissionResponses []dto.PermissionResponse
	for _, permission := range searchResult.Items {
		var response dto.PermissionResponse
		response.ConvertFromPermission(&permission)
		permissionResponses = append(permissionResponses, response)
	}

	resp := dto.SearchResponse[dto.PermissionResponse]{
		Items:      permissionResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}
	return &resp, nil
}

// UpdatePermission updates an existing permission
func UpdatePermission(req *dto.UpdatePermissionRequest, permissionID int) (*dto.PermissionResponse, error) {
	var updatedPermission *database.Permission

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		permission, err := repository.GetPermissionByID(tx, permissionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: permission not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get permission: %w", err)
		}

		if permission.IsSystem {
			return fmt.Errorf("%w: cannot update system permission", consts.ErrPermissionDenied)
		}

		if req.ResourceID != nil {
			resource, err := repository.GetResourceByID(tx, *req.ResourceID)
			if err != nil {
				if errors.Is(err, consts.ErrNotFound) {
					return fmt.Errorf("%w: resource not found", consts.ErrNotFound)
				}
				return fmt.Errorf("failed to get resource: %w", err)
			}
			permission.ResourceID = resource.ID
		}

		req.PatchPermissionModel(permission)

		if err := repository.UpdatePermission(tx, permission); err != nil {
			return fmt.Errorf("failed to update permission: %w", err)
		}

		updatedPermission = permission
		return nil
	})
	if err != nil {
		return nil, err
	}

	var resp dto.PermissionResponse
	resp.ConvertFromPermission(updatedPermission)
	return &resp, nil
}

// ListPermissionRoles lists roles assigned to a specific permission
func ListPermissionRoles(permissionID int) ([]dto.RoleResponse, error) {
	permission, err := repository.GetPermissionByID(database.DB, permissionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: permission not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	roles, err := repository.GetPermissionRoles(database.DB, permission.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission roles: %w", err)
	}

	var roleResps []dto.RoleResponse
	for _, role := range roles {
		var roleResp dto.RoleResponse
		roleResp.ConvertFromRole(&role)
		roleResps = append(roleResps, roleResp)
	}

	return roleResps, nil
}

// actionDisplayName returns the display name for an action
func actionDisplayName(action consts.ActionName) string {
	switch action {
	case "read":
		return "View"
	case "write":
		return "Edit"
	case "delete":
		return "Delete"
	case "execute":
		return "Execute"
	case "manage":
		return "Manage"
	default:
		return action.String()
	}
}

func getPermissionName(action consts.ActionName, resourceName string) string {
	return fmt.Sprintf("%s_%s", action, resourceName)
}

func getPermissionDisplayName(action consts.ActionName, resourceDisplayName string) string {
	return fmt.Sprintf("%s %s", actionDisplayName(action), resourceDisplayName)
}
