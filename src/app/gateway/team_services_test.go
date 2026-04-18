package gatewayapp

import (
	"context"
	"testing"

	"aegis/dto"
	teammodule "aegis/module/team"
)

type iamTeamClientStub struct {
	enabled bool
}

func (s *iamTeamClientStub) Enabled() bool { return s.enabled }

func (s *iamTeamClientStub) CreateTeam(context.Context, *teammodule.CreateTeamReq, int) (*teammodule.TeamResp, error) {
	return &teammodule.TeamResp{ID: 1, Name: "core"}, nil
}
func (s *iamTeamClientStub) DeleteTeam(context.Context, int) error { return nil }
func (s *iamTeamClientStub) GetTeam(context.Context, int) (*teammodule.TeamDetailResp, error) {
	return &teammodule.TeamDetailResp{TeamResp: teammodule.TeamResp{ID: 1, Name: "core"}}, nil
}
func (s *iamTeamClientStub) ListTeams(context.Context, *teammodule.ListTeamReq, int, bool) (*dto.ListResp[teammodule.TeamResp], error) {
	return &dto.ListResp[teammodule.TeamResp]{Items: []teammodule.TeamResp{{ID: 1, Name: "core"}}}, nil
}
func (s *iamTeamClientStub) UpdateTeam(context.Context, *teammodule.UpdateTeamReq, int) (*teammodule.TeamResp, error) {
	return &teammodule.TeamResp{ID: 1, Name: "core"}, nil
}
func (s *iamTeamClientStub) ListTeamProjects(context.Context, *teammodule.TeamProjectListReq, int) (*dto.ListResp[teammodule.TeamProjectItem], error) {
	return &dto.ListResp[teammodule.TeamProjectItem]{}, nil
}
func (s *iamTeamClientStub) AddTeamMember(context.Context, *teammodule.AddTeamMemberReq, int) error {
	return nil
}
func (s *iamTeamClientStub) RemoveTeamMember(context.Context, int, int, int) error { return nil }
func (s *iamTeamClientStub) UpdateTeamMemberRole(context.Context, *teammodule.UpdateTeamMemberRoleReq, int, int, int) error {
	return nil
}
func (s *iamTeamClientStub) ListTeamMembers(context.Context, *teammodule.ListTeamMemberReq, int) (*dto.ListResp[teammodule.TeamMemberResp], error) {
	return &dto.ListResp[teammodule.TeamMemberResp]{}, nil
}

func TestRemoteAwareTeamServiceRequiresIAM(t *testing.T) {
	service := remoteAwareTeamService{}
	if _, err := service.ListTeams(context.Background(), &teammodule.ListTeamReq{}, 7, true); err == nil {
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
