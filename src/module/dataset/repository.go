package dataset

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/searchx"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	datasetCommonOmitFields       = "active_name"
	datasetVersionModelOmitFields = "active_version_key"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) getRoleByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ? and status != ?", name, consts.CommonDeleted).First(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to find role with name %s: %w", name, err)
	}
	return &role, nil
}

func (r *Repository) createDataset(dataset *model.Dataset) error {
	if err := r.db.Omit(datasetCommonOmitFields).Create(dataset).Error; err != nil {
		return fmt.Errorf("failed to create dataset: %v", err)
	}
	return nil
}

func (r *Repository) createUserDataset(userDataset *model.UserDataset) error {
	if err := r.db.Omit("active_user_dataset").Create(userDataset).Error; err != nil {
		return fmt.Errorf("failed to create user-dataset association: %w", err)
	}
	return nil
}

func (r *Repository) batchDeleteDatasetVersions(datasetID int) (int64, error) {
	result := r.db.Model(&model.DatasetVersion{}).
		Where("dataset_id = ? AND status != ?", datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to batch soft delete dataset versions for dataset %d: %w", datasetID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) removeUsersFromDataset(datasetID int) (int64, error) {
	result := r.db.Model(&model.UserDataset{}).
		Where("dataset_id = ? AND status != ?", datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete user-dataset associations for dataset %d: %w", datasetID, err)
	}
	return result.RowsAffected, nil
}

func (r *Repository) deleteDataset(datasetID int) (int64, error) {
	result := r.db.Model(&model.Dataset{}).
		Where("id = ? AND status != ?", datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete dataset: %v", err)
	}
	return result.RowsAffected, nil
}

func (r *Repository) getDatasetByID(datasetID int) (*model.Dataset, error) {
	var dataset model.Dataset
	if err := r.db.Where("id = ? AND status != ?", datasetID, consts.CommonDeleted).First(&dataset).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset: %v", err)
	}
	return &dataset, nil
}

func (r *Repository) listDatasetVersionsByDatasetID(datasetID int) ([]model.DatasetVersion, error) {
	var versions []model.DatasetVersion
	if err := r.db.Where("dataset_id = ?", datasetID).Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list dataset versions for dataset %d: %w", datasetID, err)
	}
	return versions, nil
}

func (r *Repository) batchGetDatasetVersions(datasetNames []string, userID int) ([]model.DatasetVersion, error) {
	if len(datasetNames) == 0 {
		return []model.DatasetVersion{}, nil
	}

	var versions []model.DatasetVersion
	query := r.db.Table("dataset_versions dv").
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
			r.db.Where("d.is_public = ?", true).Or("ud.dataset_id IS NOT NULL"),
		)
	}

	if err := query.Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to query dataset versions: %w", err)
	}
	return versions, nil
}

func (r *Repository) listDatasets(limit, offset int, datasetType string, isPublic *bool, status *consts.StatusType) ([]model.Dataset, int64, error) {
	var (
		datasets []model.Dataset
		total    int64
	)

	query := r.db.Model(&model.Dataset{})
	if datasetType != "" {
		query = query.Where("type = ?", datasetType)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count datasets: %v", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&datasets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list datasets: %v", err)
	}
	return datasets, total, nil
}

func (r *Repository) searchDatasets(searchReq *dto.SearchReq[consts.DatasetField]) ([]model.Dataset, int64, error) {
	qb := searchx.NewQueryBuilder(r.db, consts.DatasetAllowedFields)
	qb.ApplySearchReq(searchReq.Filters, searchReq.Keyword, searchReq.Sort, searchReq.GroupBy, model.Dataset{})
	qb.ApplyIncludes(searchReq.Includes)
	qb.ApplyIncludeFields(searchReq.IncludeFields)
	qb.ApplyExcludeFields(searchReq.ExcludeFields, model.Dataset{})

	total, err := qb.GetCount()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count searched datasets: %w", err)
	}

	query := qb.Query()
	if searchReq.Size != 0 && searchReq.Page != 0 {
		query = query.Offset(searchReq.GetOffset()).Limit(int(searchReq.Size))
	}

	var items []model.Dataset
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to execute dataset search: %w", err)
	}
	return items, total, nil
}

func (r *Repository) listDatasetLabels(datasetIDs []int) (map[int][]model.Label, error) {
	if len(datasetIDs) == 0 {
		return nil, nil
	}

	type datasetLabelResult struct {
		model.Label
		DatasetID int `gorm:"column:dataset_id"`
	}

	var flatResults []datasetLabelResult
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN dataset_labels dl ON dl.label_id = labels.id").
		Where("dl.dataset_id IN (?)", datasetIDs).
		Select("labels.*, dl.dataset_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query dataset labels: %w", err)
	}

	labelsMap := make(map[int][]model.Label, len(datasetIDs))
	for _, id := range datasetIDs {
		labelsMap[id] = []model.Label{}
	}
	for _, res := range flatResults {
		labelsMap[res.DatasetID] = append(labelsMap[res.DatasetID], res.Label)
	}
	return labelsMap, nil
}

func (r *Repository) updateDataset(dataset *model.Dataset) error {
	if err := r.db.Omit(datasetCommonOmitFields).Save(dataset).Error; err != nil {
		return fmt.Errorf("failed to update dataset: %v", err)
	}
	return nil
}

func (r *Repository) addDatasetLabels(datasetLabels []model.DatasetLabel) error {
	if len(datasetLabels) == 0 {
		return nil
	}
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "dataset_id"}, {Name: "label_id"}},
		DoNothing: true,
	}).Create(&datasetLabels).Error; err != nil {
		return fmt.Errorf("failed to add dataset-label associations: %w", err)
	}
	return nil
}

func (r *Repository) listLabelIDsByKeyAndDatasetID(datasetID int, keys []string) ([]int, error) {
	var labelIDs []int
	if err := r.db.Table("labels l").
		Select("l.id").
		Joins("JOIN dataset_labels dl ON dl.label_id = l.id").
		Where("dl.dataset_id = ? AND l.label_key IN (?)", datasetID, keys).
		Pluck("l.id", &labelIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%s': %w", keys, err)
	}
	return labelIDs, nil
}

func (r *Repository) clearDatasetLabels(datasetIDs []int, labelIDs []int) error {
	if len(datasetIDs) == 0 {
		return nil
	}

	query := r.db.Table("dataset_labels").Where("dataset_id IN (?)", datasetIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}
	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear dataset-label associations: %w", err)
	}
	return nil
}

func (r *Repository) batchDecreaseLabelUsages(labelIDs []int, decrement int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := gorm.Expr("GREATEST(0, usage_count - ?)", decrement)
	if err := r.db.Model(&model.Label{}).
		Where("id IN (?)", labelIDs).
		Clauses(clause.Returning{}).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

func (r *Repository) listLabelsByDatasetID(datasetID int) ([]model.Label, error) {
	var labels []model.Label
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN dataset_labels dl ON dl.label_id = labels.id").
		Where("dl.dataset_id = ?", datasetID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for dataset %d: %w", datasetID, err)
	}
	return labels, nil
}

func (r *Repository) batchCreateDatasetVersions(versions []model.DatasetVersion) error {
	if len(versions) == 0 {
		return fmt.Errorf("no dataset versions to create")
	}
	if err := r.db.Omit(datasetVersionModelOmitFields).Create(&versions).Error; err != nil {
		return fmt.Errorf("failed to batch create dataset versions: %w", err)
	}
	return nil
}

func (r *Repository) deleteDatasetVersion(versionID int) (int64, error) {
	result := r.db.Model(&model.DatasetVersion{}).
		Where("id = ? AND status != ?", versionID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete dataset version %d: %w", versionID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) getDatasetVersionByID(versionID int) (*model.DatasetVersion, error) {
	var version model.DatasetVersion
	if err := r.db.Preload("Datapacks").Where("id = ?", versionID).First(&version).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset version: %v", err)
	}
	return &version, nil
}

func (r *Repository) listDatasetVersions(limit, offset int, datasetID int, status *consts.StatusType) ([]model.DatasetVersion, int64, error) {
	var (
		versions []model.DatasetVersion
		total    int64
	)

	query := r.db.Model(&model.DatasetVersion{}).Where("dataset_id = ?", datasetID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count dataset versions: %v", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list dataset versions: %v", err)
	}
	return versions, total, nil
}

func (r *Repository) updateDatasetVersion(version *model.DatasetVersion) error {
	if err := r.db.Omit(datasetVersionModelOmitFields).Save(version).Error; err != nil {
		return fmt.Errorf("failed to update dataset version: %w", err)
	}
	return nil
}

func (r *Repository) listInjectionIDsByNames(names []string) (map[string]int, error) {
	if len(names) == 0 {
		return map[string]int{}, nil
	}

	var records []struct {
		Name string `gorm:"column:name"`
		ID   int    `gorm:"column:id"`
	}
	if err := r.db.Model(&model.FaultInjection{}).
		Select("name, id").
		Where("state = ? AND status = ?", consts.DatapackBuildSuccess, consts.CommonEnabled).
		Where("name IN (?)", names).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query injection IDs: %w", err)
	}

	result := make(map[string]int, len(records))
	for _, record := range records {
		result[record.Name] = record.ID
	}
	return result, nil
}

func (r *Repository) addDatasetVersionInjections(items []model.DatasetVersionInjection) error {
	if len(items) == 0 {
		return nil
	}
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "dataset_version_id"}, {Name: "injection_id"}},
		DoNothing: true,
	}).Create(&items).Error; err != nil {
		return fmt.Errorf("failed to add dataset-version-injection associations: %w", err)
	}
	return nil
}

func (r *Repository) clearDatasetVersionInjections(datasetVersionIDs []int, injectionIDs []int) error {
	if len(datasetVersionIDs) == 0 {
		return nil
	}

	query := r.db.Table("dataset_version_injections").Where("dataset_version_id IN (?)", datasetVersionIDs)
	if len(injectionIDs) > 0 {
		query = query.Where("injection_id IN (?)", injectionIDs)
	}
	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear dataset-version-injection associations: %w", err)
	}
	return nil
}

func (r *Repository) ListInjectionsByDatasetVersionID(versionID int, includeLabels bool) ([]model.FaultInjection, error) {
	query := r.db.Model(&model.FaultInjection{})
	if includeLabels {
		query = query.Preload("Labels")
	}

	var injections []model.FaultInjection
	if err := query.
		Joins("JOIN dataset_version_injections dvi ON dvi.injection_id = id").
		Where("state = ? AND status != ?", consts.DatapackBuildSuccess, consts.CommonDeleted).
		Where("dvi.dataset_version_id = ?", versionID).
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to list fault injections for dataset version %d: %w", versionID, err)
	}
	return injections, nil
}
