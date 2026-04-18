package datasetmodule

import (
	"aegis/model"
)

func (r *Repository) CreateDatasetCore(dataset *model.Dataset, versions []model.DatasetVersion, userID int) (*model.Dataset, error) {
	service := NewService(r, NewDatapackFileStore())
	return service.createDatasetCore(r, dataset, versions, userID)
}
