package labelmodule

import (
	"aegis/consts"
	"aegis/model"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const labelKeyOmitFields = "active_key_value"

type Repository struct {
	db *gorm.DB
}

type labelCountResult struct {
	LabelID int   `gorm:"column:label_id"`
	Count   int64 `gorm:"column:count"`
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListLabelsByID(db *gorm.DB, labelIDs []int) ([]model.Label, error) {
	if len(labelIDs) == 0 {
		return []model.Label{}, nil
	}

	var labels []model.Label
	if err := r.useDB(db).
		Where("id IN (?) AND status != ?", labelIDs, consts.CommonDeleted).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels by IDs: %w", err)
	}
	return labels, nil
}

func (r *Repository) BatchUpdateLabels(db *gorm.DB, labels []model.Label) error {
	if len(labels) == 0 {
		return fmt.Errorf("no labels to update")
	}

	if err := r.useDB(db).Omit(labelKeyOmitFields).Save(&labels).Error; err != nil {
		return fmt.Errorf("failed to batch update labels: %w", err)
	}
	return nil
}

func (r *Repository) BatchDeleteLabels(db *gorm.DB, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	if err := r.useDB(db).Model(&model.Label{}).
		Where("id IN (?) AND status != ?", labelIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete labels: %w", err)
	}
	return nil
}

func (r *Repository) GetLabelByKeyAndValue(db *gorm.DB, key, value string, status ...consts.StatusType) (*model.Label, error) {
	query := r.useDB(db).Where("label_key = ? AND label_value = ?", key, value)
	if len(status) == 0 {
		query = query.Where("status != ?", consts.CommonDeleted)
	} else if len(status) == 1 {
		query = query.Where("status = ?", status[0])
	} else {
		query = query.Where("status IN (?)", status)
	}

	var label model.Label
	if err := query.First(&label).Error; err != nil {
		return nil, fmt.Errorf("failed to get label: %w", err)
	}
	return &label, nil
}

func (r *Repository) batchCreateLabels(labels []model.Label) error {
	if len(labels) == 0 {
		return nil
	}
	if err := r.db.Omit(labelKeyOmitFields).Create(&labels).Error; err != nil {
		return fmt.Errorf("failed to batch upsert labels: %w", err)
	}
	return nil
}

func (r *Repository) batchIncreaseLabelUsages(labelIDs []int, increment int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := gorm.Expr("usage_count + ?", increment)
	if err := r.db.Model(&model.Label{}).
		Where("id IN (?)", labelIDs).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch increase label usages: %w", err)
	}
	return nil
}

func (r *Repository) listLabelsByConditions(conditions []map[string]string) ([]model.Label, error) {
	if len(conditions) == 0 {
		return []model.Label{}, nil
	}

	var labels []model.Label
	query := r.db.Model(&model.Label{})
	var whereClauses []string
	var whereArgs []any

	for _, condition := range conditions {
		whereClauses = append(whereClauses, "(label_key = ? AND label_value = ?)")
		whereArgs = append(whereArgs, condition["key"], condition["value"])
	}

	if len(whereClauses) > 0 {
		query = query.Where(strings.Join(whereClauses, " OR "), whereArgs...)
	}

	if err := query.Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels by conditions: %w", err)
	}
	return labels, nil
}

func (r *Repository) CreateLabel(db *gorm.DB, label *model.Label) error {
	if err := r.useDB(db).Omit(labelKeyOmitFields).Create(label).Error; err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}
	return nil
}

func (r *Repository) UpdateLabel(db *gorm.DB, label *model.Label) error {
	if err := r.useDB(db).Omit(labelKeyOmitFields).Save(label).Error; err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}
	return nil
}

func (r *Repository) GetLabelByID(db *gorm.DB, id int) (*model.Label, error) {
	var label model.Label
	if err := r.useDB(db).First(&label, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("label with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}
	return &label, nil
}

func (r *Repository) RemoveContainersFromLabel(db *gorm.DB, labelID int) (int64, error) {
	return r.removeAssociationsFromLabel(db, &model.ContainerLabel{}, labelID, "containers")
}

func (r *Repository) RemoveDatasetsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	return r.removeAssociationsFromLabel(db, &model.DatasetLabel{}, labelID, "datasets")
}

func (r *Repository) RemoveProjectsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	return r.removeAssociationsFromLabel(db, &model.ProjectLabel{}, labelID, "projects")
}

func (r *Repository) RemoveInjectionsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	return r.removeAssociationsFromLabel(db, &model.FaultInjectionLabel{}, labelID, "injection-label associations")
}

func (r *Repository) RemoveExecutionsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	return r.removeAssociationsFromLabel(db, &model.ExecutionInjectionLabel{}, labelID, "execution-label associations")
}

func (r *Repository) ListContainerLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return r.listAssociationCounts(db, &model.ContainerLabel{}, labelIDs)
}

func (r *Repository) RemoveContainersFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	return r.removeAssociationsFromLabels(db, &model.ContainerLabel{}, labelIDs, "containers")
}

func (r *Repository) ListDatasetLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return r.listAssociationCounts(db, &model.DatasetLabel{}, labelIDs)
}

func (r *Repository) RemoveDatasetsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	return r.removeAssociationsFromLabels(db, &model.DatasetLabel{}, labelIDs, "datasets")
}

func (r *Repository) ListProjectLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return r.listAssociationCounts(db, &model.ProjectLabel{}, labelIDs)
}

func (r *Repository) RemoveProjectsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	return r.removeAssociationsFromLabels(db, &model.ProjectLabel{}, labelIDs, "projects")
}

func (r *Repository) ListInjectionLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return r.listAssociationCounts(db, &model.FaultInjectionLabel{}, labelIDs)
}

func (r *Repository) RemoveInjectionsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	return r.removeAssociationsFromLabels(db, &model.FaultInjectionLabel{}, labelIDs, "injection-label associations")
}

func (r *Repository) ListExecutionLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return r.listAssociationCounts(db, &model.ExecutionInjectionLabel{}, labelIDs)
}

func (r *Repository) RemoveExecutionsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	return r.removeAssociationsFromLabels(db, &model.ExecutionInjectionLabel{}, labelIDs, "execution-label associations")
}

func (r *Repository) BatchDecreaseLabelUsages(db *gorm.DB, labelIDs []int, decrement int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := gorm.Expr("GREATEST(0, usage_count - ?)", decrement)
	if err := r.useDB(db).Model(&model.Label{}).
		Where("id IN (?)", labelIDs).
		Clauses(clause.Returning{}).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

func (r *Repository) DeleteLabel(db *gorm.DB, labelID int) (int64, error) {
	result := r.useDB(db).Model(&model.Label{}).
		Where("id = ? AND status != ?", labelID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete label %d: %w", labelID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) ListLabels(limit, offset int, filterOptions *ListLabelFilters) ([]model.Label, int64, error) {
	var (
		labels []model.Label
		total  int64
	)

	query := r.db.Model(&model.Label{})
	if filterOptions.Key != "" {
		query = query.Where("label_key = ?", filterOptions.Key)
	}
	if filterOptions.Value != "" {
		query = query.Where("label_value = ?", filterOptions.Value)
	}
	if filterOptions.Category != nil {
		query = query.Where("category = ?", *filterOptions.Category)
	}
	if filterOptions.IsSystem != nil {
		query = query.Where("is_system = ?", *filterOptions.IsSystem)
	}
	if filterOptions.Status != nil {
		query = query.Where("status = ?", *filterOptions.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count labels: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("usage_count DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list labels: %w", err)
	}
	return labels, total, nil
}

func (r *Repository) useDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func (r *Repository) removeAssociationsFromLabel(db *gorm.DB, model any, labelID int, target string) (int64, error) {
	result := r.useDB(db).Where("label_id = ?", labelID).Delete(model)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to remove %s from label %d: %w", target, labelID, err)
	}
	return result.RowsAffected, nil
}

func (r *Repository) removeAssociationsFromLabels(db *gorm.DB, model any, labelIDs []int, target string) (int64, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}

	result := r.useDB(db).Where("label_id IN (?)", labelIDs).Delete(model)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to remove %s from labels %v: %w", target, labelIDs, err)
	}
	return result.RowsAffected, nil
}

func (r *Repository) listAssociationCounts(db *gorm.DB, model any, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return map[int]int64{}, nil
	}

	var results []labelCountResult
	if err := r.useDB(db).Model(model).
		Select("label_id, COUNT(label_id) AS count").
		Where("label_id IN (?)", labelIDs).
		Group("label_id").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count associations: %w", err)
	}

	countMap := make(map[int]int64, len(results))
	for _, result := range results {
		countMap[result.LabelID] = result.Count
	}
	return countMap, nil
}
