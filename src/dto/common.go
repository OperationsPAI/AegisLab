package dto

import (
	"aegis/consts"
	"fmt"
	"strings"
	"time"
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

// validateLabelItemsFiled validates a list of LabelItem structs
func validateLabelItemsFiled(labelItems []LabelItem) error {
	for i, label := range labelItems {
		if strings.TrimSpace(label.Key) == "" {
			return fmt.Errorf("empty label key at index %d", i)
		}
		if strings.TrimSpace(label.Value) == "" {
			return fmt.Errorf("empty label value at index %d", i)
		}
	}
	return nil
}

// validateLabelField validates a list of label strings in "key:value" format
func validateLabelsField(labelStrs []string) error {
	for _, labelStr := range labelStrs {
		if strings.TrimSpace(labelStr) == "" {
			return fmt.Errorf("labels must not contain empty strings")
		}

		parts := strings.SplitN(labelStr, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format '%s'. Must be in 'key:value' format", labelStr)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return fmt.Errorf("label key in '%s' cannot be empty", labelStr)
		}

		if value == "" {
			return fmt.Errorf("label value for key '%s' cannot be empty", key)
		}
	}

	return nil
}

// validateStatusField validates a status field pointer
func validateStatusField(statusPtr *consts.StatusType, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}

	status := *statusPtr

	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d.", status)
	}

	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}

	return nil
}

// validateTimeField checks if the provided time string is in the specific format
func validateTimeField(timeStr, timeFormat string) error {
	if timeStr == "" {
		return nil
	}
	_, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		return fmt.Errorf("invalid time format: %s", timeStr)
	}
	return nil
}
