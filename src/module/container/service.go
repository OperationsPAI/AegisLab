package container

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"aegis/consts"
	"aegis/dto"
	redis "aegis/infra/redis"
	"aegis/model"
	label "aegis/module/label"
	"aegis/service/common"

	"gorm.io/gorm"
)

type Service struct {
	repo      *Repository
	build     *BuildGateway
	helmFiles *HelmFileStore
	redis     *redis.Gateway
}

func NewService(repo *Repository, build *BuildGateway, helmFiles *HelmFileStore, redis *redis.Gateway) *Service {
	return &Service{repo: repo, build: build, helmFiles: helmFiles, redis: redis}
}

func (s *Service) CreateContainer(_ context.Context, req *CreateContainerReq, userID int) (*ContainerResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	container := req.ConvertToContainer()
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		createdContainer, err := s.createContainerCore(NewRepository(tx), container, userID)
		if err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}
		container = createdContainer
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return NewContainerResp(container), nil
}

func (s *Service) DeleteContainer(_ context.Context, containerID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if _, err := repo.batchDeleteContainerVersions(containerID); err != nil {
			return fmt.Errorf("failed to delete container versions: %w", err)
		}
		if _, err := repo.removeUsersFromContainer(containerID); err != nil {
			return fmt.Errorf("failed to remove all users from container: %w", err)
		}
		if err := repo.clearContainerLabels([]int{containerID}, nil); err != nil {
			return fmt.Errorf("failed to clear container labels: %w", err)
		}
		rows, err := repo.deleteContainer(containerID)
		if err != nil {
			return fmt.Errorf("failed to delete container: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: container id %d not found", consts.ErrNotFound, containerID)
		}
		return nil
	})
}

func (s *Service) GetContainer(_ context.Context, containerID int) (*ContainerDetailResp, error) {
	container, err := s.repo.getContainerByID(containerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: container id: %d", consts.ErrNotFound, containerID)
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	versions, err := s.repo.listContainerVersionsByContainerID(container.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container versions: %w", err)
	}

	resp := NewContainerDetailResp(container)
	for _, version := range versions {
		resp.Versions = append(resp.Versions, *NewContainerVersionResp(&version))
	}

	return resp, nil
}

func (s *Service) ListContainers(_ context.Context, req *ListContainerReq) (*dto.ListResp[ContainerResp], error) {
	limit, offset := req.ToGormParams()

	containers, total, err := s.repo.listContainers(limit, offset, req.Type, req.IsPublic, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := make([]int, 0, len(containers))
	for _, container := range containers {
		containerIDs = append(containerIDs, container.ID)
	}

	labelsMap, err := s.repo.listContainerLabels(containerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list container labels: %w", err)
	}

	items := make([]ContainerResp, 0, len(containers))
	for i := range containers {
		if labels, ok := labelsMap[containers[i].ID]; ok {
			containers[i].Labels = labels
		}
		items = append(items, *NewContainerResp(&containers[i]))
	}

	return &dto.ListResp[ContainerResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateContainer(_ context.Context, req *UpdateContainerReq, containerID int) (*ContainerResp, error) {
	var updatedContainer *model.Container

	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		container, err := repo.getContainerByID(containerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: container with id %d not found", consts.ErrNotFound, containerID)
			}
			return fmt.Errorf("failed to get container: %w", err)
		}

		req.PatchContainerModel(container)
		if err := repo.updateContainer(container); err != nil {
			return fmt.Errorf("failed to update container: %w", err)
		}

		updatedContainer = container
		return nil
	}); err != nil {
		return nil, err
	}

	return NewContainerResp(updatedContainer), nil
}

func (s *Service) ManageContainerLabels(_ context.Context, req *ManageContainerLabelReq, containerID int) (*ContainerResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var managedContainer *model.Container
	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		container, err := repo.getContainerByID(containerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: container not found", consts.ErrNotFound)
			}
			return err
		}

		if len(req.AddLabels) > 0 {
			labels, err := label.NewRepository(tx).CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.ContainerCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			containerLabels := make([]model.ContainerLabel, 0, len(labels))
			for _, label := range labels {
				containerLabels = append(containerLabels, model.ContainerLabel{
					ContainerID: containerID,
					LabelID:     label.ID,
				})
			}

			if err := repo.addContainerLabels(containerLabels); err != nil {
				return fmt.Errorf("failed to add container labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repo.listLabelIDsByKeyAndContainerID(containerID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label IDs: %w", err)
			}

			if len(labelIDs) > 0 {
				if err := repo.clearContainerLabels([]int{containerID}, labelIDs); err != nil {
					return fmt.Errorf("failed to delete container-label associations: %w", err)
				}

				if err := repo.batchDecreaseLabelUsages(labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		labels, err := repo.listLabelsByContainerID(container.ID)
		if err != nil {
			return fmt.Errorf("failed to get container labels: %w", err)
		}

		container.Labels = labels
		managedContainer = container
		return nil
	}); err != nil {
		return nil, err
	}

	return NewContainerResp(managedContainer), nil
}

func (s *Service) CreateContainerVersion(_ context.Context, req *CreateContainerVersionReq, containerID, userID int) (*ContainerVersionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("create container version request is nil")
	}

	version := req.ConvertToContainerVersion()
	version.ContainerID = containerID
	version.UserID = userID

	var createdVersion *model.ContainerVersion
	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		versions, err := s.createContainerVersionsCore(repo, []model.ContainerVersion{*version})
		if err != nil {
			return fmt.Errorf("failed to create container version: %w", err)
		}

		createdVersion = &versions[0]
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create container version: %w", err)
	}

	return NewContainerVersionResp(createdVersion), nil
}

func (s *Service) DeleteContainerVersion(_ context.Context, versionID int) error {
	rows, err := s.repo.deleteContainerVersion(versionID)
	if err != nil {
		return fmt.Errorf("failed to delete container version: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: container version id %d not found", consts.ErrNotFound, versionID)
	}
	return nil
}

func (s *Service) GetContainerVersion(_ context.Context, containerID, versionID int) (*ContainerVersionDetailResp, error) {
	if _, err := s.repo.getContainerByID(containerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: container id: %d", consts.ErrNotFound, containerID)
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	version, err := s.repo.getContainerVersionByID(versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return nil, fmt.Errorf("failed to get container version: %w", err)
	}

	resp := NewContainerVersionDetailResp(version)

	helmConfig, err := s.repo.getHelmConfigByContainerVersionID(version.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}
	if helmConfig != nil {
		helmConfigResp, err := NewHelmConfigDetailResp(helmConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to convert helm config: %w", err)
		}
		resp.HelmConfig = helmConfigResp
	}

	return resp, nil
}

func (s *Service) ListContainerVersions(_ context.Context, req *ListContainerVersionReq, containerID int) (*dto.ListResp[ContainerVersionResp], error) {
	limit, offset := req.ToGormParams()

	versions, total, err := s.repo.listContainerVersions(limit, offset, containerID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list container versions: %w", err)
	}

	items := make([]ContainerVersionResp, len(versions))
	for i := range versions {
		items[i] = *NewContainerVersionResp(&versions[i])
	}

	return &dto.ListResp[ContainerVersionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateContainerVersion(_ context.Context, req *UpdateContainerVersionReq, containerID, versionID int) (*ContainerVersionResp, error) {
	_ = containerID

	var updatedVersion *model.ContainerVersion
	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		version, err := repo.getContainerVersionByID(versionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get container version: %w", err)
		}

		req.PatchContainerVersionModel(version)
		if err := repo.updateContainerVersion(version); err != nil {
			return fmt.Errorf("failed to update container version: %w", err)
		}

		updatedVersion = version

		if req.HelmConfigRequest != nil {
			helmConfig, err := repo.getHelmConfigByContainerVersionID(version.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("helm config not found for version id %d", versionID)
				}
				return fmt.Errorf("failed to get helm config: %w", err)
			}

			if err := req.HelmConfigRequest.PatchHelmConfigModel(helmConfig); err != nil {
				return fmt.Errorf("failed to patch helm config model: %w", err)
			}
			if err := repo.updateHelmConfig(helmConfig); err != nil {
				return fmt.Errorf("failed to update helm config: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return NewContainerVersionResp(updatedVersion), nil
}

func (s *Service) SubmitContainerBuilding(ctx context.Context, req *SubmitBuildContainerReq, groupID string, userID int) (*SubmitContainerBuildResp, error) {
	if req == nil {
		return nil, fmt.Errorf("build container request is nil")
	}
	db := s.repo.db

	sourcePath, err := s.build.PrepareGitHubSource(req)
	if err != nil {
		return nil, fmt.Errorf("failed to process GitHub source: %w", err)
	}

	if err := req.ValidateInfoContent(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid container info content: %w", err)
	}
	if err := req.Options.ValidateRequiredFiles(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid container options: %w", err)
	}

	imageRef := s.build.BuildImageRef(req.ImageName, req.Tag)
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

	if err := common.SubmitTaskWithDB(ctx, db, s.redis, task); err != nil {
		return nil, fmt.Errorf("failed to submit container building task: %w", err)
	}

	return &SubmitContainerBuildResp{
		GroupID: task.GroupID,
		TraceID: task.TraceID,
		TaskID:  task.TaskID,
	}, nil
}

func (s *Service) UploadHelmChart(_ context.Context, file *multipart.FileHeader, containerID, versionID, userID int) (*UploadHelmChartResp, error) {
	_ = userID

	containerVersion, err := s.validateHelmConfigVersion(containerID, versionID)
	if err != nil {
		return nil, err
	}

	targetPath, checksum, err := s.helmFiles.SaveChart(containerVersion.Container.Name, file)
	if err != nil {
		return nil, err
	}
	filename := file.Filename
	containerVersion.HelmConfig.LocalPath = targetPath
	containerVersion.HelmConfig.Checksum = checksum
	if err := s.repo.updateHelmConfig(containerVersion.HelmConfig); err != nil {
		return nil, fmt.Errorf("failed to update helm config: %w", err)
	}

	return &UploadHelmChartResp{
		FilePath: targetPath,
		FileName: filename,
		Checksum: checksum,
	}, nil
}

func (s *Service) UploadHelmValueFile(_ context.Context, file *multipart.FileHeader, containerID, versionID, userID int) (*UploadHelmValueFileResp, error) {
	_ = userID

	containerVersion, err := s.validateHelmConfigVersion(containerID, versionID)
	if err != nil {
		return nil, err
	}

	if err := s.uploadHelmValueFileCore(containerVersion.Container.Name, containerVersion.HelmConfig, file, ""); err != nil {
		return nil, fmt.Errorf("failed to upload helm value file: %w", err)
	}

	return &UploadHelmValueFileResp{
		FilePath: containerVersion.HelmConfig.ValueFile,
		FileName: file.Filename,
	}, nil
}

func (s *Service) createContainerCore(repo *Repository, container *model.Container, userID int) (*model.Container, error) {
	role, err := repo.getRoleByName(consts.RoleContainerAdmin.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: role %v not found", consts.ErrNotFound, consts.RoleContainerAdmin)
		}
		return nil, fmt.Errorf("failed to get project owner role: %w", err)
	}

	if err := repo.createContainer(container); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, consts.ErrAlreadyExists
		}
		return nil, err
	}

	if err := repo.createUserContainer(&model.UserContainer{
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

		if _, err := s.createContainerVersionsCore(repo, container.Versions); err != nil {
			return nil, fmt.Errorf("failed to create container versions: %w", err)
		}
	}

	return container, nil
}

func (s *Service) createContainerVersionsCore(repo *Repository, versions []model.ContainerVersion) ([]model.ContainerVersion, error) {
	if len(versions) == 0 {
		return nil, nil
	}

	if err := repo.batchCreateContainerVersions(versions); err != nil {
		return nil, fmt.Errorf("failed to create container versions: %w", err)
	}

	type envVarWithVersionIdx struct {
		envVar     model.ParameterConfig
		versionIdx int
	}

	envVarsWithIdx := make([]envVarWithVersionIdx, 0)
	for versionIdx, version := range versions {
		for _, envVar := range version.EnvVars {
			envVarsWithIdx = append(envVarsWithIdx, envVarWithVersionIdx{
				envVar:     envVar,
				versionIdx: versionIdx,
			})
		}
	}

	if len(envVarsWithIdx) > 0 {
		envVars := make([]model.ParameterConfig, len(envVarsWithIdx))
		for i, item := range envVarsWithIdx {
			envVars[i] = item.envVar
		}

		if err := repo.batchCreateOrFindParameterConfigs(envVars); err != nil {
			return nil, fmt.Errorf("failed to create parameter configs: %w", err)
		}

		actualEnvVars, err := repo.listParameterConfigsByKeys(envVars)
		if err != nil {
			return nil, fmt.Errorf("failed to list parameter configs: %w", err)
		}

		configMap := make(map[string]int, len(actualEnvVars))
		for _, cfg := range actualEnvVars {
			key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
			configMap[key] = cfg.ID
		}

		relations := make([]model.ContainerVersionEnvVar, 0, len(envVarsWithIdx))
		for _, item := range envVarsWithIdx {
			cfg := item.envVar
			key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
			paramID, ok := configMap[key]
			if !ok {
				return nil, fmt.Errorf("parameter config not found after creation: %s", key)
			}
			relations = append(relations, model.ContainerVersionEnvVar{
				ContainerVersionID: versions[item.versionIdx].ID,
				ParameterConfigID:  paramID,
			})
		}

		if err := repo.addContainerVersionEnvVars(relations); err != nil {
			return nil, fmt.Errorf("failed to create container version env var relations: %w", err)
		}
	}

	helmConfigs := make([]*model.HelmConfig, 0)
	for versionIdx := range versions {
		if versions[versionIdx].HelmConfig != nil {
			versions[versionIdx].HelmConfig.ContainerVersionID = versions[versionIdx].ID
			helmConfigs = append(helmConfigs, versions[versionIdx].HelmConfig)
		}
	}

	if len(helmConfigs) == 0 {
		return versions, nil
	}

	if err := repo.batchCreateHelmConfigs(helmConfigs); err != nil {
		return nil, fmt.Errorf("failed to create helm configs: %w", err)
	}

	type helmValueWithConfigIdx struct {
		value         model.ParameterConfig
		helmConfigIdx int
	}

	helmValuesWithIdx := make([]helmValueWithConfigIdx, 0)
	for helmConfigIdx, helmConfig := range helmConfigs {
		for _, value := range helmConfig.DynamicValues {
			helmValuesWithIdx = append(helmValuesWithIdx, helmValueWithConfigIdx{
				value:         value,
				helmConfigIdx: helmConfigIdx,
			})
		}
	}

	if len(helmValuesWithIdx) == 0 {
		return versions, nil
	}

	helmValues := make([]model.ParameterConfig, len(helmValuesWithIdx))
	for i, item := range helmValuesWithIdx {
		helmValues[i] = item.value
	}

	if err := repo.batchCreateOrFindParameterConfigs(helmValues); err != nil {
		return nil, fmt.Errorf("failed to create helm parameter configs: %w", err)
	}

	actualHelmValues, err := repo.listParameterConfigsByKeys(helmValues)
	if err != nil {
		return nil, fmt.Errorf("failed to list helm parameter configs: %w", err)
	}

	configMap := make(map[string]int, len(actualHelmValues))
	for _, cfg := range actualHelmValues {
		key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
		configMap[key] = cfg.ID
	}

	relations := make([]model.HelmConfigValue, 0, len(helmValuesWithIdx))
	for _, item := range helmValuesWithIdx {
		cfg := item.value
		key := fmt.Sprintf("%s:%d:%d", cfg.Key, cfg.Type, cfg.Category)
		paramID, ok := configMap[key]
		if !ok {
			return nil, fmt.Errorf("helm parameter config not found after creation: %s", key)
		}
		relations = append(relations, model.HelmConfigValue{
			HelmConfigID:      helmConfigs[item.helmConfigIdx].ID,
			ParameterConfigID: paramID,
		})
	}

	if err := repo.addHelmConfigValues(relations); err != nil {
		return nil, fmt.Errorf("failed to create helm config value relations: %w", err)
	}

	return versions, nil
}

func (s *Service) uploadHelmValueFileCore(containerName string, helmConfig *model.HelmConfig, srcFileHeader *multipart.FileHeader, srcFilePath string) error {
	targetPath, err := s.helmFiles.SaveValueFile(containerName, srcFileHeader, srcFilePath)
	if err != nil {
		return err
	}
	helmConfig.ValueFile = targetPath
	if err := s.repo.updateHelmConfig(helmConfig); err != nil {
		return fmt.Errorf("failed to update helm config: %w", err)
	}

	return nil
}
func (s *Service) validateHelmConfigVersion(containerID, versionID int) (*model.ContainerVersion, error) {
	containerVersion, err := s.repo.getContainerVersionByID(versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: container version %d not found", consts.ErrNotFound, versionID)
		}
		return nil, fmt.Errorf("failed to get container version: %w", err)
	}

	if containerVersion.ContainerID != containerID {
		return nil, fmt.Errorf("version %d does not belong to container %d", versionID, containerID)
	}
	if containerVersion.Container == nil || containerVersion.Container.Type != consts.ContainerTypePedestal {
		return nil, fmt.Errorf("only pedestal container versions support Helm configurations")
	}
	if containerVersion.HelmConfig == nil {
		return nil, fmt.Errorf("container version %d does not have an associated Helm configuration", versionID)
	}

	return containerVersion, nil
}
