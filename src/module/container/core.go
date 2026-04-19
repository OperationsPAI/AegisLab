package container

import (
	"aegis/model"
)

func (r *Repository) CreateContainerCore(container *model.Container, userID int) (*model.Container, error) {
	service := NewService(r, NewBuildGateway(), NewHelmFileStore(), nil)
	return service.createContainerCore(r, container, userID)
}

func (r *Repository) UploadHelmValueFileFromPath(containerName string, helmConfig *model.HelmConfig, srcFilePath string) error {
	store := NewHelmFileStore()
	targetPath, err := store.SaveValueFile(containerName, nil, srcFilePath)
	if err != nil {
		return err
	}

	helmConfig.ValueFile = targetPath
	return r.updateHelmConfig(helmConfig)
}
