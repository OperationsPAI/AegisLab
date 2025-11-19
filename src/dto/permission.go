package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// CreatePermissionReq represents permission creation request
type CreatePermissionReq struct {
	DisplayName string            `json:"display_name" binding:"omitempty"`
	Description string            `json:"description" binding:"omitempty"`
	Action      consts.ActionName `json:"action" binding:"required"`
	ResourceID  int               `json:"resource_id" binding:"required,min=1"`
}

func (req *CreatePermissionReq) Validate() error {
	if req.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	if _, ok := consts.ValidActions[consts.ActionName(req.Action)]; !ok {
		return fmt.Errorf("invalid action: %s", req.Action)
	}
	return nil
}

func (req *CreatePermissionReq) ConvertToPermission() *database.Permission {
	return &database.Permission{
		Description: req.Description,
		Action:      string(req.Action),
		IsSystem:    false,
		Status:      consts.CommonEnabled,
	}
}

// ListPermissionReq represents permission list query parameters
type ListPermissionReq struct {
	PaginationReq
	Action   consts.ActionName  `form:"action" binding:"omitempty"`
	IsSystem *bool              `form:"is_system" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListPermissionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if _, exists := consts.ValidActions[req.Action]; !exists {
		return fmt.Errorf("invalid action: %s", req.Action)
	}
	if req.Status != nil {
		return validateStatusField(req.Status, false)
	}
	return nil
}

// SearchPermissionReq represents advanced permission search with complex filtering
type SearchPermissionReq struct {
	AdvancedSearchReq

	// Permission-specific filter shortcuts
	NamePattern        string   `json:"name_pattern,omitempty"`         // Fuzzy match for permission name
	DisplayNamePattern string   `json:"display_name_pattern,omitempty"` // Fuzzy match for display name
	DescriptionPattern string   `json:"description_pattern,omitempty"`  // Fuzzy match for description
	Actions            []string `json:"actions,omitempty"`              // Action filter
	ResourceIDs        []int    `json:"resource_ids,omitempty"`         // Resource ID filter
	ResourceNames      []string `json:"resource_names,omitempty"`       // Resource name filter
	IsSystem           *bool    `json:"is_system,omitempty"`            // Is system permission
	RoleIDs            []int    `json:"role_ids,omitempty"`             // Role IDs that have this permission
}

// ConvertToSearchRequest converts PermissionSearchReq to SearchRequest with permission-specific filters
func (psr *SearchPermissionReq) ConvertToSearchRequest() *SearchReq {
	sr := psr.ConvertAdvancedToSearch()

	// Add permission-specific filters
	if psr.NamePattern != "" {
		sr.AddFilter("name", OpLike, psr.NamePattern)
	}

	if psr.DisplayNamePattern != "" {
		sr.AddFilter("display_name", OpLike, psr.DisplayNamePattern)
	}

	if psr.DescriptionPattern != "" {
		sr.AddFilter("description", OpLike, psr.DescriptionPattern)
	}

	if len(psr.Actions) > 0 {
		values := make([]string, len(psr.Actions))
		for i, v := range psr.Actions {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "action",
			Operator: OpIn,
			Values:   values,
		})
	}

	if len(psr.ResourceIDs) > 0 {
		values := make([]string, len(psr.ResourceIDs))
		for i, v := range psr.ResourceIDs {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "resource_id",
			Operator: OpIn,
			Values:   values,
		})
	}

	if len(psr.ResourceNames) > 0 {
		values := make([]string, len(psr.ResourceNames))
		for i, v := range psr.ResourceNames {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "resource_name",
			Operator: OpIn,
			Values:   values,
		})
	}

	if psr.IsSystem != nil {
		sr.AddFilter("is_system", OpEqual, *psr.IsSystem)
	}

	if len(psr.RoleIDs) > 0 {
		values := make([]string, len(psr.RoleIDs))
		for i, v := range psr.RoleIDs {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "role_id",
			Operator: OpIn,
			Values:   values,
		})
	}

	return sr
}

// UpdatePermissionReq represents permission update request
type UpdatePermissionReq struct {
	DisplayName *string            `json:"display_name" binding:"omitempty"`
	Description *string            `json:"description" binding:"omitempty"`
	Action      *consts.ActionName `json:"action" binding:"omitempty"`
	ResourceID  *int               `json:"resource_id" binding:"omitempty,min_ptr=1"`
	Status      *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdatePermissionReq) Validate() error {
	if req.DisplayName != nil {
		if *req.DisplayName != "" {
			*req.DisplayName = strings.TrimSpace(*req.DisplayName)
		}
	}

	if req.Action != nil {
		if *req.Action == "" {
			return fmt.Errorf("action cannot be empty")
		}
		if _, ok := consts.ValidActions[consts.ActionName(*req.Action)]; !ok {
			return fmt.Errorf("invalid action: %s", *req.Action)
		}
	}

	return validateStatusField(req.Status, true)
}

func (req *UpdatePermissionReq) PatchPermissionModel(target *database.Permission) {
	if req.DisplayName != nil {
		target.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Action != nil {
		target.Action = string(*req.Action)
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// PermissionResp represents permission summary information
type PermissionResp struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name"`
	Action      string              `json:"action"`
	Resource    consts.ResourceName `json:"resource_name"`
	IsSystem    bool                `json:"is_system"`
	Status      string              `json:"status"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func NewPermissionResp(perm *database.Permission) *PermissionResp {
	resp := &PermissionResp{
		ID:          perm.ID,
		Name:        perm.Name,
		DisplayName: perm.DisplayName,
		Action:      perm.Action,
		IsSystem:    perm.IsSystem,
		Status:      consts.GetStatusTypeName(perm.Status),
		UpdatedAt:   perm.UpdatedAt,
	}

	if perm.Resource != nil {
		resp.Resource = perm.Resource.Name
	}
	return resp
}

type PermissionDetailResp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	IsSystem    bool      `json:"is_system"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Resource *ResourceResp `json:"resource,omitempty"`
}

func NewPermissionDetailResp(perm *database.Permission) *PermissionDetailResp {
	resp := &PermissionDetailResp{
		ID:          perm.ID,
		Name:        perm.Name,
		DisplayName: perm.DisplayName,
		Description: perm.Description,
		Action:      perm.Action,
		IsSystem:    perm.IsSystem,
		Status:      consts.GetStatusTypeName(perm.Status),
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}

	if perm.Resource != nil {
		resp.Resource = NewResourceResp(perm.Resource)
	}
	return resp
}
