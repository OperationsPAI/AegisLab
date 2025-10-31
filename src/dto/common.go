package dto

import (
	"aegis/consts"
	"fmt"
)

// PaginationRequest represents pagination parameters in requests
type PaginationRequest struct {
	Page int `form:"page,default=1" binding:"omitempty,min=1"`
	Size int `form:"size,default=20" binding:"omitempty,min=1,max=100"`
}

func (p *PaginationRequest) ToGormParams() (limit int, offset int) {
	limit = p.Size
	if limit == 0 {
		limit = 20
	}

	page := p.Page
	if page < 1 {
		page = 1
	}
	offset = (page - 1) * limit

	return limit, offset
}

func (p *PaginationRequest) ConvertToPaginationInfo(total int64) *PaginationInfo {
	totalPages := int((total + int64(p.Size) - 1) / int64(p.Size))
	return &PaginationInfo{
		Page:       p.Page,
		Size:       p.Size,
		Total:      total,
		TotalPages: totalPages,
	}
}

// SortField represents a single sort field
type SortField struct {
	Field string `json:"field" binding:"required" example:"created_at"`
	Order string `json:"order" binding:"required,oneof=asc desc" example:"desc"`
}

// validateStatusField validates a status field pointer
func validateStatusField(statusPtr *int, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}

	status := *statusPtr

	if _, exists := consts.ValidCommonStatus[status]; !exists {
		validKeys := consts.GetValidStatusKeys()
		return fmt.Errorf("invalid status value: %d. Status must be one of %v", status, validKeys)
	}

	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}

	return nil
}
