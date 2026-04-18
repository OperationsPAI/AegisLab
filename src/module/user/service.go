package usermodule

import (
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	rbacmodule "aegis/module/rbac"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(_ context.Context, req *CreateUserReq) (*UserResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   consts.CommonEnabled,
		IsActive: true,
	}

	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		return repo.createUserIfUnique(user)
	}); err != nil {
		return nil, err
	}

	return NewUserResp(user), nil
}

func (s *Service) DeleteUser(_ context.Context, userID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		rows, err := repo.deleteUserCascade(userID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: user id %d not found", consts.ErrNotFound, userID)
		}
		return nil
	})
}

func (s *Service) GetUserDetail(_ context.Context, userID int) (*UserDetailResp, error) {
	user, err := s.repo.getUserDetailBase(userID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: user with ID %d not found", consts.ErrNotFound, userID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	resp := NewUserDetailResp(user)

	globalRoles, permissions, userContainers, userDatasets, userProjects, err := s.repo.loadUserDetailRelations(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user detail relations: %w", err)
	}
	resp.GlobalRoles = make([]rbacmodule.RoleResp, len(globalRoles))
	for i, role := range globalRoles {
		resp.GlobalRoles[i] = *rbacmodule.NewRoleResp(&role)
	}

	resp.Permissions = make([]rbacmodule.PermissionResp, len(permissions))
	for i, permission := range permissions {
		resp.Permissions[i] = *rbacmodule.NewPermissionResp(&permission)
	}

	containerRoles, datasetRoles, projectRoles := buildUserResourceRoles(userContainers, userDatasets, userProjects)
	resp.ContainerRoles = containerRoles
	resp.DatasetRoles = datasetRoles
	resp.ProjectRoles = projectRoles

	return resp, nil
}

func (s *Service) ListUsers(_ context.Context, req *ListUserReq) (*dto.ListResp[UserResp], error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	limit, offset := req.ToGormParams()
	users, total, err := s.repo.listUserViews(limit, offset, req.IsActive, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	items := make([]UserResp, len(users))
	for i, user := range users {
		items[i] = *NewUserResp(&user)
	}

	return &dto.ListResp[UserResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateUser(_ context.Context, req *UpdateUserReq, userID int) (*UserResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedUser *model.User
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		user, err := repo.updateMutableUser(userID, func(existingUser *model.User) {
			req.PatchUserModel(existingUser)
		})
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		updatedUser = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewUserResp(updatedUser), nil
}

func (s *Service) AssignRole(_ context.Context, userID, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.assignGlobalRole(userID, roleID); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				if userErr := repo.ensureActiveRecordExists(&model.User{}, userID, "user"); userErr != nil {
					return fmt.Errorf("%w: user not found", consts.ErrNotFound)
				}
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already has this role", consts.ErrAlreadyExists)
			}
			return err
		}
		return nil
	})
}

func (s *Service) RemoveRole(_ context.Context, userID, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.removeGlobalRole(userID, roleID); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				if userErr := repo.ensureActiveRecordExists(&model.User{}, userID, "user"); userErr != nil {
					return fmt.Errorf("%w: user not found", consts.ErrNotFound)
				}
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}
		return nil
	})
}

func (s *Service) AssignPermissions(_ context.Context, req *AssignUserPermissionReq, userID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		userPermissions, err := repo.buildUserPermissions(userID, req.Items)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: failed to resolve permission assignment targets", consts.ErrNotFound)
			}
			return err
		}

		if err := repo.batchCreateUserPermissions(userPermissions); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already has one or more of these permissions", consts.ErrAlreadyExists)
			}
			return fmt.Errorf("failed to assign permissions to user: %w", err)
		}
		return nil
	})
}

func (s *Service) RemovePermissions(_ context.Context, req *RemoveUserPermissionReq, userID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.batchDeleteUserPermissions(userID, req.PermissionIDs); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: failed to resolve user or permissions", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to remove permissions from user: %w", err)
		}
		return nil
	})
}

func (s *Service) AssignContainer(_ context.Context, userID, containerID, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.assignContainerRole(userID, containerID, roleID); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user/container/role not found", consts.ErrNotFound)
			}
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already assigned to this container", consts.ErrAlreadyExists)
			}
			return err
		}
		return nil
	})
}

func (s *Service) RemoveContainer(_ context.Context, userID, containerID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		rows, err := repo.removeContainerRole(userID, containerID)
		if err != nil {
			return fmt.Errorf("failed to remove user from container: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: user is not assigned to this container", consts.ErrNotFound)
		}
		return nil
	})
}

func (s *Service) AssignDataset(_ context.Context, userID, datasetID, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.assignDatasetRole(userID, datasetID, roleID); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user/dataset/role not found", consts.ErrNotFound)
			}
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already assigned to this dataset", consts.ErrAlreadyExists)
			}
			return err
		}
		return nil
	})
}

func (s *Service) RemoveDataset(_ context.Context, userID, datasetID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		rows, err := repo.removeDatasetRole(userID, datasetID)
		if err != nil {
			return fmt.Errorf("failed to remove user from dataset: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: user is not assigned to this dataset", consts.ErrNotFound)
		}
		return nil
	})
}

func (s *Service) AssignProject(_ context.Context, userID, projectID, roleID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if err := repo.assignProjectRole(userID, projectID, roleID); err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user/project/role not found", consts.ErrNotFound)
			}
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already assigned to this project", consts.ErrAlreadyExists)
			}
			return err
		}
		return nil
	})
}

func (s *Service) RemoveProject(_ context.Context, userID, projectID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		rows, err := repo.removeProjectRole(userID, projectID)
		if err != nil {
			return fmt.Errorf("failed to remove user from project: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: user is not assigned to this project", consts.ErrNotFound)
		}
		return nil
	})
}

func buildUserResourceRoles(userContainers []model.UserContainer, userDatasets []model.UserDataset, userProjects []model.UserProject) ([]UserContainerInfo, []UserDatasetInfo, []UserProjectInfo) {
	containerRoles := make([]UserContainerInfo, 0, len(userContainers))
	for _, uc := range userContainers {
		containerRoles = append(containerRoles, *NewUserContainerInfo(&uc))
	}

	datasetRoles := make([]UserDatasetInfo, 0, len(userDatasets))
	for _, ud := range userDatasets {
		datasetRoles = append(datasetRoles, *NewUserDatasetInfo(&ud))
	}

	projectRoles := make([]UserProjectInfo, 0, len(userProjects))
	for _, up := range userProjects {
		projectRoles = append(projectRoles, *NewUserProjectInfo(&up))
	}

	return containerRoles, datasetRoles, projectRoles
}
