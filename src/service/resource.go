package service

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"
)

func ListResourcePermissions(resourceID int) ([]dto.PermissionResponse, error) {
	resource, err := repository.GetResourceByID(database.DB, resourceID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: resource with ID %d not found", consts.ErrNotFound, resourceID)
		}
	}

	permissions, err := repository.GetPermissionsByResource(database.DB, resource.ID)
	if err != nil {
		return nil, err
	}

	var permissionResps []dto.PermissionResponse
	for _, permission := range permissions {
		var response dto.PermissionResponse
		response.ConvertFromPermission(&permission)
		permissionResps = append(permissionResps, response)
	}

	return permissionResps, nil
}
