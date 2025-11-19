package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	datasetVersionOmitFields = "active_version_key"
)

// =====================================================================
// Dataset Repository Functions
// =====================================================================

// CreateDataset creates a new dataset record
func CreateDataset(db *gorm.DB, dataset *database.Dataset) error {
	if err := db.Omit(commonOmitFields).Create(dataset).Error; err != nil {
		return fmt.Errorf("failed to create dataset: %v", err)
	}
	return nil
}

// DeleteDataset soft deletes a dataset by setting its status to deleted
func DeleteDataset(db *gorm.DB, id int) (int64, error) {
	result := db.Model(&database.Dataset{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete dataset: %v", err)
	}
	return result.RowsAffected, nil
}

// GetDatasetByID gets dataset by ID
func GetDatasetByID(db *gorm.DB, id int) (*database.Dataset, error) {
	var dataset database.Dataset
	if err := db.Where("id = ? AND status != ?", id, consts.CommonDeleted).First(&dataset).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset: %v", err)
	}
	return &dataset, nil
}

// ListDatasets gets dataset list
func ListDatasets(db *gorm.DB, limit, offset int, datasetType string, isPublic *bool, status *consts.StatusType) ([]database.Dataset, int64, error) {
	var datasets []database.Dataset
	var total int64

	query := db.Model(&database.Dataset{})
	if datasetType != "" {
		query = query.Where("type = ?", datasetType)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count datasets: %v", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&datasets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list datasets: %v", err)
	}

	return datasets, total, nil
}

// ListDatasetsByID retrieves multiple datasets by their IDs
func ListDatasetsByID(db *gorm.DB, datasetIDs []int) ([]database.Dataset, error) {
	if len(datasetIDs) == 0 {
		return []database.Dataset{}, nil
	}

	var datasets []database.Dataset
	if err := db.
		Where("id IN (?) AND status != ?", datasetIDs, consts.CommonDeleted).
		Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("failed to query datasets: %w", err)
	}

	return datasets, nil
}

// UpdateDataset updates dataset information
func UpdateDataset(db *gorm.DB, dataset *database.Dataset) error {
	if err := db.Omit(commonOmitFields).Save(dataset).Error; err != nil {
		return fmt.Errorf("failed to update dataset: %v", err)
	}
	return nil
}

// GetDatasetStatistics returns statistics about datasets
func GetDatasetStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total datasets
	var total int64
	if err := database.DB.Model(&database.Dataset{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total datasets: %v", err)
	}
	stats["total"] = total

	// Active datasets
	var active int64
	if err := database.DB.Model(&database.Dataset{}).Where("status = ?", consts.DatapackInjectSuccess).Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active datasets: %v", err)
	}
	stats["active"] = active

	// Disabled datasets
	var disabled int64
	if err := database.DB.Model(&database.Dataset{}).Where("status = ?", consts.DatapackInitial).Count(&disabled).Error; err != nil {
		return nil, fmt.Errorf("failed to count disabled datasets: %v", err)
	}
	stats["disabled"] = disabled

	// Deleted datasets
	var deleted int64
	if err := database.DB.Model(&database.Dataset{}).Where("status = ?", consts.CommonDeleted).Count(&deleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count deleted datasets: %v", err)
	}
	stats["deleted"] = deleted

	return stats, nil
}

// =====================================================================
// DatasetVersion Repository Functions
// =====================================================================

// BatchCreateDatasetVersions creates multiple dataset versions
func BatchCreateDatasetVersions(db *gorm.DB, versions []database.DatasetVersion) error {
	if len(versions) == 0 {
		return fmt.Errorf("no dataset versions to create")
	}

	if err := db.Omit(datasetVersionOmitFields).Create(&versions).Error; err != nil {
		return fmt.Errorf("failed to batch create dataset versions: %w", err)
	}

	return nil
}

// BatchDeleteDatasetVersions soft deletes all versions of a specific dataset
func BatchDeleteDatasetVersions(db *gorm.DB, datasetID int) (int64, error) {
	result := db.Model(&database.DatasetVersion{}).
		Where("dataset_id = ? AND status != ?", datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to batch soft delete dataset versions for dataset %d: %w", datasetID, result.Error)
	}
	return result.RowsAffected, nil
}

// BatchGetDatasetVersions retrieves dataset versions for multiple dataset names
func BatchGetDatasetVersions(db *gorm.DB, datasetNames []string, userID int) ([]database.DatasetVersion, error) {
	if len(datasetNames) == 0 {
		return []database.DatasetVersion{}, nil
	}

	var versions []database.DatasetVersion

	query := db.Table("dataset_versions dv").
		Preload("Dataset").
		Where("dv.status = ?", consts.CommonEnabled).
		Order("dv.dataset_id DESC, dv.name_major DESC, dv.name_minor DESC, dv.name_patch DESC")

	query = query.Joins("INNER JOIN datasets d ON d.id = dv.dataset_id").
		Where("d.name IN (?) AND d.status = ?", datasetNames, consts.CommonEnabled)

	if userID > 0 {
		query = query.Joins(
			"LEFT JOIN user_datasets ud ON ud.dataset_id = d.id AND ud.user_id = ? AND ud.status = ?",
			userID, consts.CommonEnabled,
		).Where(
			db.Where("d.is_public = ?", true).
				Or("ud.dataset_id IS NOT NULL"),
		)
	}

	if err := query.Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to query dataset versions: %w", err)
	}

	return versions, nil
}

// DeleteDatasetVersion performs a soft delete on the dataset version by setting its status to  deleted
func DeleteDatasetVersion(db *gorm.DB, versionID int) (int64, error) {
	result := db.Model(&database.DatasetVersion{}).
		Where("id = ? AND status != ?", versionID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete dataset version %d: %w", versionID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetDatasetVersionByID retrieves a dataset version by its ID
func GetDatasetVersionByID(db *gorm.DB, id int) (*database.DatasetVersion, error) {
	var version database.DatasetVersion
	if err := db.Where("id = ?", id).First(&version).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset version: %v", err)
	}
	return &version, nil
}

// ListDatasetVersions lists dataset versions with pagination and optional status filtering
func ListDatasetVersions(db *gorm.DB, limit, offset int, datasetID int, status *consts.StatusType) ([]database.DatasetVersion, int64, error) {
	var versions []database.DatasetVersion
	var total int64

	query := db.Model(&database.DatasetVersion{}).Where("dataset_id = ?", datasetID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count dataset versions: %v", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list dataset versions: %v", err)
	}

	return versions, total, nil
}

// ListDatasetVersions lists all versions of a specific dataset
func ListDatasetVersionsByDatasetID(db *gorm.DB, datasetID int) ([]database.DatasetVersion, error) {
	var versions []database.DatasetVersion
	if err := db.Where("dataset_id = ?", datasetID).Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list dataset versions for dataset %d: %w", datasetID, err)
	}
	return versions, nil
}

// UpdateDatasetVersion updates a dataset version
func UpdateDatasetVersion(db *gorm.DB, version *database.DatasetVersion) error {
	if err := db.Omit(datasetVersionOmitFields).Save(version).Error; err != nil {
		return fmt.Errorf("failed to update dataset version: %w", err)
	}
	return nil
}

// =====================================================================
// DatasetLabel Repository Functions
// =====================================================================

// AddDatasetLabels adds multiple dataset-label associations in a batch
func AddDatasetLabels(db *gorm.DB, datasetLabels []database.DatasetLabel) error {
	if len(datasetLabels) == 0 {
		return nil
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "dataset_id"}, {Name: "label_id"}},
		DoNothing: true,
	}).Create(&datasetLabels).Error; err != nil {
		return fmt.Errorf("failed to add dataset-label associations: %w", err)
	}
	return nil
}

// ClearDatasetLabels removes label associations from specified datasets
func ClearDatasetLabels(db *gorm.DB, datasetIDs []int, labelIDs []int) error {
	if len(datasetIDs) == 0 {
		return nil
	}

	query := db.Table("dataset_labels").
		Where("dataset_id IN (?)", datasetIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}

	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear dataset-label associations: %w", err)
	}
	return nil
}

// RemoveLabelsFromDataset removes all label associations from a specific dataset
func RemoveLabelsFromDataset(db *gorm.DB, datasetID int) error {
	if err := db.Where("dataset_id = ?", datasetID).
		Delete(&database.DatasetLabel{}).Error; err != nil {
		return fmt.Errorf("failed to delete all labels from dataset %d: %w", datasetID, err)
	}
	return nil
}

// RemoveDatasetsFromLabel removes all dataset associations from a specific label
func RemoveDatasetsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	result := db.Where("label_id = ?", labelID).
		Delete(&database.DatasetLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all datasets from label %d: %w", labelID, err)
	}
	return result.RowsAffected, nil
}

// RemoveDatasetsFromLabels removes all dataset associations from multiple labels
func RemoveDatasetsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}

	result := db.Where("label_id IN (?)", labelIDs).
		Delete(&database.DatasetLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all datasets from labels %v: %w", labelIDs, err)
	}
	return result.RowsAffected, nil
}

// ListDatasetLabelCounts retrieves the count of datasets associated with each label ID
func ListDatasetLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return make(map[int]int64), nil
	}

	type datasetLabelResult struct {
		labelID int `gorm:"column:label_id"`
		count   int64
	}

	var results []datasetLabelResult
	if err := db.Model(&database.DatasetLabel{}).
		Select("label_id, count(label_id) as count").
		Where("label_id IN (?)", labelIDs).
		Group("label_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count dataset-label associations: %w", err)
	}

	countMap := make(map[int]int64, len(results))
	for _, result := range results {
		countMap[result.labelID] = result.count
	}

	return countMap, nil
}

// ListDatasetLabels lists all labels associated with multiple datasets
func ListDatasetLabels(db *gorm.DB, datasetIDs []int) (map[int][]database.Label, error) {
	if len(datasetIDs) == 0 {
		return nil, nil
	}

	type datasetLabelResult struct {
		database.Label
		datasetID int `gorm:"column:dataset_id"`
	}

	var flatResults []datasetLabelResult
	if err := db.Model(&database.Label{}).
		Joins("JOIN dataset_labels dl ON dl.label_id = labels.id").
		Where("dl.dataset_id IN (?)", datasetIDs).
		Select("labels.*, dl.dataset_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query dataset labels: %w", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, id := range datasetIDs {
		labelsMap[id] = []database.Label{}
	}

	for _, res := range flatResults {
		label := res.Label
		labelsMap[res.datasetID] = append(labelsMap[res.datasetID], label)
	}

	return labelsMap, nil
}

// ListLabelsByDatasetID lists all labels associated with a specific dataset
func ListLabelsByDatasetID(db *gorm.DB, datasetID int) ([]database.Label, error) {
	var labels []database.Label
	if err := db.Model(&database.Label{}).
		Joins("JOIN dataset_labels dl ON dl.label_id = labels.id").
		Where("dl.dataset_id = ?", datasetID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for dataset %d: %w", datasetID, err)
	}
	return labels, nil
}

// ListLabelIDsByKeyAndInjectionID finds label IDs by keys associated with a specific injection
func ListLabelIDsByKeyAndDatasetID(db *gorm.DB, datasetID int, keys []string) ([]int, error) {
	var labelIDs []int

	err := db.Table("labels l").
		Select("l.id").
		Joins("JOIN dataset_labels dl ON dl.label_id = l.id").
		Where("dl.dataset_id = ? AND l.label_key IN (?)", datasetID, keys).
		Pluck("l.id", &labelIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%s': %w", keys, err)
	}

	return labelIDs, nil
}

// =====================================================================
// DatasetVersionInjection Repository Functions
// =====================================================================

// AddDatasetVersionInjections adds multiple dataset-version-injection associations in a batch
func AddDatasetVersionInjections(db *gorm.DB, datasetVersionInjections []database.DatasetVersionInjection) error {
	if len(datasetVersionInjections) == 0 {
		return nil
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "dataset_version_id"}, {Name: "injection_id"}},
		DoNothing: true,
	}).Create(&datasetVersionInjections).Error; err != nil {
		return fmt.Errorf("failed to add dataset-version-injection associations: %w", err)
	}
	return nil
}

// ClearDatasetVersionInjections removes fault injection associations from specified dataset versions
func ClearDatasetVersionInjections(db *gorm.DB, datasetVersionIDs []int, injectionIDs []int) error {
	if len(datasetVersionIDs) == 0 {
		return nil
	}

	query := db.Table("dataset_version_injections").
		Where("dataset_version_id IN (?)", datasetVersionIDs)
	if len(injectionIDs) > 0 {
		query = query.Where("injection_id IN (?)", injectionIDs)
	}

	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear dataset-version-injection associations: %w", err)
	}
	return nil
}

// RemoveInjectionsFromDatasetVersion deletes all injection associations for a given dataset version
func RemoveInjectionsFromDatasetVersion(db *gorm.DB, datasetVersionID int) error {
	if err := db.Where("dataset_version_id = ?", datasetVersionID).
		Delete(&database.DatasetVersionInjection{}).Error; err != nil {
		return fmt.Errorf("failed to delete all injections from dataset version %d: %w", datasetVersionID, err)
	}
	return nil
}

// RemoveDatasetVersionsFromInjection deletes all dataset version associations for a given fault injection
func RemoveDatasetVersionsFromInjection(db *gorm.DB, faultInjectionID int) error {
	if err := db.Where("injection_id = ?", faultInjectionID).
		Delete(&database.DatasetVersionInjection{}).Error; err != nil {
		return fmt.Errorf("failed to delete all dataset versions from fault injection %d: %w", faultInjectionID, err)
	}
	return nil
}

// ListInjectionsByDatasetVersionID lists all fault injections associated with a specific dataset version
func ListInjectionsByDatasetVersionID(db *gorm.DB, datasetVersionID int) ([]database.FaultInjection, error) {
	var injections []database.FaultInjection
	if err := db.Model(&database.FaultInjection{}).
		Joins("JOIN dataset_version_injections dvi ON dvi.injection_id = fault_injections.id").
		Where("dvi.dataset_version_id = ?", datasetVersionID).
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to list fault injections for dataset version %d: %w", datasetVersionID, err)
	}
	return injections, nil
}
