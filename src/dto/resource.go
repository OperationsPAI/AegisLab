package dto

import (
	"aegis/consts"
	"aegis/database"
	"fmt"
)

// ListResourceReq represents request for listing resources
type ListResourceReq struct {
	PaginationReq

	Type     *consts.ResourceType     `form:"type" binding:"omitempty"`
	Category *consts.ResourceCategory `form:"category" binding:"omitempty"`
}

func (req *ListResourceReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if req.Type != nil {
		if _, exists := consts.ValidResourceTypes[*req.Type]; !exists {
			return fmt.Errorf("invalid resource type: %d", *req.Type)
		}
	}
	if req.Category != nil {
		if _, exists := consts.ValidResourceCategories[*req.Category]; !exists {
			return fmt.Errorf("invalid resource category: %d", *req.Category)
		}
	}
	return nil
}

type ResourceResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	ParentID    *int   `json:"parent_id,omitempty"`
}

func NewResourceResp(resource *database.Resource) *ResourceResp {
	return &ResourceResp{
		ID:          resource.ID,
		Name:        resource.Name.String(),
		DisplayName: resource.DisplayName,
		Type:        consts.GetResourceTypeName(resource.Type),
		Category:    consts.GetResourceCategoryName(resource.Category),
		ParentID:    resource.ParentID,
	}
}

type ResourceDetailResp struct {
	ResourceResp

	Description string `json:"description,omitempty"`
}

func NewResourceDetailResp(resource *database.Resource) *ResourceDetailResp {
	return &ResourceDetailResp{
		ResourceResp: *NewResourceResp(resource),
		Description:  resource.Description,
	}
}
