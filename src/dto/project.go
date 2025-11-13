package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// ===================== Project CRUD DTOs =====================

// CreateProjectReq represents project creation request
type CreateProjectReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"omitempty"`
	IsPublic    *bool  `json:"is_public" binding:"omitempty"`
}

func (req *CreateProjectReq) Validate() error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if req.IsPublic == nil {
		defaultPublic := true
		req.IsPublic = &defaultPublic
	}
	return nil
}

func (req *CreateProjectReq) ConvertToProject() *database.Project {
	return &database.Project{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    *req.IsPublic,
		Status:      consts.CommonEnabled,
	}
}

// ListProjectReq represents project list query parameters
type ListProjectReq struct {
	PaginationReq
	IsPublic *bool              `form:"is_public" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListProjectReq) Validate() error {
	return validateStatusField(req.Status, false)
}

// SearchProjectReq represents advanced project search
type SearchProjectReq struct {
	AdvancedSearchReq

	NamePattern        string `json:"name_pattern,omitempty"`
	DescriptionPattern string `json:"description_pattern,omitempty"`
	IsPublic           *bool  `json:"is_public,omitempty"`
}

func (req *SearchProjectReq) ConvertToSearchRequest() *SearchReq {
	sr := req.ConvertAdvancedToSearch()

	if req.NamePattern != "" {
		sr.AddFilter("name", OpLike, req.NamePattern)
	}
	if req.DescriptionPattern != "" {
		sr.AddFilter("description", OpLike, req.DescriptionPattern)
	}
	if req.IsPublic != nil {
		sr.AddFilter("is_public", OpEqual, *req.IsPublic)
	}

	return sr
}

// UpdateProjectReq represents project update request
type UpdateProjectReq struct {
	Description *string            `json:"description,omitempty"`
	IsPublic    *bool              `json:"is_public,omitempty"`
	Status      *consts.StatusType `json:"status,omitempty"`
}

func (req *UpdateProjectReq) Validate() error {
	return validateStatusField(req.Status, true)
}

func (req *UpdateProjectReq) PatchProjectModel(target *database.Project) {
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.IsPublic != nil {
		target.IsPublic = *req.IsPublic
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// ProjectResp represents basic project response
type ProjectResp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Labels []LabelItem `json:"labels,omitempty"`
}

func NewProjectResp(project *database.Project) *ProjectResp {
	resp := &ProjectResp{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		IsPublic:    project.IsPublic,
		Status:      consts.GetStatusTypeName(project.Status),
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	if project.Labels != nil {
		resp.Labels = make([]LabelItem, len(project.Labels))
		for i, label := range project.Labels {
			resp.Labels[i] = LabelItem{
				Key:   label.Key,
				Value: label.Value,
			}
		}
	}
	return resp
}

// ProjectDetailResp represents detailed project response
type ProjectDetailResp struct {
	ProjectResp

	Containers []ContainerResp `json:"containers,omitempty"`
	Datapacks  []InjectionResp `json:"datapacks,omitempty"`
	Datasets   []DatasetResp   `json:"datasets,omitempty"`
	UserCount  int             `json:"user_count"`
}

func NewProjectDetailResp(project *database.Project) *ProjectDetailResp {
	return &ProjectDetailResp{
		ProjectResp: *NewProjectResp(project),
	}
}

// ===================== Project-Label DTOs =====================

// ManageProjectLabelReq represents project label management request
type ManageProjectLabelReq struct {
	AddLabels    []LabelItem `json:"add_labels" binding:"omitempty"`    // List of labels to add
	RemoveLabels []string    `json:"remove_labels" binding:"omitempty"` // List of label keys to remove
}

func (req *ManageProjectLabelReq) Validate() error {
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		return fmt.Errorf("at least one of add_labels or remove_labels must be provided")
	}

	for i, label := range req.AddLabels {
		if strings.TrimSpace(label.Key) == "" {
			return fmt.Errorf("empty label key at index %d in add_labels", i)
		}
		if strings.TrimSpace(label.Value) == "" {
			return fmt.Errorf("empty label value at index %d in add_labels", i)
		}
	}

	for i, key := range req.RemoveLabels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty label key at index %d in remove_labels", i)
		}
	}

	return nil
}
