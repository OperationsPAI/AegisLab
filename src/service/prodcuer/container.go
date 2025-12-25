package producer

import (
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// =====================================================================
// Container Service Layer
// =====================================================================

// CreateContainer handles the atomic creation of a new container resource,
// including its initial versions and assigning the creating user as container administrator
func CreateContainer(req *dto.CreateContainerReq, userID int) (*dto.ContainerResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	container := req.ConvertToContainer()

	var createdContainer *database.Container
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		container, err := CreateContainerCore(tx, container, userID)

		if err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}

		createdContainer = container
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return dto.NewContainerResp(createdContainer), nil
}

// CreateContainerCore performs the core logic of creating a container within a transaction
func CreateContainerCore(tx *gorm.DB, container *database.Container, userID int) (*database.Container, error) {
	role, err := repository.GetRoleByName(tx, consts.RoleContainerAdmin)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role %v not found", err, consts.RoleContainerAdmin)
		}
		return nil, fmt.Errorf("failed to get project owner role: %w", err)
	}

	if err := repository.CreateContainer(tx, container); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, consts.ErrAlreadyExists
		}
		return nil, err
	}

	if err := repository.CreateUserContainer(tx, &database.UserContainer{
		UserID:      userID,
		ContainerID: container.ID,
		RoleID:      role.ID,
		Status:      consts.CommonEnabled,
	}); err != nil {
		return nil, fmt.Errorf("failed to associate container with user: %w", err)
	}

	if len(container.Versions) > 0 {
		for i := range container.Versions {
			container.Versions[i].ContainerID = container.ID
			container.Versions[i].UserID = userID
		}

		_, err = createContainerVersionsCore(tx, container.Versions)
		if err != nil {
			return nil, fmt.Errorf("failed to create container versions: %w", err)
		}
	}

	return container, nil
}

// DeleteContainer deletes an existing container (Service Layer)
func DeleteContainer(containerID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := repository.BatchDeleteContainerVersions(tx, containerID); err != nil {
			return fmt.Errorf("failed to delete container versions: %w", err)
		}

		if _, err := repository.RemoveUsersFromContainer(tx, containerID); err != nil {
			return fmt.Errorf("failed to remove all users from container: %w", err)
		}

		if err := repository.ClearContainerLabels(tx, []int{containerID}, nil); err != nil {
			return fmt.Errorf("failed to clear container labels: %w", err)
		}

		rows, err := repository.DeleteContainer(tx, containerID)
		if err != nil {
			return fmt.Errorf("failed to delete container: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: container id %d not found", consts.ErrNotFound, containerID)
		}

		return nil
	})
}

// GetContainerDetail retrieves detailed information about a specific container,
// including its versions and associated Helm configurations
func GetContainerDetail(containerID int) (*dto.ContainerDetailResp, error) {
	container, err := repository.GetContainerByID(database.DB, containerID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: container id: %d", consts.ErrNotFound, containerID)
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	versions, err := repository.ListContainerVersionsByContainerID(database.DB, container.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container versions: %w", err)
	}

	resp := dto.NewContainerDetailResp(container)
	for _, version := range versions {
		resp.Versions = append(resp.Versions, *dto.NewContainerVersionResp(&version))
	}

	return resp, nil
}

// ListContainers lists containers based on the provided filters
func ListContainers(req *dto.ListContainerReq) (*dto.ListResp[dto.ContainerResp], error) {
	limit, offset := req.ToGormParams()

	containers, total, err := repository.ListContainers(database.DB, limit, offset, req.Type, req.IsPublic, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := make([]int, 0, len(containers))
	for _, c := range containers {
		containerIDs = append(containerIDs, c.ID)
	}

	labelsMap, err := repository.ListContainerLabels(database.DB, containerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list container labels: %w", err)
	}

	containerResps := make([]dto.ContainerResp, 0, len(containers))
	for _, container := range containers {
		if labels, exists := labelsMap[container.ID]; exists {
			container.Labels = labels
		}
		containerResps = append(containerResps, *dto.NewContainerResp(&container))
	}

	resp := dto.ListResp[dto.ContainerResp]{
		Items:      containerResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateContainer updates an existing container's details
func UpdateContainer(req *dto.UpdateContainerReq, containerID int) (*dto.ContainerResp, error) {
	var updatedContainer *database.Container

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingContainer, err := repository.GetContainerByID(tx, containerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: container with id %d not found", consts.ErrNotFound, containerID)
			}
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

	return dto.NewContainerResp(updatedContainer), nil
}

// ===================== ContainerLabel =====================

// ManageContainerLabels handles adding and removing labels for a container
func ManageContainerLabels(req *dto.ManageContainerLabelReq, containerID int) (*dto.ContainerResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var managedContainer *database.Container
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		container, err := repository.GetContainerByID(tx, containerID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: container not found", consts.ErrNotFound)
			}
			return err
		}

		// Add labels
		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.ContainerCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			containerLabels := make([]database.ContainerLabel, 0, len(labels))
			for _, label := range labels {
				containerLabels = append(containerLabels, database.ContainerLabel{
					ContainerID: containerID,
					LabelID:     label.ID,
				})
			}

			if err := repository.AddContainerLabels(tx, containerLabels); err != nil {
				return fmt.Errorf("failed to add container labels: %w", err)
			}
		}

		// Remove labels
		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndContainerID(tx, containerID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label IDs: %w", err)
			}

			if len(labelIDs) == 0 {
				return nil
			}

			if err := repository.ClearContainerLabels(tx, []int{containerID}, labelIDs); err != nil {
				return fmt.Errorf("failed to delete container-label associations: %w", err)
			}

			if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
				return fmt.Errorf("failed to decrease label usage counts: %w", err)
			}
		}

		labels, err := repository.ListLabelsByContainerID(database.DB, container.ID)
		if err != nil {
			return fmt.Errorf("failed to get container labels: %w", err)
		}

		container.Labels = labels
		managedContainer = container
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewContainerResp(managedContainer), nil
}

// =====================================================================
// ContainerVersion Service Layer
// =====================================================================

// CreateContainerVersion creates a new version for an existing container
func CreateContainerVersion(req *dto.CreateContainerVersionReq, containerID, userID int) (*dto.ContainerVersionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("create container version request is nil")
	}

	version := req.ConvertToContainerVersion()
	version.ContainerID = containerID
	version.UserID = userID

	var createdVersion *database.ContainerVersion
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		versions, err := createContainerVersionsCore(tx, []database.ContainerVersion{*version})
		if err != nil {
			return fmt.Errorf("failed to create container version: %w", err)
		}

		createdVersion = &versions[0]
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container version: %w", err)
	}

	return dto.NewContainerVersionResp(createdVersion), nil
}

// DeleteContainerVersion deletes a specific version of a container
func DeleteContainerVersion(versionID int) error {
	rows, err := repository.DeleteContainer(database.DB, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete container version: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: container version id %d not found", consts.ErrNotFound, versionID)
	}
	return nil
}

// GetContainerVersionDetail retrieves detailed information about a specific container version,
// including its Helm configuration if available
func GetContainerVersionDetail(containerID, versionID int) (*dto.ContainerVersionDetailResp, error) {
	_, err := repository.GetContainerByID(database.DB, containerID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: container id: %d", consts.ErrNotFound, containerID)
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	version, err := repository.GetContainerVersionByID(database.DB, versionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return nil, fmt.Errorf("failed to get container version: %w", err)
	}

	resp := dto.NewContainerVersionDetailResp(version)

	helmConfig, err := repository.GetHelmConfigByContainerVersionID(database.DB, version.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}
	if helmConfig != nil {
		helmConfigResp, err := dto.NewHelmConfigDetailResp(helmConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to convert helm config: %w", err)
		}
		resp.HelmConfig = helmConfigResp
	}

	return resp, nil
}

// ListContainerVersions lists container versions with pagination and optional status filtering
func ListContainerVersions(req *dto.ListContainerVersionReq, containerID int) (*dto.ListResp[dto.ContainerVersionResp], error) {
	limit, offset := req.ToGormParams()

	versions, total, err := repository.ListContainerVersions(database.DB, limit, offset, containerID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list container versions: %w", err)
	}

	versionResps := make([]dto.ContainerVersionResp, len(versions))
	for i, v := range versions {
		versionResps[i] = *dto.NewContainerVersionResp(&v)
	}

	resp := dto.ListResp[dto.ContainerVersionResp]{
		Items:      versionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateContainerVersion updates an existing container version's details
func UpdateContainerVersion(req *dto.UpdateContainerVersionReq, containerID, versionID int) (*dto.ContainerVersionResp, error) {
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
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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

	return dto.NewContainerVersionResp(updatedVersion), nil
}

// =====================================================================
// Container Building Task Service Layer
// =====================================================================

// ProduceContainerBuildingTask produces a container building task into Redis based on the provided request
func ProduceContainerBuildingTask(ctx context.Context, req *dto.SubmitBuildContainerReq, groupID string, userID int) (*dto.SubmitContainerBuildResp, error) {
	if req == nil {
		return nil, fmt.Errorf("build container request is nil")
	}

	sourcePath, err := processGitHubSource(req)
	if err != nil {
		return nil, fmt.Errorf("failed to process GitHub source: %w", err)
	}

	if err := req.ValidateInfoContent(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid container info content: %w", err)
	}
	if err := req.Options.ValidateRequiredFiles(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid container options: %w", err)
	}

	imageRef := fmt.Sprintf("%s/%s/%s:%s", config.GetString("harbor.registry"), config.GetString("harbor.namespace"), req.ImageName, req.Tag)
	payload := map[string]any{
		consts.BuildImageRef:     imageRef,
		consts.BuildSourcePath:   sourcePath,
		consts.BuildBuildOptions: req.Options,
	}

	task := &dto.UnifiedTask{
		Type:      consts.TaskTypeBuildContainer,
		Immediate: true,
		Payload:   payload,
		GroupID:   groupID,
		UserID:    userID,
		State:     consts.TaskPending,
	}
	task.SetGroupCtx(ctx)

	err = common.SubmitTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to submit container building task: %w", err)
	}

	resp := &dto.SubmitContainerBuildResp{
		GroupID: task.GroupID,
		TraceID: task.TraceID,
		TaskID:  task.TaskID,
	}
	return resp, nil
}

// createContainerVersionCore performs the core logic of creating container versions within a transaction
func createContainerVersionsCore(db *gorm.DB, versions []database.ContainerVersion) ([]database.ContainerVersion, error) {
	if len(versions) == 0 {
		return nil, nil
	}

	if err := repository.BatchCreateContainerVersions(db, versions); err != nil {
		return nil, fmt.Errorf("failed to create container versions: %w", err)
	}

	// Collect all envVars with their corresponding version index
	type envVarWithVersionIdx struct {
		envVar     database.ParameterConfig
		versionIdx int
		envVarIdx  int
	}

	envVarsWithIdx := []envVarWithVersionIdx{}
	for versionIdx, version := range versions {
		for envVarIdx, envVar := range version.EnvVars {
			envVarsWithIdx = append(envVarsWithIdx, envVarWithVersionIdx{
				envVar:     envVar,
				versionIdx: versionIdx,
				envVarIdx:  envVarIdx,
			})
		}
	}

	if len(envVarsWithIdx) > 0 {
		// Extract envVars for batch creation/upsert
		envVars := make([]database.ParameterConfig, len(envVarsWithIdx))
		for i, item := range envVarsWithIdx {
			envVars[i] = item.envVar
		}

		// Use OnConflict to insert or ignore existing configs
		if err := repository.BatchCreateOrFindParameterConfigs(db, envVars); err != nil {
			return nil, fmt.Errorf("failed to create parameter configs: %w", err)
		}

		// Query back the actual IDs from database (including existing ones)
		actualEnvVars, err := repository.ListParameterConfigsByKeys(db, envVars)
		if err != nil {
			return nil, fmt.Errorf("failed to list parameter configs: %w", err)
		}

		// Build a map for quick lookup: (key, type, category) -> ID
		configMap := make(map[string]int)
		for _, cfg := range actualEnvVars {
			key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
			configMap[key] = cfg.ID
		}

		// Build relations using the actual IDs from database
		relations := make([]database.ContainerVersionEnvVar, 0, len(envVarsWithIdx))
		for _, item := range envVarsWithIdx {
			cfg := item.envVar
			key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
			if paramID, exists := configMap[key]; exists {
				relations = append(relations, database.ContainerVersionEnvVar{
					ContainerVersionID: versions[item.versionIdx].ID,
					ParameterConfigID:  paramID,
				})
			} else {
				return nil, fmt.Errorf("parameter config not found after creation: %s", key)
			}
		}

		if err := repository.AddContainerVersionEnvVars(db, relations); err != nil {
			return nil, fmt.Errorf("failed to create container version env var relations: %w", err)
		}
	}

	var helmConfigs []database.HelmConfig
	helmConfigIdxMap := make(map[int]int) // map from helmConfig index to version index
	for versionIdx, version := range versions {
		if version.HelmConfig != nil {
			version.HelmConfig.ContainerVersionID = versions[versionIdx].ID
			helmConfigIdxMap[len(helmConfigs)] = versionIdx
			helmConfigs = append(helmConfigs, *version.HelmConfig)
		}
	}

	if len(helmConfigs) == 0 {
		return versions, nil
	}

	if err := repository.BatchCreateHelmConfigs(db, helmConfigs); err != nil {
		return nil, fmt.Errorf("failed to create helm configs: %w", err)
	}

	// Collect all helm values with their corresponding helmConfig index
	type helmValueWithConfigIdx struct {
		value         database.ParameterConfig
		helmConfigIdx int
		valueIdx      int
	}

	helmValuesWithIdx := []helmValueWithConfigIdx{}
	for helmConfigIdx, helmConfig := range helmConfigs {
		for valueIdx, value := range helmConfig.Values {
			helmValuesWithIdx = append(helmValuesWithIdx, helmValueWithConfigIdx{
				value:         value,
				helmConfigIdx: helmConfigIdx,
				valueIdx:      valueIdx,
			})
		}
	}

	if len(helmValuesWithIdx) > 0 {
		// Extract helm values for batch creation/upsert
		helmValues := make([]database.ParameterConfig, len(helmValuesWithIdx))
		for i, item := range helmValuesWithIdx {
			helmValues[i] = item.value
		}

		// Use OnConflict to insert or ignore existing configs
		if err := repository.BatchCreateOrFindParameterConfigs(db, helmValues); err != nil {
			return nil, fmt.Errorf("failed to create helm parameter configs: %w", err)
		}

		// Query back the actual IDs from database (including existing ones)
		actualHelmValues, err := repository.ListParameterConfigsByKeys(db, helmValues)
		if err != nil {
			return nil, fmt.Errorf("failed to list helm parameter configs: %w", err)
		}

		// Build a map for quick lookup: (key, type, category) -> ID
		configMap := make(map[string]int)
		for _, cfg := range actualHelmValues {
			key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
			configMap[key] = cfg.ID
		}

		// Build relations using the actual IDs from database
		relations := make([]database.HelmConfigValue, 0, len(helmValuesWithIdx))
		for _, item := range helmValuesWithIdx {
			cfg := item.value
			key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
			if paramID, exists := configMap[key]; exists {
				relations = append(relations, database.HelmConfigValue{
					HelmConfigID:      helmConfigs[item.helmConfigIdx].ID,
					ParameterConfigID: paramID,
				})
			} else {
				return nil, fmt.Errorf("helm parameter config not found after creation: %s", key)
			}
		}

		if err := repository.AddHelmConfigValues(db, relations); err != nil {
			return nil, fmt.Errorf("failed to create helm config value relations: %w", err)
		}
	}

	return versions, nil
}

// fetchContainersMapByIDBatch fetches containers by their IDs and returns a map of container ID to Container
func fetchContainersMapByIDBatch(db *gorm.DB, containerIDs []int) (map[int]database.Container, error) {
	if len(containerIDs) == 0 {
		return make(map[int]database.Container), nil
	}

	containers, err := repository.ListContainersByID(db, utils.ToUniqueSlice(containerIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to list containers by IDs: %w", err)
	}

	containerMap := make(map[int]database.Container, len(containers))
	for _, c := range containers {
		containerMap[c.ID] = c
	}

	return containerMap, nil
}

// processGitHubSource processes the GitHub source for building the container
func processGitHubSource(req *dto.SubmitBuildContainerReq) (string, error) {
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
