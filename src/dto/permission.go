package dto

import (
	"fmt"
	"time"

	"rcabench/database"
)

// CreatePermissionRequest represents permission creation request
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required" example:"read_datasets"`
	DisplayName string `json:"display_name" binding:"required" example:"Read Datasets"`
	Description string `json:"description,omitempty" example:"Permission to read dataset information"`
	Action      string `json:"action" binding:"required" example:"read"`
	ResourceID  int    `json:"resource_id" binding:"required" example:"1"`
}

// UpdatePermissionRequest represents permission update request
type UpdatePermissionRequest struct {
	DisplayName string `json:"display_name,omitempty" example:"Read All Datasets"`
	Description string `json:"description,omitempty" example:"Updated permission description"`
	Action      string `json:"action,omitempty" example:"read"`
	ResourceID  *int   `json:"resource_id,omitempty" example:"2"`
	Status      *int   `json:"status,omitempty" example:"1"`
}

// PermissionResponse represents permission response
type PermissionResponse struct {
	ID          int               `json:"id" example:"1"`
	Name        string            `json:"name" example:"read_datasets"`
	DisplayName string            `json:"display_name" example:"Read Datasets"`
	Description string            `json:"description" example:"Permission to read dataset information"`
	Action      string            `json:"action" example:"read"`
	ResourceID  int               `json:"resource_id" example:"1"`
	IsSystem    bool              `json:"is_system" example:"false"`
	Status      int               `json:"status" example:"1"`
	CreatedAt   time.Time         `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Resource    *ResourceResponse `json:"resource,omitempty"`
	Roles       []RoleResponse    `json:"roles,omitempty"` // Roles that have this permission
}

// PermissionListRequest represents permission list query parameters
type PermissionListRequest struct {
	Page       int    `form:"page,default=1" binding:"min=1" example:"1"`
	Size       int    `form:"size,default=20" binding:"min=1,max=100" example:"20"`
	Action     string `form:"action" example:"read"`
	ResourceID *int   `form:"resource_id" example:"1"`
	Status     *int   `form:"status" example:"1"`
	IsSystem   *bool  `form:"is_system" example:"false"`
	Name       string `form:"name" example:"read_datasets"`
}

// PermissionListResponse represents paginated permission list response
type PermissionListResponse struct {
	Items      []PermissionResponse `json:"items"`
	Pagination PaginationInfo       `json:"pagination"`
}

// PermissionSearchRequest represents advanced permission search with complex filtering
type PermissionSearchRequest struct {
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
func (psr *PermissionSearchRequest) ConvertToSearchRequest() *SearchRequest {
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

// PermissionSearchFilters represents simple search filters for backward compatibility
type PermissionSearchFilters struct {
	Name        []string `json:"name,omitempty"`
	DisplayName []string `json:"display_name,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	ResourceIDs []int    `json:"resource_ids,omitempty"`
	Status      []int    `json:"status,omitempty"`
	IsSystem    []bool   `json:"is_system,omitempty"`
}

// ResourceResponse represents resource response (simplified)
type ResourceResponse struct {
	ID          int    `json:"id" example:"1"`
	Name        string `json:"name" example:"datasets"`
	DisplayName string `json:"display_name" example:"Datasets"`
	Type        string `json:"type" example:"table"`
	Category    string `json:"category" example:"data"`
}

// ConvertFromPermission converts database Permission to PermissionResponse DTO
func (p *PermissionResponse) ConvertFromPermission(permission *database.Permission) {
	p.ID = permission.ID
	p.Name = permission.Name
	p.DisplayName = permission.DisplayName
	p.Description = permission.Description
	p.Action = permission.Action
	p.ResourceID = permission.ResourceID
	p.IsSystem = permission.IsSystem
	p.Status = permission.Status
	p.CreatedAt = permission.CreatedAt
	p.UpdatedAt = permission.UpdatedAt

	if permission.Resource != nil {
		p.Resource = &ResourceResponse{
			ID:          permission.Resource.ID,
			Name:        permission.Resource.Name,
			DisplayName: permission.Resource.DisplayName,
			Type:        permission.Resource.Type,
			Category:    permission.Resource.Category,
		}
	}
}
