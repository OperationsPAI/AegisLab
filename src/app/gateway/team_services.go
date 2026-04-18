package gatewayapp

import (
	"context"

	"aegis/dto"
	teammodule "aegis/module/team"
)

type teamIAMClient interface {
	Enabled() bool
	CreateTeam(context.Context, *teammodule.CreateTeamReq, int) (*teammodule.TeamResp, error)
	DeleteTeam(context.Context, int) error
	GetTeam(context.Context, int) (*teammodule.TeamDetailResp, error)
	ListTeams(context.Context, *teammodule.ListTeamReq, int, bool) (*dto.ListResp[teammodule.TeamResp], error)
	UpdateTeam(context.Context, *teammodule.UpdateTeamReq, int) (*teammodule.TeamResp, error)
	ListTeamProjects(context.Context, *teammodule.TeamProjectListReq, int) (*dto.ListResp[teammodule.TeamProjectItem], error)
	AddTeamMember(context.Context, *teammodule.AddTeamMemberReq, int) error
	RemoveTeamMember(context.Context, int, int, int) error
	UpdateTeamMemberRole(context.Context, *teammodule.UpdateTeamMemberRoleReq, int, int, int) error
	ListTeamMembers(context.Context, *teammodule.ListTeamMemberReq, int) (*dto.ListResp[teammodule.TeamMemberResp], error)
}

type remoteAwareTeamService struct {
	teammodule.HandlerService
	iam teamIAMClient
}

func (s remoteAwareTeamService) CreateTeam(ctx context.Context, req *teammodule.CreateTeamReq, userID int) (*teammodule.TeamResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CreateTeam(ctx, req, userID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) DeleteTeam(ctx context.Context, teamID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DeleteTeam(ctx, teamID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) GetTeamDetail(ctx context.Context, teamID int) (*teammodule.TeamDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetTeam(ctx, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) ListTeams(ctx context.Context, req *teammodule.ListTeamReq, userID int, isAdmin bool) (*dto.ListResp[teammodule.TeamResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListTeams(ctx, req, userID, isAdmin)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) UpdateTeam(ctx context.Context, req *teammodule.UpdateTeamReq, teamID int) (*teammodule.TeamResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.UpdateTeam(ctx, req, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) ListTeamProjects(ctx context.Context, req *teammodule.TeamProjectListReq, teamID int) (*dto.ListResp[teammodule.TeamProjectItem], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListTeamProjects(ctx, req, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) AddMember(ctx context.Context, req *teammodule.AddTeamMemberReq, teamID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AddTeamMember(ctx, req, teamID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) RemoveMember(ctx context.Context, teamID, currentUserID, targetUserID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveTeamMember(ctx, teamID, currentUserID, targetUserID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) UpdateMemberRole(ctx context.Context, req *teammodule.UpdateTeamMemberRoleReq, teamID, targetUserID, currentUserID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.UpdateTeamMemberRole(ctx, req, teamID, targetUserID, currentUserID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) ListMembers(ctx context.Context, req *teammodule.ListTeamMemberReq, teamID int) (*dto.ListResp[teammodule.TeamMemberResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListTeamMembers(ctx, req, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}
