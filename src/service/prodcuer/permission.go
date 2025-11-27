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

// CreatePermission creates a new permission
func CreatePermission(req *dto.CreatePermissionReq) (*dto.PermissionResp, error) {
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

		permission.Name = GetPermissionName(req.Action, resource.Name)
		permission.ResourceID = resource.ID
		if req.DisplayName != "" {
			permission.DisplayName = req.DisplayName
		} else {
			permission.DisplayName = GetPermissionDisplayName(req.Action, resource.DisplayName)
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

	return dto.NewPermissionResp(createdPermission), nil
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

		if err := repository.RemoveRolesFromPermission(tx, permissionID); err != nil {
			return fmt.Errorf("failed to remove roles from permission: %w", err)
		}
		if err := repository.RemoveRolesFromUser(tx, permissionID); err != nil {
			return fmt.Errorf("failed to remove users from permission: %w", err)
		}

		row, err := repository.DeletePermission(tx, permissionID)
		if err != nil {
			return fmt.Errorf("failed to delete permission: %w", err)
		}
		if row == 0 {
			return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, permissionID)
		}

		return nil
	})
}

// GetPermissionDetail retrieves detailed information about a permission by its ID
func GetPermissionDetail(permissionID int) (*dto.PermissionDetailResp, error) {
	permission, err := repository.GetPermissionByID(database.DB, permissionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: permission not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return dto.NewPermissionDetailResp(permission), nil
}

// GetPermissionName generates the permission name based on action and resource name
func GetPermissionName(action consts.ActionName, resourceName consts.ResourceName) string {
	return fmt.Sprintf("%s_%s", action, resourceName.String())
}

// GetPermissionDisplayName generates the permission display name based on action and resource display name
func GetPermissionDisplayName(action consts.ActionName, resourceDisplayName string) string {
	return fmt.Sprintf("%s %s", actionDisplayName(action), resourceDisplayName)
}

// ListPermissions lists permissions based on the provided request parameters
func ListPermissions(req *dto.ListPermissionReq) (*dto.ListResp[dto.PermissionResp], error) {
	limit, offset := req.ToGormParams()

	permissions, total, err := repository.ListPermissions(database.DB, limit, offset, req.Action, req.IsSystem, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	permissionResps := make([]dto.PermissionResp, len(permissions))
	for i, permission := range permissions {
		permissionResps[i] = *dto.NewPermissionResp(&permission)
	}

	resp := dto.ListResp[dto.PermissionResp]{
		Items:      permissionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdatePermission updates an existing permission
func UpdatePermission(req *dto.UpdatePermissionReq, permissionID int) (*dto.PermissionResp, error) {
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

	return dto.NewPermissionResp(updatedPermission), nil
}

func ListRolesFromPermission(permissionID int) ([]dto.RoleResp, error) {
	permission, err := repository.GetPermissionByID(database.DB, permissionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: permission not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	roles, err := repository.ListRolesByPermissionID(database.DB, permission.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission roles: %w", err)
	}

	var roleResps []dto.RoleResp
	for _, role := range roles {
		roleResps = append(roleResps, *dto.NewRoleResp(&role))
	}

	return roleResps, nil
}

// fetchPermissionsMapByIDBatch fetches permissions by their IDs and returns a map of permission ID to Permission
func fetchPermissionsMapByIDBatch(db *gorm.DB, permissionIDs []int) (map[int]database.Permission, error) {
	if len(permissionIDs) == 0 {
		return make(map[int]database.Permission), nil
	}

	uniqueIDs := make(map[int]struct{})
	for _, id := range permissionIDs {
		uniqueIDs[id] = struct{}{}
	}

	deduplicatedIDs := make([]int, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		deduplicatedIDs = append(deduplicatedIDs, id)
	}

	permissions, err := repository.ListPermissionsByID(db, deduplicatedIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by IDs: %w", err)
	}

	permissionMap := make(map[int]database.Permission, len(permissions))
	for _, perm := range permissions {
		permissionMap[perm.ID] = perm
	}

	return permissionMap, nil
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
		return string(action)
	}
}
