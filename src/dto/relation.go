package dto

import (
	"aegis/database"
	"time"
)

// AssignRolePermissionRequest represents role-permission assignment request
type AssignRolePermissionRequest struct {
	RoleID        int   `json:"role_id" binding:"required" example:"1"`
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1" example:"[1,2,3]"`
}

// RemoveRolePermissionRequest represents role-permission removal request
type RemoveRolePermissionRequest struct {
	RoleID        int   `json:"role_id" binding:"required" example:"1"`
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1" example:"[1,2]"`
}

// AssignUserPermissionRequest represents direct user-permission assignment request
type AssignUserPermissionRequest struct {
	PermissionID int        `json:"permission_id" binding:"required"`
	ProjectID    *int       `json:"project_id,omitempty"`
	ContainerID  *int       `json:"container_id,omitempty"`
	GrantType    string     `json:"grant_type" binding:"required,oneof=grant deny"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

func (req *AssignUserPermissionRequest) ConvertToUserPermission() *database.UserPermission {
	return &database.UserPermission{
		GrantType: req.GrantType,
		ExpiresAt: req.ExpiresAt,
	}
}

// RemoveUserPermissionRequest represents direct user-permission removal request
type RemoveUserPermissionRequest struct {
	UserID       int  `json:"user_id" binding:"required" example:"1"`
	PermissionID int  `json:"permission_id" binding:"required" example:"1"`
	ProjectID    *int `json:"project_id,omitempty" example:"1"`
}

// AssignDatasetLabelRequest represents dataset-label assignment request
type AssignDatasetLabelRequest struct {
	DatasetID int   `json:"dataset_id" binding:"required" example:"1"`
	LabelIDs  []int `json:"label_ids" binding:"required,min=1" example:"[1,2,3]"`
}

// RemoveDatasetLabelRequest represents dataset-label removal request
type RemoveDatasetLabelRequest struct {
	DatasetID int   `json:"dataset_id" binding:"required" example:"1"`
	LabelIDs  []int `json:"label_ids" binding:"required,min=1" example:"[1,2]"`
}

// AssignContainerLabelRequest represents container-label assignment request
type AssignContainerLabelRequest struct {
	ContainerID int   `json:"container_id" binding:"required" example:"1"`
	LabelIDs    []int `json:"label_ids" binding:"required,min=1" example:"[1,2,3]"`
}

// RemoveContainerLabelRequest represents container-label removal request
type RemoveContainerLabelRequest struct {
	ContainerID int   `json:"container_id" binding:"required" example:"1"`
	LabelIDs    []int `json:"label_ids" binding:"required,min=1" example:"[1,2]"`
}

// AssignProjectLabelRequest represents project-label assignment request
type AssignProjectLabelRequest struct {
	ProjectID int   `json:"project_id" binding:"required" example:"1"`
	LabelIDs  []int `json:"label_ids" binding:"required,min=1" example:"[1,2,3]"`
}

// RemoveProjectLabelRequest represents project-label removal request
type RemoveProjectLabelRequest struct {
	ProjectID int   `json:"project_id" binding:"required" example:"1"`
	LabelIDs  []int `json:"label_ids" binding:"required,min=1" example:"[1,2]"`
}

// AssignFaultInjectionLabelRequest represents fault injection-label assignment request
type AssignFaultInjectionLabelRequest struct {
	FaultInjectionID int   `json:"fault_injection_id" binding:"required" example:"1"`
	LabelIDs         []int `json:"label_ids" binding:"required,min=1" example:"[1,2,3]"`
}

// RemoveFaultInjectionLabelRequest represents fault injection-label removal request
type RemoveFaultInjectionLabelRequest struct {
	FaultInjectionID int   `json:"fault_injection_id" binding:"required" example:"1"`
	LabelIDs         []int `json:"label_ids" binding:"required,min=1" example:"[1,2]"`
}

// RelationResponse represents a generic relationship response
type RelationResponse struct {
	ID        int            `json:"id" example:"1"`
	Type      string         `json:"type" example:"user_role"`
	Source    RelationEntity `json:"source"`
	Target    RelationEntity `json:"target"`
	CreatedAt time.Time      `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// RelationEntity represents an entity in a relationship
type RelationEntity struct {
	ID   int    `json:"id" example:"1"`
	Type string `json:"type" example:"user"`
	Name string `json:"name" example:"admin"`
}

// RelationListRequest represents relation list query parameters
type RelationListRequest struct {
	Page       int    `form:"page,default=1" binding:"min=1" example:"1"`
	Size       int    `form:"size,default=20" binding:"min=1,max=100" example:"20"`
	Type       string `form:"type" example:"user_role"`
	SourceType string `form:"source_type" example:"user"`
	TargetType string `form:"target_type" example:"role"`
	SourceID   *int   `form:"source_id" example:"1"`
	TargetID   *int   `form:"target_id" example:"2"`
}

// RelationListResponse represents paginated relation list response
type RelationListResponse struct {
	Items      []RelationResponse `json:"items"`
	Pagination PaginationInfo     `json:"pagination"`
}

// BatchRelationRequest represents batch relationship operations
type BatchRelationRequest struct {
	Operations []RelationOperation `json:"operations" binding:"required,min=1"`
}

// RelationOperation represents a single relationship operation
type RelationOperation struct {
	Action   string `json:"action" binding:"required,oneof=assign remove"`
	Type     string `json:"type" binding:"required"`
	SourceID int    `json:"source_id" binding:"required"`
	TargetID int    `json:"target_id" binding:"required"`
}

// RelationStatisticsResponse represents relationship statistics
type RelationStatisticsResponse struct {
	UserRoles            int `json:"user_roles"`
	RolePermissions      int `json:"role_permissions"`
	UserPermissions      int `json:"user_permissions"`
	UserProjects         int `json:"user_projects"`
	DatasetLabels        int `json:"dataset_labels"`
	ContainerLabels      int `json:"container_labels"`
	ProjectLabels        int `json:"project_labels"`
	FaultInjectionLabels int `json:"fault_injection_labels"`
}
