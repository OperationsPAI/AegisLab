package projectmodule

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	containermodule "aegis/module/container"
	datasetmodule "aegis/module/dataset"
	injectionmodule "aegis/module/injection"
)

type ProjectContainerItem = containermodule.ContainerResp
type ProjectDatasetItem = datasetmodule.DatasetResp

// CreateProjectReq represents project creation request.
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

func (req *CreateProjectReq) ConvertToProject() *model.Project {
	return &model.Project{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    *req.IsPublic,
		Status:      consts.CommonEnabled,
	}
}

// ListProjectReq represents project list query parameters.
type ListProjectReq struct {
	dto.PaginationReq
	IsPublic          *bool              `form:"is_public" binding:"omitempty"`
	Status            *consts.StatusType `form:"status" binding:"omitempty"`
	TeamID            *int               `form:"team_id" binding:"omitempty"`
	IncludeStatistics *bool              `form:"include_statistics" binding:"omitempty"`
}

func (req *ListProjectReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if req.TeamID != nil && *req.TeamID <= 0 {
		return fmt.Errorf("team_id must be greater than 0")
	}
	return validateStatus(req.Status, false)
}

// UpdateProjectReq represents project update request.
type UpdateProjectReq struct {
	Description *string            `json:"description,omitempty"`
	IsPublic    *bool              `json:"is_public,omitempty"`
	Status      *consts.StatusType `json:"status,omitempty"`
}

func (req *UpdateProjectReq) Validate() error {
	return validateStatus(req.Status, true)
}

func (req *UpdateProjectReq) PatchProjectModel(target *model.Project) {
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

// ManageProjectLabelReq represents project label management request.
type ManageProjectLabelReq struct {
	AddLabels    []dto.LabelItem `json:"add_labels" binding:"omitempty"`
	RemoveLabels []string        `json:"remove_labels" binding:"omitempty"`
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

// ProjectResp represents basic project response.
type ProjectResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsPublic  bool      `json:"is_public"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	LastInjectionAt *time.Time      `json:"last_injection_at,omitempty"`
	LastExecutionAt *time.Time      `json:"last_execution_at,omitempty"`
	InjectionCount  int             `json:"injection_count"`
	ExecutionCount  int             `json:"execution_count"`
	Labels          []dto.LabelItem `json:"labels,omitempty"`
}

func NewProjectResp(project *model.Project, stats *dto.ProjectStatistics) *ProjectResp {
	resp := &ProjectResp{
		ID:        project.ID,
		Name:      project.Name,
		IsPublic:  project.IsPublic,
		Status:    consts.GetStatusTypeName(project.Status),
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}

	if stats != nil {
		resp.LastInjectionAt = stats.LastInjectionAt
		resp.LastExecutionAt = stats.LastExecutionAt
		resp.InjectionCount = stats.InjectionCount
		resp.ExecutionCount = stats.ExecutionCount
	}

	if project.Labels != nil {
		resp.Labels = make([]dto.LabelItem, len(project.Labels))
		for i, label := range project.Labels {
			resp.Labels[i] = dto.LabelItem{Key: label.Key, Value: label.Value}
		}
	}
	return resp
}

// ProjectDetailResp represents detailed project response.
type ProjectDetailResp struct {
	ProjectResp

	Containers []ProjectContainerItem          `json:"containers,omitempty"`
	Datapacks  []injectionmodule.InjectionResp `json:"datapacks,omitempty"`
	Datasets   []ProjectDatasetItem            `json:"datasets,omitempty"`
	UserCount  int                             `json:"user_count"`
}

func NewProjectDetailResp(project *model.Project, stats *dto.ProjectStatistics) *ProjectDetailResp {
	return &ProjectDetailResp{
		ProjectResp: *NewProjectResp(project, stats),
	}
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
