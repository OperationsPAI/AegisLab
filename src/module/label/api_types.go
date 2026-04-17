package labelmodule

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"
)

// BatchDeleteLabelReq represents the request to batch delete labels.
type BatchDeleteLabelReq struct {
	IDs []int `json:"ids" binding:"omitempty"`
}

func (req *BatchDeleteLabelReq) Validate() error {
	if len(req.IDs) == 0 {
		return fmt.Errorf("ids cannot be empty")
	}
	for i, id := range req.IDs {
		if id <= 0 {
			return fmt.Errorf("invalid id at index %d: %d", i, id)
		}
	}
	return nil
}

// CreateLabelReq represents label creation request.
type CreateLabelReq struct {
	Key         string               `json:"key" binding:"required"`
	Value       string               `json:"value" binding:"required"`
	Category    consts.LabelCategory `json:"category" bindging:"required"`
	Description string               `json:"description" binding:"omitempty"`
	Color       *string              `json:"color" binding:"omitempty"`
}

func (req *CreateLabelReq) Validate() error {
	if err := validateKeyAndValue(req.Key, req.Value); err != nil {
		return err
	}
	if err := validateLabelCategory(req.Category); err != nil {
		return err
	}
	if err := validateColor(req.Color); err != nil {
		return err
	}
	return nil
}

func (req *CreateLabelReq) ConvertToLabel() *model.Label {
	return &model.Label{
		Key:         req.Key,
		Value:       req.Value,
		Category:    req.Category,
		Description: req.Description,
		Color:       utils.GetStringValue(req.Color, "#1890ff"),
		IsSystem:    false,
		Usage:       consts.DefaultLabelUsage,
	}
}

// ListLabelReq is the list-label query contract for the label module.
type ListLabelReq struct {
	dto.PaginationReq

	Key      string                `form:"key" binding:"omitempty"`
	Value    string                `form:"value" binding:"omitempty"`
	Category *consts.LabelCategory `form:"category" binding:"omitempty"`
	IsSystem *bool                 `form:"is_system" binding:"omitempty"`
	Status   *consts.StatusType    `form:"status" binding:"omitempty"`
}

type ListLabelFilters struct {
	Key      string
	Value    string
	Category *consts.LabelCategory
	IsSystem *bool
	Status   *consts.StatusType
}

func (req *ListLabelReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateKeyAndValue(req.Key, req.Value); err != nil {
		return err
	}
	if req.Category != nil {
		if err := validateLabelCategory(*req.Category); err != nil {
			return err
		}
	}
	return validateStatus(req.Status, false)
}

func (req *ListLabelReq) ToFilterOptions() *ListLabelFilters {
	return &ListLabelFilters{
		Key:      req.Key,
		Value:    req.Value,
		Category: req.Category,
		IsSystem: req.IsSystem,
		Status:   req.Status,
	}
}

// UpdateLabelReq represents label update request.
type UpdateLabelReq struct {
	Description *string            `json:"description" binding:"omitempty"`
	Color       *string            `json:"color" binding:"omitempty"`
	Status      *consts.StatusType `json:"status,omitempty"`
}

func (req *UpdateLabelReq) Validate() error {
	if err := validateColor(req.Color); err != nil {
		return err
	}
	return validateStatus(req.Status, true)
}

func (req *UpdateLabelReq) PatchLabelModel(target *model.Label) {
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Color != nil {
		target.Color = *req.Color
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// LabelResp represents a label response.
type LabelResp struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Category  string    `json:"category"`
	Color     string    `json:"color"`
	Usage     int       `json:"usage"`
	IsSystem  bool      `json:"is_system"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewLabelResp(label *model.Label) *LabelResp {
	return &LabelResp{
		ID:        label.ID,
		Key:       label.Key,
		Value:     label.Value,
		Category:  consts.GetLabelCategoryName(label.Category),
		Color:     label.Color,
		Usage:     label.Usage,
		IsSystem:  label.IsSystem,
		Status:    consts.GetStatusTypeName(label.Status),
		CreatedAt: label.CreatedAt,
		UpdatedAt: label.UpdatedAt,
	}
}

// LabelDetailResp represents a detailed label response.
type LabelDetailResp struct {
	LabelResp

	Description string `json:"description"`
}

func NewLabelDetailResp(label *model.Label) *LabelDetailResp {
	return &LabelDetailResp{
		LabelResp:   *NewLabelResp(label),
		Description: label.Description,
	}
}

func validateColor(color *string) error {
	if color == nil {
		return nil
	}
	if !utils.IsValidHexColor(*color) {
		return fmt.Errorf("invalid color format: %s", *color)
	}
	return nil
}

func validateKeyAndValue(key, value string) error {
	if key == "" && value == "" {
		return nil
	}
	if key == "" {
		return fmt.Errorf("label key cannot be empty when value is provided")
	}
	if value == "" {
		return fmt.Errorf("label value cannot be empty when key is provided")
	}
	return nil
}

func validateLabelCategory(category consts.LabelCategory) error {
	if _, exists := consts.ValidLabelCategories[category]; !exists {
		return fmt.Errorf("invalid label category: %d", category)
	}
	return nil
}

func validateStatus(statusPtr *consts.StatusType, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}
	status := *statusPtr
	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d", status)
	}
	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}
	return nil
}
