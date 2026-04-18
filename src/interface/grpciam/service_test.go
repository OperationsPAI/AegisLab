package grpciaminterface

import (
	"context"
	"reflect"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"
	authmodule "aegis/module/auth"
	teammodule "aegis/module/team"
	iamv1 "aegis/proto/iam/v1"
	"aegis/utils"

	"google.golang.org/protobuf/types/known/structpb"
)

type middlewareStub struct {
	allowed       bool
	teamAdmin     bool
	teamMember    bool
	teamPublic    bool
	projectAdmin  bool
	projectMember bool
}

func (middlewareStub) VerifyToken(context.Context, string) (*utils.Claims, error) {
	return &utils.Claims{UserID: 1}, nil
}
func (middlewareStub) VerifyServiceToken(context.Context, string) (*utils.ServiceClaims, error) {
	return &utils.ServiceClaims{
		TaskID:           "task-1",
		RegisteredClaims: utils.ServiceClaims{}.RegisteredClaims,
	}, nil
}
func (m middlewareStub) CheckUserPermission(context.Context, *dto.CheckPermissionParams) (bool, error) {
	return m.allowed, nil
}
func (m middlewareStub) IsUserTeamAdmin(context.Context, int, int) (bool, error) {
	return m.teamAdmin, nil
}
func (m middlewareStub) IsUserInTeam(context.Context, int, int) (bool, error) {
	return m.teamMember, nil
}
func (m middlewareStub) IsTeamPublic(context.Context, int) (bool, error) { return m.teamPublic, nil }
func (m middlewareStub) IsUserProjectAdmin(context.Context, int, int) (bool, error) {
	return m.projectAdmin, nil
}
func (m middlewareStub) IsUserInProject(context.Context, int, int) (bool, error) {
	return m.projectMember, nil
}
func (middlewareStub) LogFailedAction(string, string, string, string, int, int, consts.ResourceName) error {
	return nil
}
func (middlewareStub) LogUserAction(string, string, string, string, int, int, consts.ResourceName) error {
	return nil
}

var _ middleware.Service = middlewareStub{}

type teamHandlerStub struct {
	createResp         *teammodule.TeamResp
	detailResp         *teammodule.TeamDetailResp
	listResp           *dto.ListResp[teammodule.TeamResp]
	projectsResp       *dto.ListResp[teammodule.TeamProjectItem]
	membersResp        *dto.ListResp[teammodule.TeamMemberResp]
	updateResp         *teammodule.TeamResp
	createCalled       bool
	listCalled         bool
	listProjectsCalled bool
}

func (s *teamHandlerStub) CreateTeam(context.Context, *teammodule.CreateTeamReq, int) (*teammodule.TeamResp, error) {
	s.createCalled = true
	return s.createResp, nil
}
func (*teamHandlerStub) DeleteTeam(context.Context, int) error { return nil }
func (s *teamHandlerStub) GetTeamDetail(context.Context, int) (*teammodule.TeamDetailResp, error) {
	return s.detailResp, nil
}
func (s *teamHandlerStub) ListTeams(context.Context, *teammodule.ListTeamReq, int, bool) (*dto.ListResp[teammodule.TeamResp], error) {
	s.listCalled = true
	return s.listResp, nil
}
func (s *teamHandlerStub) UpdateTeam(context.Context, *teammodule.UpdateTeamReq, int) (*teammodule.TeamResp, error) {
	return s.updateResp, nil
}
func (s *teamHandlerStub) ListTeamProjects(context.Context, *teammodule.TeamProjectListReq, int) (*dto.ListResp[teammodule.TeamProjectItem], error) {
	s.listProjectsCalled = true
	return s.projectsResp, nil
}
func (*teamHandlerStub) AddMember(context.Context, *teammodule.AddTeamMemberReq, int) error {
	return nil
}
func (*teamHandlerStub) RemoveMember(context.Context, int, int, int) error { return nil }
func (*teamHandlerStub) UpdateMemberRole(context.Context, *teammodule.UpdateTeamMemberRoleReq, int, int, int) error {
	return nil
}
func (s *teamHandlerStub) ListMembers(context.Context, *teammodule.ListTeamMemberReq, int) (*dto.ListResp[teammodule.TeamMemberResp], error) {
	return s.membersResp, nil
}

func TestIAMServerVerifyTokenUser(t *testing.T) {
	token, expiresAt, err := utils.GenerateToken(7, "demo", "demo@example.com", true, false, []string{"user"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	authSvc := authmodule.NewService(nil, nil, nil, nil)
	server := newIAMServer(authSvc, authSvc, &teamHandlerStub{}, nil, nil, middlewareStub{allowed: true})
	resp, err := server.VerifyToken(context.Background(), &iamv1.VerifyTokenRequest{Token: token})
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if !resp.Valid || resp.TokenType != "user" || resp.UserId != 7 {
		t.Fatalf("VerifyToken() unexpected response: %+v", resp)
	}
	if resp.ExpiresAtUnix != expiresAt.Unix() {
		t.Fatalf("VerifyToken() expires_at_unix = %d, want %d", resp.ExpiresAtUnix, expiresAt.Unix())
	}
}

func TestIAMServerVerifyTokenAPIKeyScopes(t *testing.T) {
	token, _, err := utils.GenerateAPIKeyToken(7, "demo", "demo@example.com", true, false, []string{"user"}, 11, []string{"project:read", "execution:write"})
	if err != nil {
		t.Fatalf("GenerateAPIKeyToken() error = %v", err)
	}

	authSvc := authmodule.NewService(nil, nil, nil, nil)
	server := newIAMServer(authSvc, authSvc, &teamHandlerStub{}, nil, nil, middlewareStub{allowed: true})
	resp, err := server.VerifyToken(context.Background(), &iamv1.VerifyTokenRequest{Token: token})
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if resp.AuthType != "api_key" || resp.KeyId != 11 {
		t.Fatalf("VerifyToken() unexpected api-key response: %+v", resp)
	}
	if !reflect.DeepEqual(resp.ApiKeyScopes, []string{"project:read", "execution:write"}) {
		t.Fatalf("VerifyToken() api_key_scopes = %v", resp.ApiKeyScopes)
	}
}

func TestIAMServerCheckPermission(t *testing.T) {
	authSvc := authmodule.NewService(nil, nil, nil, nil)
	server := newIAMServer(authSvc, authSvc, &teamHandlerStub{}, nil, nil, middlewareStub{allowed: true})
	resp, err := server.CheckPermission(context.Background(), &iamv1.CheckPermissionRequest{
		UserId:       7,
		Action:       string(consts.ActionRead),
		Scope:        string(consts.ScopeAll),
		ResourceName: string(consts.ResourceProject),
	})
	if err != nil {
		t.Fatalf("CheckPermission() error = %v", err)
	}
	if !resp.Allowed {
		t.Fatalf("CheckPermission() allowed = false, want true")
	}
}

func TestIAMServerVerifyTokenService(t *testing.T) {
	token, _, err := utils.GenerateServiceToken("task-123")
	if err != nil {
		t.Fatalf("GenerateServiceToken() error = %v", err)
	}

	authSvc := authmodule.NewService(nil, nil, nil, nil)
	server := newIAMServer(authSvc, authSvc, &teamHandlerStub{}, nil, nil, middlewareStub{allowed: true})
	resp, err := server.VerifyToken(context.Background(), &iamv1.VerifyTokenRequest{Token: token})
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
	if !resp.Valid || resp.TokenType != "service" || resp.TaskId != "task-123" {
		t.Fatalf("VerifyToken() unexpected service response: %+v", resp)
	}
	if resp.ExpiresAtUnix <= time.Now().Unix() {
		t.Fatalf("VerifyToken() service expiry = %d, want future timestamp", resp.ExpiresAtUnix)
	}
}

func TestIAMServerMembershipChecks(t *testing.T) {
	authSvc := authmodule.NewService(nil, nil, nil, nil)
	server := newIAMServer(authSvc, authSvc, &teamHandlerStub{}, nil, nil, middlewareStub{
		teamAdmin:     true,
		teamMember:    true,
		teamPublic:    true,
		projectAdmin:  true,
		projectMember: true,
	})

	t.Run("team admin", func(t *testing.T) {
		resp, err := server.IsUserTeamAdmin(context.Background(), &iamv1.UserTeamRequest{UserId: 7, TeamId: 9})
		if err != nil {
			t.Fatalf("IsUserTeamAdmin() error = %v", err)
		}
		if !resp.GetValue() {
			t.Fatalf("IsUserTeamAdmin() value = false, want true")
		}
	})

	t.Run("team member", func(t *testing.T) {
		resp, err := server.IsUserInTeam(context.Background(), &iamv1.UserTeamRequest{UserId: 7, TeamId: 9})
		if err != nil {
			t.Fatalf("IsUserInTeam() error = %v", err)
		}
		if !resp.GetValue() {
			t.Fatalf("IsUserInTeam() value = false, want true")
		}
	})

	t.Run("team public", func(t *testing.T) {
		resp, err := server.IsTeamPublic(context.Background(), &iamv1.TeamRequest{TeamId: 9})
		if err != nil {
			t.Fatalf("IsTeamPublic() error = %v", err)
		}
		if !resp.GetValue() {
			t.Fatalf("IsTeamPublic() value = false, want true")
		}
	})

	t.Run("project admin", func(t *testing.T) {
		resp, err := server.IsUserProjectAdmin(context.Background(), &iamv1.UserProjectRequest{UserId: 7, ProjectId: 11})
		if err != nil {
			t.Fatalf("IsUserProjectAdmin() error = %v", err)
		}
		if !resp.GetValue() {
			t.Fatalf("IsUserProjectAdmin() value = false, want true")
		}
	})

	t.Run("project member", func(t *testing.T) {
		resp, err := server.IsUserInProject(context.Background(), &iamv1.UserProjectRequest{UserId: 7, ProjectId: 11})
		if err != nil {
			t.Fatalf("IsUserInProject() error = %v", err)
		}
		if !resp.GetValue() {
			t.Fatalf("IsUserInProject() value = false, want true")
		}
	})
}

func TestIAMServerTeamRPCs(t *testing.T) {
	teamStub := &teamHandlerStub{
		createResp: &teammodule.TeamResp{ID: 9, Name: "core"},
		detailResp: &teammodule.TeamDetailResp{
			TeamResp:     teammodule.TeamResp{ID: 9, Name: "core"},
			UserCount:    2,
			ProjectCount: 3,
		},
		listResp: &dto.ListResp[teammodule.TeamResp]{
			Items:      []teammodule.TeamResp{{ID: 9, Name: "core"}},
			Pagination: &dto.PaginationInfo{Page: 1, Size: 20, Total: 1, TotalPages: 1},
		},
		projectsResp: &dto.ListResp[teammodule.TeamProjectItem]{
			Items:      []teammodule.TeamProjectItem{{ID: 11, Name: "proj-a"}},
			Pagination: &dto.PaginationInfo{Page: 1, Size: 20, Total: 1, TotalPages: 1},
		},
	}
	authSvc := authmodule.NewService(nil, nil, nil, nil)
	server := newIAMServer(authSvc, authSvc, teamStub, nil, nil, middlewareStub{})

	createBody, _ := structpb.NewStruct(map[string]any{"name": "core"})
	createResp, err := server.CreateTeam(context.Background(), &iamv1.CreateTeamRequest{
		UserId: 7,
		Body:   createBody,
	})
	if err != nil {
		t.Fatalf("CreateTeam() error = %v", err)
	}
	if createResp.GetData().AsMap()["id"] != float64(9) || !teamStub.createCalled {
		t.Fatalf("CreateTeam() unexpected response: %+v", createResp.GetData().AsMap())
	}

	listQuery, _ := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	listResp, err := server.ListTeams(context.Background(), &iamv1.ListTeamsRequest{
		UserId:  7,
		IsAdmin: true,
		Query:   listQuery,
	})
	if err != nil {
		t.Fatalf("ListTeams() error = %v", err)
	}
	items, ok := listResp.GetData().AsMap()["items"].([]any)
	if !ok || len(items) != 1 || !teamStub.listCalled {
		t.Fatalf("ListTeams() unexpected response: %+v", listResp.GetData().AsMap())
	}

	getResp, err := server.GetTeam(context.Background(), &iamv1.TeamRequest{TeamId: 9})
	if err != nil {
		t.Fatalf("GetTeam() error = %v", err)
	}
	if getResp.GetData().AsMap()["project_count"] != float64(3) {
		t.Fatalf("GetTeam() unexpected response: %+v", getResp.GetData().AsMap())
	}

	projectResp, err := server.ListTeamProjects(context.Background(), &iamv1.ListTeamProjectsRequest{
		TeamId: 9,
		Query:  listQuery,
	})
	if err != nil {
		t.Fatalf("ListTeamProjects() error = %v", err)
	}
	projectItems, ok := projectResp.GetData().AsMap()["items"].([]any)
	if !ok || len(projectItems) != 1 || !teamStub.listProjectsCalled {
		t.Fatalf("ListTeamProjects() unexpected response: %+v", projectResp.GetData().AsMap())
	}
}
