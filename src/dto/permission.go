package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// CreatePermissionRequest represents permission creation request
type CreatePermissionRequest struct {
	DisplayName string            `json:"display_name" binding:"omitempty"`
	Description string            `json:"description" binding:"omitempty"`
	Action      consts.ActionName `json:"action" binding:"required"`
	ResourceID  int               `json:"resource_id" binding:"required,min=1"`
}

func (req *CreatePermissionRequest) Validate() error {
	if req.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	if _, ok := consts.ValidActions[consts.ActionName(req.Action)]; !ok {
		return fmt.Errorf("invalid action: %s", req.Action)
	}
	return nil
}

func (req *CreatePermissionRequest) ConvertToPermission() *database.Permission {
	return &database.Permission{
		Description: req.Description,
		Action:      req.Action.String(),
		IsSystem:    false,
		Status:      consts.CommonEnabled,
	}
}

// ListPermissionRequest represents permission list query parameters
type ListPermissionRequest struct {
	PaginationRequest
	Action   consts.ActionName `form:"action" binding:"omitempty"`
	IsSystem *bool             `form:"is_system" binding:"omitempty"`
	Status   *int              `form:"status" binding:"omitempty"`
}

func (req *ListPermissionRequest) Validate() error {
	if req.Action != "" {
		if _, exists := consts.ValidActions[consts.ActionName(req.Action)]; !exists {
			return fmt.Errorf("invalid action: %s", req.Action)
		}
	}
	if req.Status != nil {
		return validateStatusField(req.Status, false)
	}
	return nil
}

// SearchPermissionRequest represents advanced permission search with complex filtering
type SearchPermissionRequest struct {
	AdvancedSearchRequest

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

// ConvertToSearchRequest converts PermissionSearchRequest to SearchRequest with permission-specific filters
func (psr *SearchPermissionRequest) ConvertToSearchRequest() *SearchRequest {
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

// SearchPermissionFilters represents filters for searching permissions
type SearchPermissionFilters struct {
	Name        []string `json:"name,omitempty"`
	DisplayName []string `json:"display_name,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	ResourceIDs []int    `json:"resource_ids,omitempty"`
	Status      []int    `json:"status,omitempty"`
	IsSystem    []bool   `json:"is_system,omitempty"`
}

// UpdatePermissionRequest represents permission update request
type UpdatePermissionRequest struct {
	DisplayName *string            `json:"display_name" binding:"omitempty"`
	Description *string            `json:"description" binding:"omitempty"`
	Action      *consts.ActionName `json:"action" binding:"omitempty"`
	ResourceID  *int               `json:"resource_id" binding:"omitempty,min_ptr=1"`
	Status      *int               `json:"status" binding:"omitempty"`
}

func (req *UpdatePermissionRequest) Validate() error {
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

	if req.Status != nil {
		return validateStatusField(req.Status, true)
	}

	return nil
}

func (req *UpdatePermissionRequest) PatchPermissionModel(target *database.Permission) {
	if req.DisplayName != nil {
		target.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Action != nil {
		target.Action = req.Action.String()
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// PermissionResponse represents permission summary information
type PermissionResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource_name"`
	IsSystem    bool      `json:"is_system"`
	Status      int       `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConvertFromPermission converts database Permission to PermissionResponse DTO
func (resp *PermissionResponse) ConvertFromPermission(permission *database.Permission) {
	resp.ID = permission.ID
	resp.Name = permission.Name
	resp.DisplayName = permission.DisplayName
	resp.Action = permission.Action
	resp.IsSystem = permission.IsSystem
	resp.Status = permission.Status
	resp.UpdatedAt = permission.UpdatedAt

	if permission.Resource != nil {
		resp.Resource = permission.Resource.Name
	}
}

type PermissionDetailResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	IsSystem    bool      `json:"is_system"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Resource *ResourceResponse `json:"resource,omitempty"`
}

func (resp *PermissionDetailResponse) ConvertFromPermission(permission *database.Permission) {
	resp.ID = permission.ID
	resp.Name = permission.Name
	resp.DisplayName = permission.DisplayName
	resp.Description = permission.Description
	resp.Action = permission.Action
	resp.IsSystem = permission.IsSystem
	resp.Status = permission.Status
	resp.CreatedAt = permission.CreatedAt
	resp.UpdatedAt = permission.UpdatedAt

	if permission.Resource != nil {
		var resourceResp ResourceResponse
		resourceResp.ConvertFromResource(permission.Resource)
		resp.Resource = &resourceResp
	}
}

type ListPermissionResponse struct {
	Items      []PermissionResponse `json:"items"`
	Pagination PaginationInfo       `json:"pagination"`
}

type ResourceResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	IsSystem    bool   `json:"is_system"`
}

func (resp *ResourceResponse) ConvertFromResource(resource *database.Resource) {
	resp.ID = resource.ID
	resp.Name = resource.Name
	resp.DisplayName = resource.DisplayName
	resp.Type = resource.Type
	resp.Category = resource.Category
	resp.IsSystem = resource.IsSystem
}
