package producer

import (
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"gorm.io/gorm"
)

// =====================================================================
// Dataset Service Layer
// =====================================================================

// CreateDataset creates a new dataset
func CreateDataset(req *dto.CreateDatasetReq, userID int) (*dto.DatasetResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	dataset := req.ConvertToDataset()

	var version *database.DatasetVersion
	if req.VersionReq != nil {
		version = req.VersionReq.ConvertToDatasetVersion()
	}

	var createdDataset *database.Dataset
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if version != nil {
			dataset, err = CreateDatasetCore(tx, dataset, []database.DatasetVersion{*version}, userID)
		} else {
			dataset, err = CreateDatasetCore(tx, dataset, nil, userID)
		}

		if err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}

		createdDataset = dataset
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	return dto.NewDatasetResp(createdDataset), nil
}

// CreateDatasetCore performs the core logic of creating a dataset within a transaction
func CreateDatasetCore(tx *gorm.DB, dataset *database.Dataset, versions []database.DatasetVersion, userID int) (*database.Dataset, error) {
	role, err := repository.GetRoleByName(tx, consts.RoleDatasetAdmin)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role %v not found", err, consts.RoleDatasetAdmin)
		}
		return nil, fmt.Errorf("failed to get dataset owner role: %w", err)
	}

	if err := repository.CreateDataset(tx, dataset); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, consts.ErrAlreadyExists
		}

		return nil, err
	}

	if err := repository.CreateUserDataset(tx, &database.UserDataset{
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

		_, err = createDatasetVersionsCore(tx, versions)
		if err != nil {
			return nil, fmt.Errorf("failed to create dataset versions: %w", err)
		}
	}

	return dataset, nil
}

func DeleteDataset(datasetID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := repository.BatchDeleteDatasetVersions(tx, datasetID); err != nil {
			return fmt.Errorf("failed to delete dataset versions: %w", err)
		}

		if _, err := repository.RemoveUsersFromDataset(tx, datasetID); err != nil {
			return fmt.Errorf("failed to remove all users from dataset: %w", err)
		}

		rows, err := repository.DeleteDataset(tx, datasetID)
		if err != nil {
			return fmt.Errorf("failed to delete dataset: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: dataset id %d not found", consts.ErrNotFound, datasetID)
		}

		return nil
	})
}

func GetDatasetDetail(datasetID int) (*dto.DatasetDetailResp, error) {
	dataset, err := repository.GetDatasetByID(database.DB, datasetID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
		}
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	versions, err := repository.ListDatasetVersionsByDatasetID(database.DB, dataset.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset versions: %w", err)
	}

	resp := dto.NewDatasetDetailResp(dataset)

	for _, version := range versions {
		resp.Versions = append(resp.Versions, *dto.NewDatasetVersionResp(&version))
	}

	return dto.NewDatasetDetailResp(dataset), nil
}

// ListDatasets lists datasets with pagination and optional filtering
func ListDatasets(req *dto.ListDatasetReq) (*dto.ListResp[dto.DatasetResp], error) {
	limit, offset := req.ToGormParams()

	datasets, total, err := repository.ListDatasets(database.DB, limit, offset, req.Type, req.IsPublic, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	datasetIDs := make([]int, 0, len(datasets))
	for _, d := range datasets {
		datasetIDs = append(datasetIDs, d.ID)
	}

	labelsMap, err := repository.ListDatasetLabels(database.DB, datasetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list dataset labels: %w", err)
	}

	datasetResps := make([]dto.DatasetResp, 0, len(datasets))
	for _, dataset := range datasets {
		if labels, exists := labelsMap[dataset.ID]; exists {
			dataset.Labels = labels
		}
		datasetResps = append(datasetResps, *dto.NewDatasetResp(&dataset))
	}

	resp := dto.ListResp[dto.DatasetResp]{
		Items:      datasetResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// SearchDataset searches datasets based on the provided search request
func SearchDatasets(req *dto.SearchDatasetReq) (*dto.ListResp[dto.DatasetDetailResp], error) {
	if req == nil {
		return nil, fmt.Errorf("search dataset request is nil")
	}

	searchReq := req.ConvertToSearchReq()
	dataests, total, err := repository.ExecuteSearch(database.DB, searchReq, database.Dataset{})
	if err != nil {
		return nil, fmt.Errorf("failed to search datasets: %w", err)
	}

	datasetResps := make([]dto.DatasetDetailResp, 0, len(dataests))
	for _, dataset := range dataests {
		datasetResps = append(datasetResps, *dto.NewDatasetDetailResp(&dataset))
	}

	resp := dto.ListResp[dto.DatasetDetailResp]{
		Items:      datasetResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

func UpdateDataset(req *dto.UpdateDatasetReq, datasetID int) (*dto.DatasetResp, error) {
	var updatedDataset *database.Dataset

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingDataset, err := repository.GetDatasetByID(tx, datasetID)
		if err != nil {
			return fmt.Errorf("failed to get dataset: %w", err)
		}

		req.PatchDatasetModel(existingDataset)

		if err := repository.UpdateDataset(tx, existingDataset); err != nil {
			return fmt.Errorf("failed to update dataset: %w", err)
		}

		updatedDataset = existingDataset
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewDatasetResp(updatedDataset), nil
}

// ===================== Dataset-Label =====================

func ManageDatasetLabels(req *dto.ManageDatasetLabelReq, datasetID int) (*dto.DatasetResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage dataset labels request is nil")
	}

	var managedDataset *database.Dataset
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		dataset, err := repository.GetDatasetByID(tx, datasetID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
			}
			return fmt.Errorf("failed to get dataset: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.DatasetCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			datasetLabels := make([]database.DatasetLabel, 0, len(labels))
			for _, label := range labels {
				datasetLabels = append(datasetLabels, database.DatasetLabel{
					DatasetID: datasetID,
					LabelID:   label.ID,
				})
			}

			if err := repository.AddDatasetLabels(tx, datasetLabels); err != nil {
				return fmt.Errorf("failed to add dataset labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndDatasetID(tx, datasetID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) > 0 {
				if err := repository.ClearDatasetLabels(tx, []int{datasetID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear dataset labels: %w", err)
				}

				if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		labels, err := repository.ListLabelsByDatasetID(database.DB, dataset.ID)
		if err != nil {
			return fmt.Errorf("failed to get dataset labels: %w", err)
		}

		dataset.Labels = labels
		managedDataset = dataset
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewDatasetResp(managedDataset), nil
}

// =====================================================================
// DatasetVersion Service Layer
// =====================================================================

func CreateDatasetVersion(req *dto.CreateDatasetVersionReq, datasetID, userID int) (*dto.DatasetVersionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("create dataset version request is nil")
	}

	version := req.ConvertToDatasetVersion()
	version.DatasetID = datasetID
	version.UserID = userID

	var createdVersion *database.DatasetVersion
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		versions, err := createDatasetVersionsCore(tx, []database.DatasetVersion{*version})
		if err != nil {
			return fmt.Errorf("failed to create dataset version: %w", err)
		}

		version := versions[0]
		if len(req.Datapacks) > 0 {
			if err := linkDatapacksToDatasetVersion(tx, version.ID, req.Datapacks); err != nil {
				return fmt.Errorf("failed to link datapacks to dataset version: %w", err)
			}
		}

		createdVersion = &version
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create dataset version: %w", err)
	}

	return dto.NewDatasetVersionResp(createdVersion), nil
}

// DeleteDatasetVersion deletes a specific version of a dataset
func DeleteDatasetVersion(versionID int) error {
	rows, err := repository.DeleteDatasetVersion(database.DB, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset version: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: dataset version id %d not found", consts.ErrNotFound, versionID)
	}
	return nil
}

// GetDatasetVersionDetail retrieves the details of a specific dataset version by its ID
func GetDatasetVersionDetail(datasetID, versionID int) (*dto.DatasetVersionDetailResp, error) {
	_, err := repository.GetDatasetByID(database.DB, datasetID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
		}
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	version, err := repository.GetDatasetVersionByID(database.DB, versionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return nil, fmt.Errorf("failed to get dataset version: %w", err)
	}

	return dto.NewDatasetVersionDetailResp(version), nil
}

// ListDatasetVersions lists dataset versions with pagination and optional status filtering
func ListDatasetVersions(req *dto.ListDatasetVersionReq, datasetID int) (*dto.ListResp[dto.DatasetVersionResp], error) {
	limit, offset := req.ToGormParams()

	versions, total, err := repository.ListDatasetVersions(database.DB, limit, offset, datasetID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list dataset versions: %w", err)
	}

	versionResps := make([]dto.DatasetVersionResp, 0, len(versions))
	for _, version := range versions {
		versionResps = append(versionResps, *dto.NewDatasetVersionResp(&version))
	}

	resp := dto.ListResp[dto.DatasetVersionResp]{
		Items:      versionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateDatasetVersion updates the details of a specific dataset version
func UpdateDatasetVersion(req *dto.UpdateDatasetVersionReq, datasetID, versionID int) (*dto.DatasetVersionResp, error) {
	var updatedVersion *database.DatasetVersion

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		version, err := repository.GetDatasetVersionByID(tx, versionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get dataset version: %w", err)
		}

		req.PatchDatasetVersionModel(version)

		if err := repository.UpdateDatasetVersion(tx, version); err != nil {
			return fmt.Errorf("failed to update dataset version: %w", err)
		}

		updatedVersion = version
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update dataset version: %w", err)
	}

	return dto.NewDatasetVersionResp(updatedVersion), nil
}

// GetFilename generates a filename for the dataset version download
func GetFilename(datasetID, versionID int) (string, error) {
	dataset, err := repository.GetDatasetByID(database.DB, datasetID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return "", fmt.Errorf("%w: dataset id: %d", consts.ErrNotFound, datasetID)
		}
		return "", fmt.Errorf("failed to get dataset: %w", err)
	}

	version, err := repository.GetDatasetVersionByID(database.DB, versionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return "", fmt.Errorf("%w: version id: %d", consts.ErrNotFound, versionID)
		}
		return "", fmt.Errorf("failed to get dataset version: %w", err)
	}

	return fmt.Sprintf("%s-%s", dataset.Name, version.Name), nil
}

// DownloadDatasetVersion handles the downloading of a specific dataset version
func DownloadDatasetVersion(zipWriter *zip.Writer, excludeRules []utils.ExculdeRule, versionID int) error {
	if zipWriter == nil {
		return fmt.Errorf("zip writer cannot be nil")
	}

	datapacks, err := repository.ListInjectionsByDatasetVersionID(database.DB, versionID, false)
	if err != nil {
		return fmt.Errorf("failed to list datapacks for dataset version: %w", err)
	}

	datapackNames := make([]string, 0, len(datapacks))
	for _, dp := range datapacks {
		datapackNames = append(datapackNames, dp.Name)
	}

	if err := packageDatasetToZip(zipWriter, datapackNames, excludeRules); err != nil {
		return fmt.Errorf("failed to package dataset to zip: %w", err)
	}

	return nil
}

// ===================== DatasetVersion-Injection =====================

func ManageDatasetVersionInjections(req *dto.ManageDatasetVersionInjectionReq, versionID int) (*dto.DatasetVersionDetailResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage dataset version injections request is nil")
	}

	var managedVersion *database.DatasetVersion
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		version, err := repository.GetDatasetVersionByID(tx, versionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: dataset version id: %d", consts.ErrNotFound, versionID)
			}
			return fmt.Errorf("failed to get dataset version: %w", err)
		}

		if len(req.AddDatapacks) > 0 {
			if err := linkDatapacksToDatasetVersion(tx, versionID, req.AddDatapacks); err != nil {
				return fmt.Errorf("failed to link datapacks to dataset version: %w", err)
			}
		}

		if len(req.RemoveDatapacks) > 0 {
			injectionIDMap, err := repository.ListInjectionIDsByNames(tx, req.AddDatapacks)
			if err != nil {
				return fmt.Errorf("failed to list injections by names: %w", err)
			}

			if len(injectionIDMap) != len(req.RemoveDatapacks) {
				return fmt.Errorf("some datapacks to remove were not found")
			}

			injectionIDs := make([]int, 0, len(req.RemoveDatapacks))
			for _, datapack := range req.RemoveDatapacks {
				injectionID, exists := injectionIDMap[datapack]
				if !exists {
					return fmt.Errorf("injection not found: %s", datapack)
				}
				injectionIDs = append(injectionIDs, injectionID)
			}

			if err := repository.ClearDatasetVersionInjections(tx, []int{version.ID}, injectionIDs); err != nil {
				return fmt.Errorf("failed to remove dataset version datapacks: %w", err)
			}
		}

		datapacks, err := repository.ListInjectionsByDatasetVersionID(tx, version.ID, false)
		if err != nil {
			return fmt.Errorf("failed to list datapacks for dataset version: %w", err)
		}

		version.Datapacks = datapacks
		version.FileCount = version.FileCount + len(req.AddDatapacks) - len(req.RemoveDatapacks)
		if err := repository.UpdateDatasetVersion(tx, version); err != nil {
			return fmt.Errorf("failed to update dataset version file count: %w", err)
		}

		managedVersion = version
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewDatasetVersionDetailResp(managedVersion), nil
}

// createDatasetVersionCore performs the core logic of creating dataset versions within a transaction
func createDatasetVersionsCore(db *gorm.DB, versions []database.DatasetVersion) ([]database.DatasetVersion, error) {
	if len(versions) == 0 {
		return nil, nil
	}

	if err := repository.BatchCreateDatasetVersions(db, versions); err != nil {
		return nil, fmt.Errorf("failed to create dataset versions: %w", err)
	}

	return versions, nil
}

// fetchDatasetsMapByIDBatch fetches datasets by their IDs and returns a map of dataset ID to Dataset
func fetchDatasetsMapByIDBatch(db *gorm.DB, datasetIDs []int) (map[int]database.Dataset, error) {
	if len(datasetIDs) == 0 {
		return make(map[int]database.Dataset), nil
	}

	datasets, err := repository.ListDatasetsByID(db, utils.ToUniqueSlice(datasetIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets by IDs: %w", err)
	}

	datasetMap := make(map[int]database.Dataset, len(datasets))
	for _, d := range datasets {
		datasetMap[d.ID] = d
	}

	return datasetMap, nil
}

// linkDatapacksToDatasetVersion links the specified datapacks to the given dataset version
func linkDatapacksToDatasetVersion(db *gorm.DB, versionID int, datapacks []string) error {
	injectionIDMap, err := repository.ListInjectionIDsByNames(db, datapacks)
	if err != nil {
		return fmt.Errorf("failed to list injections by names: %w", err)
	}

	datasetVersionInjections := make([]database.DatasetVersionInjection, 0, len(datapacks))
	for _, datapack := range datapacks {
		injectionID, exists := injectionIDMap[datapack]
		if !exists {
			return fmt.Errorf("injection not found: %s", datapack)
		}
		datasetVersionInjections = append(datasetVersionInjections, database.DatasetVersionInjection{
			DatasetVersionID: versionID,
			InjectionID:      injectionID,
		})
	}

	if err := repository.AddDatasetVersionInjections(db, datasetVersionInjections); err != nil {
		return fmt.Errorf("failed to add dataset version injections: %w", err)
	}

	return nil
}

// packageDatasetToZip packages the specified datapacks into a zip archive, applying exclusion rules
func packageDatasetToZip(zipWriter *zip.Writer, datapackNames []string, excludeRules []utils.ExculdeRule) error {
	for _, name := range datapackNames {
		workDir := filepath.Join(config.GetString("jfs.dataset_path"), name)
		if !utils.IsAllowedPath(workDir) {
			return fmt.Errorf("Invalid path access to %s", workDir)
		}

		err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
			if err != nil || dir.IsDir() {
				return err
			}

			relPath, _ := filepath.Rel(workDir, path)
			fullRelPath := filepath.Join(consts.DownloadFilename, filepath.Base(workDir), relPath)
			fileName := filepath.Base(path)

			// Apply exclusion rules
			for _, rule := range excludeRules {
				if utils.MatchFile(fileName, rule) {
					return nil
				}
			}

			// Get file info to read modification time
			fileInfo, err := dir.Info()
			if err != nil {
				return err
			}

			// Convert path separators to "/"
			zipPath := filepath.ToSlash(fullRelPath)
			return utils.AddToZip(zipWriter, fileInfo, path, zipPath)
		})
		if err != nil {
			return fmt.Errorf("failed to package: %w", err)
		}
	}

	return nil
}
