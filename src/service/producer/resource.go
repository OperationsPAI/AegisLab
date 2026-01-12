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

// GetResourceDetail retrieves detailed information about a resource by its ID
func GetResourceDetail(resourceID int) (*dto.ResourceResp, error) {
	resource, err := repository.GetResourceByID(database.DB, resourceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: resource with ID %d not found", consts.ErrNotFound, resourceID)
		}
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	return dto.NewResourceResp(resource), nil
}

// ListResources lists resources based on the provided filters
func ListResources(req *dto.ListResourceReq) (*dto.ListResp[dto.ResourceResp], error) {
	limit, offset := req.ToGormParams()

	resources, total, err := repository.ListResources(database.DB, limit, offset, req.Type, req.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	resourceResps := make([]dto.ResourceResp, 0, len(resources))
	for i := range resources {
		resourceResps = append(resourceResps, *dto.NewResourceResp(&resources[i]))
	}

	resp := dto.ListResp[dto.ResourceResp]{
		Items:      resourceResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// ListResourcePermissions lists permissions associated with a specific resource
func ListResourcePermissions(resourceID int) ([]dto.PermissionResp, error) {
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

	var permissionResps []dto.PermissionResp
	for _, permission := range permissions {
		permissionResps = append(permissionResps, *dto.NewPermissionResp(&permission))
	}

	return permissionResps, nil
}
