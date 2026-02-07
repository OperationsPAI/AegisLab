package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// ===================== Team CRUD DTOs =====================

// CreateTeamReq represents team creation request
type CreateTeamReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"omitempty"`
	IsPublic    *bool  `json:"is_public" binding:"omitempty"`
}

func (req *CreateTeamReq) Validate() error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fmt.Errorf("team name cannot be empty")
	}
	if req.IsPublic == nil {
		defaultPublic := true
		req.IsPublic = &defaultPublic
	}
	return nil
}

func (req *CreateTeamReq) ConvertToTeam() *database.Team {
	return &database.Team{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    *req.IsPublic,
		Status:      consts.CommonEnabled,
	}
}

// ListTeamReq represents team list query parameters
type ListTeamReq struct {
	PaginationReq
	IsPublic *bool              `form:"is_public" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListTeamReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatusField(req.Status, false)
}

// UpdateTeamReq represents team update request
type UpdateTeamReq struct {
	Description *string            `json:"description,omitempty"`
	IsPublic    *bool              `json:"is_public,omitempty"`
	Status      *consts.StatusType `json:"status,omitempty"`
}

func (req *UpdateTeamReq) Validate() error {
	return validateStatusField(req.Status, true)
}

func (req *UpdateTeamReq) PatchTeamModel(target *database.Team) {
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

// TeamResp represents basic team response
type TeamResp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTeamResp(team *database.Team) *TeamResp {
	return &TeamResp{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		IsPublic:    team.IsPublic,
		Status:      consts.GetStatusTypeName(team.Status),
		CreatedAt:   team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
	}
}

// TeamDetailResp represents detailed team response
type TeamDetailResp struct {
	TeamResp

	UserCount    int           `json:"user_count"`
	ProjectCount int           `json:"project_count"`
	Projects     []ProjectResp `json:"projects,omitempty"`
}

func NewTeamDetailResp(team *database.Team) *TeamDetailResp {
	return &TeamDetailResp{
		TeamResp: *NewTeamResp(team),
	}
}

// ===================== Team-User DTOs =====================

// ListTeamMemberReq represents team member list query parameters
type ListTeamMemberReq struct {
	PaginationReq
}

func (req *ListTeamMemberReq) Validate() error {
	return req.PaginationReq.Validate()
}

// AddTeamMemberReq represents request to add a user to team
type AddTeamMemberReq struct {
	Username string `json:"username" binding:"required"`
	RoleID   int    `json:"role_id" binding:"required"`
}

func (req *AddTeamMemberReq) Validate() error {
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if req.RoleID <= 0 {
		return fmt.Errorf("invalid role_id")
	}
	return nil
}

// UpdateTeamMemberRoleReq represents request to update team member's role
type UpdateTeamMemberRoleReq struct {
	RoleID int `json:"role_id" binding:"required"`
}

func (req *UpdateTeamMemberRoleReq) Validate() error {
	if req.RoleID <= 0 {
		return fmt.Errorf("invalid role_id")
	}
	return nil
}

// TeamMemberResp represents team member information
type TeamMemberResp struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
	RoleID   int       `json:"role_id"`
	RoleName string    `json:"role_name"`
	JoinedAt time.Time `json:"joined_at"`
}
