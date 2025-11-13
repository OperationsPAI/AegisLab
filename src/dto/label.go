package dto

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/utils"
)

type LabelItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ConvertLabelItemssToConditions converts a slice of LabelItem to a slice of map conditions
func ConvertLabelItemsToConditions(labelItems []LabelItem) []map[string]string {
	if len(labelItems) == 0 {
		return []map[string]string{}
	}

	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, label := range labelItems {
		labelConditions = append(labelConditions, map[string]string{
			"key":   label.Key,
			"value": label.Value,
		})
	}

	return labelConditions
}

// =====================================================================
// Label DTOs
// =====================================================================

// BatchDeleteLabelReq represents the request to batch delete labels
type BatchDeleteLabelReq struct {
	IDs []int `json:"ids" binding:"omitempty"` // List of injection IDs for deletion
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

// CreateLabelReq represents label creation request
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
	if err := validateLabelCategory(&req.Category); err != nil {
		return err
	}
	if err := validateColor(req.Color); err != nil {
		return err
	}
	return nil
}

func (req *CreateLabelReq) ConvertToLabel() *database.Label {
	return &database.Label{
		Key:         req.Key,
		Value:       req.Value,
		Category:    req.Category,
		Description: req.Description,
		Color:       utils.GetStringValue(req.Color, "#1890ff"),
		IsSystem:    false,
		Usage:       consts.DefaultLabelUsage,
	}
}

type ListLabelFilters struct {
	Key      string
	Value    string
	Category *consts.LabelCategory
	IsSystem *bool
	Status   *consts.StatusType
}

type ListLabelReq struct {
	PaginationReq

	Key      string                `form:"key" binding:"omitempty"`
	Value    string                `form:"value" binding:"omitempty"`
	Category *consts.LabelCategory `form:"category" binding:"omitempty"`
	IsSystem *bool                 `form:"is_system" binding:"omitempty"`
	Status   *consts.StatusType    `form:"status" binding:"omitempty"`
}

func (req *ListLabelReq) Validate() error {
	if err := validateKeyAndValue(req.Key, req.Value); err != nil {
		return err
	}
	if err := validateLabelCategory(req.Category); err != nil {
		return err
	}
	return validateStatusField(req.Status, false)
}

type UpdateLabelReq struct {
	Description *string            `json:"description" binding:"omitempty"`
	Color       *string            `json:"color" binding:"omitempty"`
	Status      *consts.StatusType `json:"status,omitempty"`
}

func (req *UpdateLabelReq) Validate() error {
	if err := validateColor(req.Color); err != nil {
		return err
	}
	return validateStatusField(req.Status, true)
}

func (req *UpdateLabelReq) PatchLabelModel(target *database.Label) {
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

func (req *ListLabelReq) ToFilterOptions() *ListLabelFilters {
	return &ListLabelFilters{
		Key:      req.Key,
		Value:    req.Value,
		Category: req.Category,
		IsSystem: req.IsSystem,
		Status:   req.Status,
	}
}

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

func NewLabelResp(label *database.Label) *LabelResp {
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

type LabelDetailResp struct {
	LabelResp

	Description string `json:"description"`
}

func NewLabelDetailResp(label *database.Label) *LabelDetailResp {
	return &LabelDetailResp{
		LabelResp:   *NewLabelResp(label),
		Description: label.Description,
	}
}

// =====================================================================
// Validation Helpers
// =====================================================================

// validateColor checks if the provided color is a valid hex color
func validateColor(color *string) error {
	if color == nil {
		return nil
	}
	if !utils.IsValidHexColor(*color) {
		return fmt.Errorf("invalid color format: %s", *color)
	}
	return nil
}

// validateKeyAndValue checks if the label key and value are not empty
func validateKeyAndValue(key, value string) error {
	if key == "" {
		return fmt.Errorf("label key cannot be empty")
	}
	if value == "" {
		return fmt.Errorf("label value cannot be empty")
	}
	return nil
}

// validateLabelCategory validates if the provided category is valid
func validateLabelCategory(category *consts.LabelCategory) error {
	if category != nil {
		if _, exists := consts.ValidLabelCategories[*category]; !exists {
			return fmt.Errorf("invalid label category: %d", category)
		}
		return nil
	}
	return nil
}
