package containermodule

import (
	"aegis/model"

	"gorm.io/gorm"
)

func CreateContainerCore(tx *gorm.DB, container *model.Container, userID int) (*model.Container, error) {
	service := NewService(NewRepository(tx), NewBuildGateway(), NewHelmFileStore(), nil)
	return service.createContainerCore(service.repo, container, userID)
}

func UploadHelmValueFileFromPath(tx *gorm.DB, containerName string, helmConfig *model.HelmConfig, srcFilePath string) error {
	store := NewHelmFileStore()
	targetPath, err := store.SaveValueFile(containerName, nil, srcFilePath)
	if err != nil {
		return err
	}

	helmConfig.ValueFile = targetPath
	return NewRepository(tx).UpdateHelmConfig(helmConfig)
}
