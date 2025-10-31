package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description,omitempty" binding:"omitempty"`
}

// ConvertToRole converts CreateRoleRequest to database Role model
func (req *CreateRoleRequest) ConvertToRole() *database.Role {
	return &database.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		Status:      consts.CommonEnabled,
	}
}

// ListRoleRequest represents role list query parameters
type ListRoleRequest struct {
	PaginationRequest
	IsSystem *bool `form:"is_system" binding:"omitempty"`
	Status   *int  `form:"status" binding:"omitempty"`
}

func (req *ListRoleRequest) Validate() error {
	if req.Status != nil {
		return validateStatusField(req.Status, false)
	}
	return nil
}

// SearchRoleRequest represents advanced role search with complex filtering
type SearchRoleRequest struct {
	AdvancedSearchRequest

	// Role-specific filter shortcuts
	NamePattern        string       `json:"name_pattern" binding:"omitempty"`         // Role name fuzzy match
	DisplayNamePattern string       `json:"display_name_pattern" binding:"omitempty"` // Display name fuzzy match
	DescriptionPattern string       `json:"description_pattern" binding:"omitempty"`  // Description fuzzy match
	IsSystem           *bool        `json:"is_system" binding:"omitempty"`            // Whether system role
	PermissionIDs      []int        `json:"permission_ids" binding:"omitempty"`       // Permission ID filter
	UserCount          *NumberRange `json:"user_count" binding:"omitempty"`           // User count range
}

// ConvertToSearchRequest converts RoleSearchRequest to SearchRequest with role-specific filters
func (rsr *SearchRoleRequest) ConvertToSearchRequest() *SearchRequest {
	sr := rsr.ConvertAdvancedToSearch()

	if rsr.NamePattern != "" {
		sr.AddFilter("name", OpLike, rsr.NamePattern)
	}
	if rsr.DisplayNamePattern != "" {
		sr.AddFilter("display_name", OpLike, rsr.DisplayNamePattern)
	}
	if rsr.DescriptionPattern != "" {
		sr.AddFilter("description", OpLike, rsr.DescriptionPattern)
	}

	if rsr.IsSystem != nil {
		sr.AddFilter("is_system", OpEqual, *rsr.IsSystem)
	}
	if len(rsr.PermissionIDs) > 0 {
		values := make([]string, len(rsr.PermissionIDs))
		for i, v := range rsr.PermissionIDs {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "permission_id",
			Operator: OpIn,
			Values:   values,
		})
	}

	return sr
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty"`
	Description *string `json:"description" binding:"omitempty"`
	Status      *int    `json:"status" binding:"omitempty"`
}

func (req *UpdateRoleRequest) Validate() error {
	if req.DisplayName != nil {
		if *req.DisplayName != "" {
			*req.DisplayName = strings.TrimSpace(*req.DisplayName)
		}
	}
	if req.Status != nil {
		return validateStatusField(req.Status, true)
	}
	return nil
}

func (req *UpdateRoleRequest) PatchRoleModel(target *database.Role) {
	if req.DisplayName != nil {
		target.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// AssignPermissionToRoleRequest represents permission assignment to role request
type AssignPermissionToRoleRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1,non_zero_int_slice"`
}

// RemovePermissionFromRoleRequest represents permission removal from role request
type RemovePermissionFromRoleRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1,non_zero_int_slice"`
}

// RoleResponse represents role response
type RoleResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	IsSystem    bool      `json:"is_system"`
	Status      int       `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConvertFromRole converts database Role to RoleResponse DTO
func (resp *RoleResponse) ConvertFromRole(role *database.Role) {
	resp.ID = role.ID
	resp.Name = role.Name
	resp.DisplayName = role.DisplayName
	resp.IsSystem = role.IsSystem
	resp.Status = role.Status
	resp.UpdatedAt = role.UpdatedAt
}

type RoleDetailResponse struct {
	RoleResponse

	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UserCount   int64     `json:"user_count"`

	Permissions []PermissionResponse `json:"permissions"`
}

func (resp *RoleDetailResponse) ConvertFromRole(role *database.Role) {
	resp.RoleResponse.ConvertFromRole(role)
	resp.Description = role.Description
	resp.CreatedAt = role.CreatedAt
}

// ListRoleResponse represents paginated list of roles
type ListRoleResponse struct {
	Items      []RoleResponse `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}
