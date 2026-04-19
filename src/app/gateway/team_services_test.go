package gateway

import (
	"context"
	"testing"

	"aegis/dto"
	team "aegis/module/team"
)

type iamTeamClientStub struct {
	enabled bool
}

func (s *iamTeamClientStub) Enabled() bool { return s.enabled }

func (s *iamTeamClientStub) CreateTeam(context.Context, *team.CreateTeamReq, int) (*team.TeamResp, error) {
	return &team.TeamResp{ID: 1, Name: "core"}, nil
}
func (s *iamTeamClientStub) DeleteTeam(context.Context, int) error { return nil }
func (s *iamTeamClientStub) GetTeam(context.Context, int) (*team.TeamDetailResp, error) {
	return &team.TeamDetailResp{TeamResp: team.TeamResp{ID: 1, Name: "core"}}, nil
}
func (s *iamTeamClientStub) ListTeams(context.Context, *team.ListTeamReq, int, bool) (*dto.ListResp[team.TeamResp], error) {
	return &dto.ListResp[team.TeamResp]{Items: []team.TeamResp{{ID: 1, Name: "core"}}}, nil
}
func (s *iamTeamClientStub) UpdateTeam(context.Context, *team.UpdateTeamReq, int) (*team.TeamResp, error) {
	return &team.TeamResp{ID: 1, Name: "core"}, nil
}
func (s *iamTeamClientStub) ListTeamProjects(context.Context, *team.TeamProjectListReq, int) (*dto.ListResp[team.TeamProjectItem], error) {
	return &dto.ListResp[team.TeamProjectItem]{}, nil
}
func (s *iamTeamClientStub) AddTeamMember(context.Context, *team.AddTeamMemberReq, int) error {
	return nil
}
func (s *iamTeamClientStub) RemoveTeamMember(context.Context, int, int, int) error { return nil }
func (s *iamTeamClientStub) UpdateTeamMemberRole(context.Context, *team.UpdateTeamMemberRoleReq, int, int, int) error {
	return nil
}
func (s *iamTeamClientStub) ListTeamMembers(context.Context, *team.ListTeamMemberReq, int) (*dto.ListResp[team.TeamMemberResp], error) {
	return &dto.ListResp[team.TeamMemberResp]{}, nil
}

func TestRemoteAwareTeamServiceRequiresIAM(t *testing.T) {
	service := remoteAwareTeamService{}
	if _, err := service.ListTeams(context.Background(), &team.ListTeamReq{}, 7, true); err == nil {
		t.Fatal("ListTeams() error = nil, want missing dependency")
	}
}

func TestRemoteAwareTeamServiceUsesIAMClient(t *testing.T) {
	service := remoteAwareTeamService{iam: &iamTeamClientStub{enabled: true}}
	resp, err := service.GetTeamDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTeamDetail() error = %v", err)
	}
	if resp.ID != 1 || resp.Name != "core" {
		t.Fatalf("GetTeamDetail() unexpected response: %+v", resp)
	}
}
