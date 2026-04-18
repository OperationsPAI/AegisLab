package datasetmodule

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	labelmodule "aegis/module/label"
	"aegis/utils"

	"gorm.io/gorm"
)

type Service struct {
	repo      *Repository
	datapacks *DatapackFileStore
}

func NewService(repo *Repository, datapacks *DatapackFileStore) *Service {
	return &Service{repo: repo, datapacks: datapacks}
}

func (s *Service) CreateDataset(_ context.Context, req *CreateDatasetReq, userID int) (*DatasetResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	dataset := req.ConvertToDataset()
	var versions []model.DatasetVersion
	if req.VersionReq != nil {
		versions = append(versions, *req.VersionReq.ConvertToDatasetVersion())
	}

	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		createdDataset, err := s.createDatasetCore(repo, dataset, versions, userID)
		if err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}
		dataset = createdDataset
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	return NewDatasetResp(dataset), nil
}

func (s *Service) DeleteDataset(_ context.Context, datasetID int) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		if _, err := repo.batchDeleteDatasetVersions(datasetID); err != nil {
			return fmt.Errorf("failed to delete dataset versions: %w", err)
		}
		if _, err := repo.removeUsersFromDataset(datasetID); err != nil {
			return fmt.Errorf("failed to remove all users from dataset: %w", err)
		}
		rows, err := repo.deleteDataset(datasetID)
		if err != nil {
			return fmt.Errorf("failed to delete dataset: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: dataset id %d not found", consts.ErrNotFound, datasetID)
		}
		return nil
	})
}

func (s *Service) GetDataset(_ context.Context, datasetID int) (*DatasetDetailResp, error) {
	dataset, err := s.repo.getDatasetByID(datasetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
		}
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	versions, err := s.repo.listDatasetVersionsByDatasetID(dataset.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset versions: %w", err)
	}

	resp := NewDatasetDetailResp(dataset)
	for _, version := range versions {
		resp.Versions = append(resp.Versions, *NewDatasetVersionResp(&version))
	}

	return resp, nil
}

func (s *Service) ListDatasets(_ context.Context, req *ListDatasetReq) (*dto.ListResp[DatasetResp], error) {
	limit, offset := req.ToGormParams()

	datasets, total, err := s.repo.listDatasets(limit, offset, req.Type, req.IsPublic, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	datasetIDs := make([]int, 0, len(datasets))
	for _, dataset := range datasets {
		datasetIDs = append(datasetIDs, dataset.ID)
	}

	labelsMap, err := s.repo.listDatasetLabels(datasetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list dataset labels: %w", err)
	}

	items := make([]DatasetResp, 0, len(datasets))
	for i := range datasets {
		if labels, ok := labelsMap[datasets[i].ID]; ok {
			datasets[i].Labels = labels
		}
		items = append(items, *NewDatasetResp(&datasets[i]))
	}

	return &dto.ListResp[DatasetResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) SearchDatasets(_ context.Context, req *SearchDatasetReq) (*dto.ListResp[DatasetDetailResp], error) {
	if req == nil {
		return nil, fmt.Errorf("search dataset request is nil")
	}

	results, total, err := s.repo.searchDatasets(req.ConvertToSearchReq())
	if err != nil {
		return nil, fmt.Errorf("failed to search datasets: %w", err)
	}

	items := make([]DatasetDetailResp, 0, len(results))
	for i := range results {
		items = append(items, *NewDatasetDetailResp(&results[i]))
	}

	return &dto.ListResp[DatasetDetailResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateDataset(_ context.Context, req *UpdateDatasetReq, datasetID int) (*DatasetResp, error) {
	var updatedDataset *model.Dataset

	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		dataset, err := repo.getDatasetByID(datasetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
			}
			return fmt.Errorf("failed to get dataset: %w", err)
		}

		req.PatchDatasetModel(dataset)
		if err := repo.updateDataset(dataset); err != nil {
			return fmt.Errorf("failed to update dataset: %w", err)
		}

		updatedDataset = dataset
		return nil
	}); err != nil {
		return nil, err
	}

	return NewDatasetResp(updatedDataset), nil
}

func (s *Service) ManageDatasetLabels(_ context.Context, req *ManageDatasetLabelReq, datasetID int) (*DatasetResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage dataset labels request is nil")
	}

	var managedDataset *model.Dataset
	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		dataset, err := repo.getDatasetByID(datasetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
			}
			return fmt.Errorf("failed to get dataset: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := labelmodule.NewRepository(tx).CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.DatasetCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			datasetLabels := make([]model.DatasetLabel, 0, len(labels))
			for _, label := range labels {
				datasetLabels = append(datasetLabels, model.DatasetLabel{
					DatasetID: datasetID,
					LabelID:   label.ID,
				})
			}

			if err := repo.addDatasetLabels(datasetLabels); err != nil {
				return fmt.Errorf("failed to add dataset labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repo.listLabelIDsByKeyAndDatasetID(datasetID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) > 0 {
				if err := repo.clearDatasetLabels([]int{datasetID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear dataset labels: %w", err)
				}

				if err := repo.batchDecreaseLabelUsages(labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		labels, err := repo.listLabelsByDatasetID(dataset.ID)
		if err != nil {
			return fmt.Errorf("failed to get dataset labels: %w", err)
		}

		dataset.Labels = labels
		managedDataset = dataset
		return nil
	}); err != nil {
		return nil, err
	}

	return NewDatasetResp(managedDataset), nil
}

func (s *Service) CreateDatasetVersion(_ context.Context, req *CreateDatasetVersionReq, datasetID, userID int) (*DatasetVersionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("create dataset version request is nil")
	}

	version := req.ConvertToDatasetVersion()
	version.DatasetID = datasetID
	version.UserID = userID

	var createdVersion *model.DatasetVersion
	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		versions, err := s.createDatasetVersionsCore(repo, []model.DatasetVersion{*version})
		if err != nil {
			return fmt.Errorf("failed to create dataset version: %w", err)
		}

		version := versions[0]
		if len(req.Datapacks) > 0 {
			if err := s.linkDatapacksToDatasetVersion(repo, version.ID, req.Datapacks); err != nil {
				return fmt.Errorf("failed to link datapacks to dataset version: %w", err)
			}
		}

		createdVersion = &version
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create dataset version: %w", err)
	}

	return NewDatasetVersionResp(createdVersion), nil
}

func (s *Service) DeleteDatasetVersion(_ context.Context, versionID int) error {
	rows, err := s.repo.deleteDatasetVersion(versionID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset version: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: dataset version id %d not found", consts.ErrNotFound, versionID)
	}
	return nil
}

func (s *Service) GetDatasetVersion(_ context.Context, datasetID, versionID int) (*DatasetVersionDetailResp, error) {
	if _, err := s.repo.getDatasetByID(datasetID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
		}
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	version, err := s.repo.getDatasetVersionByID(versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return nil, fmt.Errorf("failed to get dataset version: %w", err)
	}

	return NewDatasetVersionDetailResp(version), nil
}

func (s *Service) ListDatasetVersions(_ context.Context, req *ListDatasetVersionReq, datasetID int) (*dto.ListResp[DatasetVersionResp], error) {
	limit, offset := req.ToGormParams()

	versions, total, err := s.repo.listDatasetVersions(limit, offset, datasetID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list dataset versions: %w", err)
	}

	items := make([]DatasetVersionResp, 0, len(versions))
	for i := range versions {
		items = append(items, *NewDatasetVersionResp(&versions[i]))
	}

	return &dto.ListResp[DatasetVersionResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) UpdateDatasetVersion(_ context.Context, req *UpdateDatasetVersionReq, datasetID, versionID int) (*DatasetVersionResp, error) {
	_ = datasetID

	var updatedVersion *model.DatasetVersion
	if err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		version, err := repo.getDatasetVersionByID(versionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get dataset version: %w", err)
		}

		req.PatchDatasetVersionModel(version)
		if err := repo.updateDatasetVersion(version); err != nil {
			return fmt.Errorf("failed to update dataset version: %w", err)
		}

		updatedVersion = version
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to update dataset version: %w", err)
	}

	return NewDatasetVersionResp(updatedVersion), nil
}

func (s *Service) GetDatasetVersionFilename(_ context.Context, datasetID, versionID int) (string, error) {
	dataset, err := s.repo.getDatasetByID(datasetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
		}
		return "", fmt.Errorf("failed to get dataset: %w", err)
	}

	version, err := s.repo.getDatasetVersionByID(versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return "", fmt.Errorf("failed to get dataset version: %w", err)
	}

	return fmt.Sprintf("%s-%s", dataset.Name, version.Name), nil
}

func (s *Service) DownloadDatasetVersion(_ context.Context, zipWriter *zip.Writer, excludeRules []utils.ExculdeRule, versionID int) error {
	if zipWriter == nil {
		return fmt.Errorf("zip writer cannot be nil")
	}

	datapacks, err := s.repo.ListInjectionsByDatasetVersionID(versionID, false)
	if err != nil {
		return fmt.Errorf("failed to list datapacks for dataset version: %w", err)
	}

	if err := s.datapacks.PackageToZip(zipWriter, datapacks, excludeRules); err != nil {
		return fmt.Errorf("failed to package dataset to zip: %w", err)
	}

	return nil
}

func (s *Service) ManageDatasetVersionInjections(_ context.Context, req *ManageDatasetVersionInjectionReq, versionID int) (*DatasetVersionDetailResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage dataset version injections request is nil")
	}

	var managedVersion *model.DatasetVersion
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		version, err := repo.getDatasetVersionByID(versionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: dataset version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get dataset version: %w", err)
		}

		if len(req.AddDatapacks) > 0 {
			if err := s.linkDatapacksToDatasetVersion(repo, versionID, req.AddDatapacks); err != nil {
				return fmt.Errorf("failed to link datapacks to dataset version: %w", err)
			}
		}

		if len(req.RemoveDatapacks) > 0 {
			injectionIDMap, err := repo.listInjectionIDsByNames(req.RemoveDatapacks)
			if err != nil {
				return fmt.Errorf("failed to list injections by names: %w", err)
			}
			if len(injectionIDMap) != len(req.RemoveDatapacks) {
				return fmt.Errorf("some datapacks to remove were not found")
			}

			injectionIDs := make([]int, 0, len(req.RemoveDatapacks))
			for _, datapack := range req.RemoveDatapacks {
				injectionID, ok := injectionIDMap[datapack]
				if !ok {
					return fmt.Errorf("injection not found: %s", datapack)
				}
				injectionIDs = append(injectionIDs, injectionID)
			}

			if err := repo.clearDatasetVersionInjections([]int{version.ID}, injectionIDs); err != nil {
				return fmt.Errorf("failed to remove dataset version datapacks: %w", err)
			}
		}

		datapacks, err := repo.ListInjectionsByDatasetVersionID(version.ID, false)
		if err != nil {
			return fmt.Errorf("failed to list datapacks for dataset version: %w", err)
		}

		version.Datapacks = datapacks
		version.FileCount = version.FileCount + len(req.AddDatapacks) - len(req.RemoveDatapacks)
		if err := repo.updateDatasetVersion(version); err != nil {
			return fmt.Errorf("failed to update dataset version file count: %w", err)
		}

		managedVersion = version
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewDatasetVersionDetailResp(managedVersion), nil
}

func (s *Service) createDatasetCore(repo *Repository, dataset *model.Dataset, versions []model.DatasetVersion, userID int) (*model.Dataset, error) {
	role, err := repo.getRoleByName(consts.RoleDatasetAdmin.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: role %v not found", consts.ErrNotFound, consts.RoleDatasetAdmin)
		}
		return nil, fmt.Errorf("failed to get dataset owner role: %w", err)
	}

	if err := repo.createDataset(dataset); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, consts.ErrAlreadyExists
		}
		return nil, err
	}

	if err := repo.createUserDataset(&model.UserDataset{
		UserID:    userID,
		DatasetID: dataset.ID,
		RoleID:    role.ID,
		Status:    consts.CommonEnabled,
	}); err != nil {
		return nil, fmt.Errorf("failed to associate dataset with user: %w", err)
	}

	if len(versions) > 0 {
		for i := range versions {
			versions[i].DatasetID = dataset.ID
			versions[i].UserID = userID
		}

		if _, err := s.createDatasetVersionsCore(repo, versions); err != nil {
			return nil, fmt.Errorf("failed to create dataset versions: %w", err)
		}
	}

	return dataset, nil
}

func (s *Service) createDatasetVersionsCore(repo *Repository, versions []model.DatasetVersion) ([]model.DatasetVersion, error) {
	if len(versions) == 0 {
		return nil, nil
	}

	if err := repo.batchCreateDatasetVersions(versions); err != nil {
		return nil, fmt.Errorf("failed to create dataset versions: %w", err)
	}

	return versions, nil
}

func (s *Service) linkDatapacksToDatasetVersion(repo *Repository, versionID int, datapacks []string) error {
	injectionIDMap, err := repo.listInjectionIDsByNames(datapacks)
	if err != nil {
		return fmt.Errorf("failed to list injections by names: %w", err)
	}

	items := make([]model.DatasetVersionInjection, 0, len(datapacks))
	for _, datapack := range datapacks {
		injectionID, ok := injectionIDMap[datapack]
		if !ok {
			return fmt.Errorf("injection not found: %s", datapack)
		}
		items = append(items, model.DatasetVersionInjection{
			DatasetVersionID: versionID,
			InjectionID:      injectionID,
		})
	}

	if err := repo.addDatasetVersionInjections(items); err != nil {
		return fmt.Errorf("failed to add dataset version injections: %w", err)
	}

	return nil
}
