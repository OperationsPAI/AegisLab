package dto

import (
	"time"

	"aegis/database"
	"aegis/utils"
)

type LabelCreateReq struct {
	Key         string  `json:"key" binding:"required,max=255"`
	Value       string  `json:"value" binding:"required,max=255"`
	Category    string  `json:"category" bindging:"required,max=50"`
	Description string  `json:"description" binding:"max=1000"`
	Color       *string `json:"color" binding:"omitempty,max=10"`
}

func (req *LabelCreateReq) ToEntity() *database.Label {
	return &database.Label{
		Key:         req.Key,
		Value:       req.Value,
		Category:    req.Category,
		Description: req.Description,
		Color:       utils.GetStringValue(req.Color, "#1890ff"),
		IsSystem:    false,
		Usage:       0,
	}
}

type LabelResponse struct {
	ID          int       `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	IsSystem    bool      `json:"is_system"`
	Usage       int       `json:"usage"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToLabelResponse(label *database.Label) *LabelResponse {
	return &LabelResponse{
		ID:          label.ID,
		Key:         label.Key,
		Value:       label.Value,
		Category:    label.Category,
		Description: label.Description,
		Color:       label.Color,
		IsSystem:    label.IsSystem,
		Usage:       label.Usage,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}
}
