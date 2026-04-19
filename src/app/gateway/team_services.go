package gateway

import (
	"context"

	"aegis/dto"
	team "aegis/module/team"
)

type teamIAMClient interface {
	Enabled() bool
	CreateTeam(context.Context, *team.CreateTeamReq, int) (*team.TeamResp, error)
	DeleteTeam(context.Context, int) error
	GetTeam(context.Context, int) (*team.TeamDetailResp, error)
	ListTeams(context.Context, *team.ListTeamReq, int, bool) (*dto.ListResp[team.TeamResp], error)
	UpdateTeam(context.Context, *team.UpdateTeamReq, int) (*team.TeamResp, error)
	ListTeamProjects(context.Context, *team.TeamProjectListReq, int) (*dto.ListResp[team.TeamProjectItem], error)
	AddTeamMember(context.Context, *team.AddTeamMemberReq, int) error
	RemoveTeamMember(context.Context, int, int, int) error
	UpdateTeamMemberRole(context.Context, *team.UpdateTeamMemberRoleReq, int, int, int) error
	ListTeamMembers(context.Context, *team.ListTeamMemberReq, int) (*dto.ListResp[team.TeamMemberResp], error)
}

type remoteAwareTeamService struct {
	team.HandlerService
	iam teamIAMClient
}

func (s remoteAwareTeamService) CreateTeam(ctx context.Context, req *team.CreateTeamReq, userID int) (*team.TeamResp, error) {
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

func (s remoteAwareTeamService) GetTeamDetail(ctx context.Context, teamID int) (*team.TeamDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetTeam(ctx, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) ListTeams(ctx context.Context, req *team.ListTeamReq, userID int, isAdmin bool) (*dto.ListResp[team.TeamResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListTeams(ctx, req, userID, isAdmin)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) UpdateTeam(ctx context.Context, req *team.UpdateTeamReq, teamID int) (*team.TeamResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.UpdateTeam(ctx, req, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) ListTeamProjects(ctx context.Context, req *team.TeamProjectListReq, teamID int) (*dto.ListResp[team.TeamProjectItem], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListTeamProjects(ctx, req, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) AddMember(ctx context.Context, req *team.AddTeamMemberReq, teamID int) error {
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

func (s remoteAwareTeamService) UpdateMemberRole(ctx context.Context, req *team.UpdateTeamMemberRoleReq, teamID, targetUserID, currentUserID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.UpdateTeamMemberRole(ctx, req, teamID, targetUserID, currentUserID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareTeamService) ListMembers(ctx context.Context, req *team.ListTeamMemberReq, teamID int) (*dto.ListResp[team.TeamMemberResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListTeamMembers(ctx, req, teamID)
	}
	return nil, missingRemoteDependency("iam-service")
}
