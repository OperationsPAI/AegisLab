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

// CheckUserPermission checks if user has specific permission using a params struct
func CheckUserPermission(params *dto.CheckPermissionParams) (bool, error) {
	if err := params.Validate(); err != nil {
		return false, fmt.Errorf("invalid request: %w", err)
	}

	permission, err := repository.GetPermissionByActionAndResource(database.DB, params.Action, params.Scope, params.ResourceName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to find target permission: %w", err)
	}

	return repository.CheckUserHasPermission(database.DB, params, permission.ID)
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
