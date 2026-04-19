package dto

import (
	"aegis/consts"
	"fmt"
)

// PaginationInfo represents pagination information in responses
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	Size       int   `json:"size" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}

// PaginationReq represents pagination parameters in requests
type PaginationReq struct {
	Page int             `form:"page" json:"page" example:"1"`
	Size consts.PageSize `form:"size" json:"size" example:"20"`
}

func (p *PaginationReq) Validate() error {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Size == 0 {
		p.Size = consts.PageSizeMedium
	}

	if p.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}
	if _, exists := consts.ValidPageSizes[p.Size]; !exists {
		return fmt.Errorf("invalid page size: %d", p.Size)
	}
	return nil
}

// ToGormParams converts pagination request to limit and offset for GORM queries
func (p *PaginationReq) ToGormParams() (limit int, offset int) {
	limit = int(p.Size)
	if limit == 0 {
		limit = 20
	}

	page := max(p.Page, 1)
	offset = (page - 1) * limit

	return limit, offset
}

func (p *PaginationReq) ConvertToPaginationInfo(total int64) *PaginationInfo {
	totalPages := 0
	if p.Size != 0 {
		totalPages = int((total + int64(p.Size) - 1) / int64(p.Size))
	}
	return &PaginationInfo{
		Page:       p.Page,
		Size:       int(p.Size),
		Total:      total,
		TotalPages: totalPages,
	}
}

// SortField represents a single sort field
type SortField struct {
	Field string `json:"field" binding:"required" example:"created_at"`
	Order string `json:"order" binding:"required,oneof=asc desc" example:"desc"`
}
