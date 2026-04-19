package team

import (
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"

	"gorm.io/gorm"
)

type Service struct {
	repo     *Repository
	projects projectReader
}

func NewService(repo *Repository, projects projectReader) *Service {
	return &Service{
		repo:     repo,
		projects: projects,
	}
}

func (s *Service) CreateTeam(_ context.Context, req *CreateTeamReq, userID int) (*TeamResp, error) {
	team := req.ConvertToTeam()

	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if err := NewRepository(tx).createTeamWithCreator(team, userID); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return consts.ErrAlreadyExists
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewTeamResp(team), nil
}

func (s *Service) DeleteTeam(_ context.Context, teamID int) error {
	rowsAffected, err := s.repo.deleteTeam(teamID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return consts.ErrNotFound
	}
	return nil
}

func (s *Service) GetTeamDetail(ctx context.Context, teamID int) (*TeamDetailResp, error) {
	team, userCount, err := s.repo.loadTeamDetailBase(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrNotFound
		}
		return nil, err
	}
	projectCount, err := s.projects.CountProjects(ctx, teamID)
	if err != nil {
		return nil, err
	}

	resp := NewTeamDetailResp(team)
	resp.UserCount = userCount
	resp.ProjectCount = projectCount

	return resp, nil
}

func (s *Service) ListTeams(_ context.Context, req *ListTeamReq, userID int, isAdmin bool) (*dto.ListResp[TeamResp], error) {
	limit, offset := req.ToGormParams()
	teams, total, err := s.repo.listVisibleTeams(limit, offset, req, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	items := make([]TeamResp, len(teams))
	for i, team := range teams {
		items[i] = *NewTeamResp(&team)
	}

	return &dto.ListResp[TeamResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateTeam(_ context.Context, req *UpdateTeamReq, teamID int) (*TeamResp, error) {
	team, err := s.repo.updateMutableTeam(teamID, func(team *model.Team) {
		req.PatchTeamModel(team)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrNotFound
		}
		return nil, err
	}

	return NewTeamResp(team), nil
}

func (s *Service) ListTeamProjects(ctx context.Context, req *TeamProjectListReq, teamID int) (*dto.ListResp[TeamProjectItem], error) {
	return s.projects.ListProjects(ctx, req, teamID)
}

func (s *Service) AddMember(_ context.Context, req *AddTeamMemberReq, teamID int) error {
	if err := s.repo.addMember(teamID, req.Username, req.RoleID); err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return consts.ErrNotFound
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user or role not found")
		}
		if errors.Is(err, consts.ErrAlreadyExists) {
			return consts.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *Service) RemoveMember(_ context.Context, teamID, currentUserID, targetUserID int) error {
	if targetUserID == currentUserID {
		return fmt.Errorf("cannot remove yourself from the team")
	}

	rowsAffected, err := s.repo.removeMember(teamID, targetUserID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrNotFound
		}
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user is not a member of this team")
	}
	return nil
}

func (s *Service) UpdateMemberRole(_ context.Context, req *UpdateTeamMemberRoleReq, teamID, targetUserID, currentUserID int) error {
	_ = currentUserID

	if err := s.repo.updateMemberRole(teamID, targetUserID, req.RoleID); err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return consts.ErrNotFound
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role not found")
		}
		return err
	}
	return nil
}

func (s *Service) ListMembers(_ context.Context, req *ListTeamMemberReq, teamID int) (*dto.ListResp[TeamMemberResp], error) {
	limit, offset := req.ToGormParams()
	members, total, err := s.repo.listTeamMembers(teamID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &dto.ListResp[TeamMemberResp]{
		Items:      members,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) IsUserInTeam(userID, teamID int) (bool, error) {
	ut, err := s.repo.loadUserTeamMembership(userID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return ut != nil, nil
}

func (s *Service) IsUserTeamAdmin(userID, teamID int) (bool, error) {
	ut, err := s.repo.loadUserTeamMembership(userID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return ut != nil && ut.Role != nil && ut.Role.Name == consts.RoleTeamAdmin.String(), nil
}

func (s *Service) IsTeamPublic(teamID int) (bool, error) {
	isPublic, err := s.repo.isTeamPublic(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return isPublic, nil
}
