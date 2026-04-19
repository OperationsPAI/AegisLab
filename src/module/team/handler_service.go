package team

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the team operations consumed by the HTTP handler.
type HandlerService interface {
	CreateTeam(context.Context, *CreateTeamReq, int) (*TeamResp, error)
	DeleteTeam(context.Context, int) error
	GetTeamDetail(context.Context, int) (*TeamDetailResp, error)
	ListTeams(context.Context, *ListTeamReq, int, bool) (*dto.ListResp[TeamResp], error)
	UpdateTeam(context.Context, *UpdateTeamReq, int) (*TeamResp, error)
	ListTeamProjects(context.Context, *TeamProjectListReq, int) (*dto.ListResp[TeamProjectItem], error)
	AddMember(context.Context, *AddTeamMemberReq, int) error
	RemoveMember(context.Context, int, int, int) error
	UpdateMemberRole(context.Context, *UpdateTeamMemberRoleReq, int, int, int) error
	ListMembers(context.Context, *ListTeamMemberReq, int) (*dto.ListResp[TeamMemberResp], error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
