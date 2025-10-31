package service

import (
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CreateContainer handles the atomic creation of a new container resource,
// including its initial versions and assigning the creating user as container administrator
func CreateContainer(req *dto.CreateContainerRequest, userID int) (*dto.ContainerResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	container := req.ConvertToContainer()

	var version *database.ContainerVersion
	var helmConfig *database.HelmConfig
	if req.VersionRequest != nil {
		version = req.VersionRequest.ConvertToContainerVersion()
		if req.VersionRequest.HelmConfigRequest != nil {
			var err error
			helmConfig, err = req.VersionRequest.HelmConfigRequest.ConvertToHelmConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to convert helm config request: %w", err)
			}
		}
	}

	var createdContainer *database.Container
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if version != nil {
			container, err = createContainerCore(tx, container, []database.ContainerVersion{*version}, []*database.HelmConfig{helmConfig}, userID)
		} else {
			container, err = createContainerCore(tx, container, nil, nil, userID)
		}

		if err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}

		createdContainer = container
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	var resp dto.ContainerResponse
	resp.ConvertFromContainer(createdContainer)
	return &resp, nil
}

// CreateContainerVersion creates a new version for an existing container
func CreateContainerVersion(req *dto.CreateContainerVersionRequest, containerID, userID int) (*dto.ContainerVersionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create container version request is nil")
	}

	version := req.ConvertToContainerVersion()
	version.ContainerID = containerID

	var helmConfig *database.HelmConfig
	if req.HelmConfigRequest != nil {
		var err error
		helmConfig, err = req.HelmConfigRequest.ConvertToHelmConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to convert helm config request: %w", err)
		}
	}

	var createdVersion *database.ContainerVersion
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		versions, err := createContainerVersionsCore(tx, []database.ContainerVersion{*version}, []*database.HelmConfig{helmConfig})
		if err != nil {
			return fmt.Errorf("failed to create container version: %w", err)
		}

		createdVersion = &versions[0]
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container version: %w", err)
	}

	var resp dto.ContainerVersionResponse
	resp.ConvertFromContainerVersion(createdVersion)
	return &resp, nil
}

// DeleteContainer deletes an existing container
func DeleteContainer(containerID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		versions, err := repository.GetContainerVersionsByContainerID(tx, containerID)
		if err != nil {
			return fmt.Errorf("failed to get container versions: %w", err)
		}

		versionIDs := make([]int, 0, len(versions))
		for _, version := range versions {
			versionIDs = append(versionIDs, version.ID)
		}

		if err := repository.BatchSoftDeleteHelmConfigsByVersionIDs(tx, versionIDs); err != nil {
			return fmt.Errorf("failed to soft delete helm configs: %w", err)
		}

		if err := repository.BatchSoftDeleteContainerVersions(tx, containerID); err != nil {
			return fmt.Errorf("failed to soft delete container versions: %w", err)
		}

		container, err := repository.GetContainerByID(tx, containerID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: container not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get container: %w", err)
		}

		container.Status = consts.CommonDeleted
		if err := repository.UpdateContainer(tx, container); err != nil {
			return fmt.Errorf("failed to update container: %w", err)
		}

		if err := repository.RemoveAllUsersFromContainer(tx, containerID); err != nil {
			return fmt.Errorf("failed to remove all users from container: %w", err)
		}

		return nil
	})
}

// DeleteContainerVersion deletes a specific version of a container
func DeleteContainerVersion(containerID, versionID int) error {
	versionIDs := []int{versionID}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.BatchSoftDeleteHelmConfigsByVersionIDs(tx, versionIDs); err != nil {
			return fmt.Errorf("failed to soft delete helm config: %w", err)
		}

		version, err := repository.GetContainerVersionByID(tx, versionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get container: %w", err)
		}

		if version.Status == consts.CommonDeleted {
			return fmt.Errorf("%w: version %d already deleted", consts.ErrNotFound, versionID)
		}

		version.Status = consts.CommonDeleted
		if err := repository.UpdateContainerVersion(tx, version); err != nil {
			return fmt.Errorf("failed to update container: %w", err)
		}

		return nil
	})
}

// GetContainerDetail retrieves detailed information about a specific container,
// including its versions and associated Helm configurations
func GetContainerDetail(containerID int) (*dto.ContainerDetailResponse, error) {
	container, err := repository.GetContainerByID(database.DB, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	versions, err := repository.GetContainerVersionsByContainerID(database.DB, container.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container versions: %w", err)
	}

	var resp dto.ContainerDetailResponse
	resp.ConvertFromContainer(container)

	for _, version := range versions {
		var versionResp dto.ContainerVersionResponse
		versionResp.ConvertFromContainerVersion(&version)
		resp.Versions = append(resp.Versions, versionResp)
	}

	return &resp, nil
}

// GetContainerVersionDetail retrieves detailed information about a specific container version,
// including its Helm configuration if available
func GetContainerVersionDetail(containerID, versionID int) (*dto.ContainerVersionDetailResponse, error) {
	version, err := repository.GetContainerVersionByID(database.DB, versionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return nil, fmt.Errorf("failed to get container version: %w", err)
	}

	var resp dto.ContainerVersionDetailResponse
	resp.ConvertFromContainerVersion(version)

	helmConfig, err := repository.GetHelmConfigByContainerVersionID(database.DB, version.ID)
	if err != nil && !errors.Is(err, consts.ErrNotFound) {
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}
	if helmConfig != nil {
		var helmConfigResp dto.HelmConfigDetailResponse
		if err := helmConfigResp.ConvertFromHelmConfig(helmConfig); err != nil {
			return nil, fmt.Errorf("failed to convert helm config: %w", err)
		}
		resp.HelmConfig = &helmConfigResp
	}

	return &resp, nil
}

// ListContainers lists containers based on the provided filters
func ListContainers(req *dto.ListContainerRequest) (*dto.ListResponse[dto.ContainerResponse], error) {
	limit, offset := req.ToGormParams()

	containers, total, err := repository.ListContainers(database.DB, limit, offset, req.Type, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerResps := make([]dto.ContainerResponse, len(containers))
	for i, c := range containers {
		var containerResp dto.ContainerResponse
		containerResp.ConvertFromContainer(&c)
		containerResps[i] = containerResp
	}

	resp := dto.ListResponse[dto.ContainerResponse]{
		Items:      containerResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateContainer updates an existing container's details
func UpdateContainer(req *dto.UpdateContainerRequest, containerID int) (*dto.ContainerResponse, error) {
	var updatedContainer *database.Container

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingContainer, err := repository.GetContainerByID(tx, containerID)
		if err != nil {
			return fmt.Errorf("failed to get container: %w", err)
		}

		req.PatchContainerModel(existingContainer)

		if err := repository.UpdateContainer(tx, existingContainer); err != nil {
			return fmt.Errorf("failed to update container: %w", err)
		}

		updatedContainer = existingContainer
		return nil
	})
	if err != nil {
		return nil, err
	}

	var resp dto.ContainerResponse
	resp.ConvertFromContainer(updatedContainer)
	return &resp, nil
}

// UpdateContainerVersion updates an existing container version's details
func UpdateContainerVersion(req *dto.UpdateContainerVersionRequest, containerID, versionID int) (*dto.ContainerVersionResponse, error) {
	var updatedVersion *database.ContainerVersion

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingVersion, err := repository.GetContainerVersionByID(tx, versionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get container: %w", err)
		}

		req.PatchContainerVersionModel(existingVersion)
		if err := repository.UpdateContainerVersion(tx, existingVersion); err != nil {
			return fmt.Errorf("failed to update container: %w", err)
		}

		updatedVersion = existingVersion

		if req.HelmConfigRequest != nil {
			existingHelmConfig, err := repository.GetHelmConfigByContainerVersionID(tx, existingVersion.ID)
			if err != nil && !errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("failed to get helm config: %w", err)
			}

			if err := req.HelmConfigRequest.PatchHelmConfigModel(existingHelmConfig); err != nil {
				return fmt.Errorf("failed to patch helm config model: %w", err)
			}
			if err := repository.UpdateHelmConfig(tx, existingHelmConfig); err != nil {
				return fmt.Errorf("failed to update helm config: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var resp dto.ContainerVersionResponse
	resp.ConvertFromContainerVersion(updatedVersion)
	return &resp, nil
}

// ProcessGitHubSource processes the GitHub source specified in the container creation request
func ProcessGitHubSource(req *dto.BuildContainerRequest) (string, error) {
	targetDir := filepath.Join(config.GetString("container.storage_path"), req.ImageName, fmt.Sprintf("build_%d", time.Now().Unix()))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	repoURL := fmt.Sprintf("https://github.com/%s.git", req.GithubRepository)
	if req.GithubToken != "" {
		repoURL = fmt.Sprintf("https://%s@github.com/%s.git", req.GithubToken, req.GithubRepository)
	}

	gitCmd := []string{"git", "clone"}
	if req.GithubBranch != "" {
		gitCmd = append(gitCmd, repoURL, targetDir)
	} else {
		gitCmd = append(gitCmd, "--branch", req.GithubBranch, "--single-branch", repoURL, targetDir)
	}

	if req.GithubCommit != "" {
		cmd := exec.Command(gitCmd[0], gitCmd[1:]...)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to clone repository: %w", err)
		}

		// Checkout specific commit
		cmd = exec.Command("git", "-C", targetDir, "checkout", req.GithubCommit)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to checkout commit %s: %w", req.GithubCommit, err)
		}
	} else {
		cmd := exec.Command(gitCmd[0], gitCmd[1:]...)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	// If a specific path is provided, copy only that subdirectory
	if req.SubPath != "" {
		sourcePath := filepath.Join(targetDir, req.SubPath)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return "", fmt.Errorf("sub path '%s' does not exist in repository", req.SubPath)
		}

		newTargetDir := filepath.Join(config.GetString("container.storage_path"), req.ImageName, fmt.Sprintf("build_final_%d", time.Now().Unix()))
		if err := utils.CopyDir(sourcePath, newTargetDir); err != nil {
			return "", fmt.Errorf("failed to copy subdirectory: %w", err)
		}

		// Clean up the full clone
		if err := os.RemoveAll(targetDir); err != nil {
			logrus.WithField("target_dir", targetDir).Warnf("failed to remove temporary directory: %v", err)
		}

		targetDir = newTargetDir
	}

	return targetDir, nil
}

// createContainerCore performs the core logic of creating a container within a transaction
func createContainerCore(tx *gorm.DB, container *database.Container, versions []database.ContainerVersion, helmConfigs []*database.HelmConfig, userID int) (*database.Container, error) {
	if err := repository.CreateContainer(tx, container); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, consts.ErrAlreadyExists
		}

		return nil, err
	}

	role, err := repository.GetRoleByName(tx, string(consts.RoleContainerAdmin))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin role: %w", err)
	}

	if err := repository.CreateUserContainer(tx, &database.UserContainer{
		UserID:      userID,
		ContainerID: container.ID,
		RoleID:      role.ID,
		Status:      consts.CommonEnabled,
	}); err != nil {
		return nil, fmt.Errorf("failed to associate container with user: %w", err)
	}

	for i := range versions {
		versions[i].ContainerID = container.ID
	}

	_, err = createContainerVersionsCore(tx, versions, helmConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create container versions: %w", err)
	}

	return container, nil
}

// createContainerVersionCore performs the core logic of creating container versions within a transaction
func createContainerVersionsCore(db *gorm.DB, versions []database.ContainerVersion, helmConfigs []*database.HelmConfig) ([]database.ContainerVersion, error) {
	if len(versions) == 0 {
		return nil, nil
	}

	if err := repository.BatchCreateContainerVersions(db, versions); err != nil {
		return nil, fmt.Errorf("failed to create container versions: %w", err)
	}

	if len(helmConfigs) == 0 {
		return versions, nil
	}

	var helmConfigsToCreate []database.HelmConfig
	for i, version := range versions {
		if i < len(helmConfigs) && helmConfigs[i] != nil {
			helmConfigs[i].ContainerVersionID = version.ID
			helmConfigsToCreate = append(helmConfigsToCreate, *helmConfigs[i])
		}
	}

	if len(helmConfigsToCreate) > 0 {
		if err := repository.BatchCreateHelmConfigs(db, helmConfigsToCreate); err != nil {
			return nil, fmt.Errorf("failed to create helm configs: %w", err)
		}
	}

	return versions, nil
}
