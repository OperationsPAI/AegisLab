package datasetmodule

import (
	"aegis/model"

	"gorm.io/gorm"
)

func CreateDatasetCore(tx *gorm.DB, dataset *model.Dataset, versions []model.DatasetVersion, userID int) (*model.Dataset, error) {
	service := NewService(NewRepository(tx), NewDatapackFileStore())
	return service.createDatasetCore(service.repo, dataset, versions, userID)
}
