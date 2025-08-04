package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreateDataset creates a dataset
func CreateDataset(dataset *database.Dataset) error {
	if err := database.DB.Create(dataset).Error; err != nil {
		return fmt.Errorf("failed to create dataset: %v", err)
	}
	return nil
}

// GetDatasetByID gets dataset by ID
func GetDatasetByID(id int) (*database.Dataset, error) {
	var dataset database.Dataset
	if err := database.DB.Where("id = ? AND status != ?", id, consts.DatasetDeleted).First(&dataset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get dataset: %v", err)
	}
	return &dataset, nil
}

// GetDatasetByNameAndVersion gets dataset by name and version
func GetDatasetByNameAndVersion(name, version string) (*database.Dataset, error) {
	var dataset database.Dataset
	if err := database.DB.
		Where("name = ? AND version = ? AND status != ?", name, version, consts.DatasetDeleted).
		First(&dataset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset '%s:%s' not found", name, version)
		}
		return nil, fmt.Errorf("failed to get dataset: %v", err)
	}
	return &dataset, nil
}

// UpdateDataset updates dataset information
func UpdateDataset(dataset *database.Dataset) error {
	if err := database.DB.Save(dataset).Error; err != nil {
		return fmt.Errorf("failed to update dataset: %v", err)
	}
	return nil
}

// DeleteDataset soft deletes dataset (sets status to -1)
func DeleteDataset(id int) error {
	if err := database.DB.Model(&database.Dataset{}).Where("id = ?", id).Update("status", consts.DatapackDeleted).Error; err != nil {
		return fmt.Errorf("failed to delete dataset: %v", err)
	}
	return nil
}

// ListDatasets gets dataset list
func ListDatasets(page, pageSize int, datasetType string, status *int, isPublic *bool) ([]database.Dataset, int64, error) {
	var datasets []database.Dataset
	var total int64

	query := database.DB.Model(&database.Dataset{})

	// Always exclude deleted datasets (status = -1) unless explicitly requested
	if status != nil {
		query = query.Where("status = ?", *status)
	} else {
		query = query.Where("status != ?", consts.DatapackDeleted)
	}

	if datasetType != "" {
		query = query.Where("type = ?", datasetType)
	}

	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count datasets: %v", err)
	}

	// Pagination query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&datasets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list datasets: %v", err)
	}

	return datasets, total, nil
}

// GetDatasetVersions gets all versions of a dataset
func GetDatasetVersions(name string) ([]database.Dataset, error) {
	var datasets []database.Dataset
	if err := database.DB.Where("name = ? AND status != ?", name, consts.DatapackDeleted).
		Order("created_at DESC").Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset versions: %v", err)
	}
	return datasets, nil
}

// GetDatasetFaultInjections gets fault injections associated with dataset
func GetDatasetFaultInjections(datasetID int) ([]database.DatasetFaultInjection, error) {
	var relations []database.DatasetFaultInjection
	if err := database.DB.Preload("FaultInjectionSchedule").
		Where("dataset_id = ?", datasetID).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset fault injections: %v", err)
	}
	return relations, nil
}

// AddDatasetToFaultInjection associates dataset with fault injection
func AddDatasetToFaultInjection(datasetID, faultInjectionID int, relationType string, isGroundTruth bool) error {
	relation := &database.DatasetFaultInjection{
		DatasetID:        datasetID,
		FaultInjectionID: faultInjectionID,
	}

	if err := database.DB.Create(relation).Error; err != nil {
		return fmt.Errorf("failed to add dataset to fault injection: %v", err)
	}
	return nil
}

// RemoveDatasetFromFaultInjection removes association between dataset and fault injection
func RemoveDatasetFromFaultInjection(datasetID, faultInjectionID int) error {
	if err := database.DB.Where("dataset_id = ? AND fault_injection_id = ?", datasetID, faultInjectionID).
		Delete(&database.DatasetFaultInjection{}).Error; err != nil {
		return fmt.Errorf("failed to remove dataset from fault injection: %v", err)
	}
	return nil
}

// GetDatasetLabels gets dataset labels (optimized)
func GetDatasetLabels(datasetID int) ([]database.Label, error) {
	labelsMap, err := GetDatasetLabelsMap([]int{datasetID})
	if err != nil {
		return nil, err
	}
	return labelsMap[datasetID], nil
}

// AddLabelToDataset adds label to dataset
func AddLabelToDataset(datasetID, labelID int) error {
	datasetLabel := &database.DatasetLabel{
		DatasetID: datasetID,
		LabelID:   labelID,
	}

	if err := database.DB.Create(datasetLabel).Error; err != nil {
		return fmt.Errorf("failed to add label to dataset: %v", err)
	}

	// Increase label usage count
	if err := database.DB.Model(&database.Label{}).Where("id = ?", labelID).
		UpdateColumn("usage", gorm.Expr("usage + 1")).Error; err != nil {
		return fmt.Errorf("failed to update label usage: %v", err)
	}

	return nil
}

// RemoveLabelFromDataset removes label from dataset
func RemoveLabelFromDataset(datasetID, labelID int) error {
	if err := database.DB.Where("dataset_id = ? AND label_id = ?", datasetID, labelID).
		Delete(&database.DatasetLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove label from dataset: %v", err)
	}

	if err := database.DB.Model(&database.Label{}).Where("id = ?", labelID).
		UpdateColumn("usage", gorm.Expr("usage - 1")).Error; err != nil {
		return fmt.Errorf("failed to update label usage: %v", err)
	}

	return nil
}

// SearchDatasetsByLabels searches datasets by labels
func SearchDatasetsByLabels(labelKeys []string, labelValues []string) ([]database.Dataset, error) {
	var datasets []database.Dataset

	query := database.DB.Model(&database.Dataset{}).
		Joins("JOIN dataset_labels ON datasets.id = dataset_labels.dataset_id").
		Joins("JOIN labels ON dataset_labels.label_id = labels.id").
		Where("datasets.status != ?", consts.DatapackDeleted)

	if len(labelKeys) > 0 {
		query = query.Where("labels.key IN ?", labelKeys)
	}

	if len(labelValues) > 0 {
		query = query.Where("labels.value IN ?", labelValues)
	}

	if err := query.Distinct().Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("failed to search datasets by labels: %v", err)
	}

	return datasets, nil
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
	if err := database.DB.Model(&database.Dataset{}).Where("status = ?", consts.DatapackDeleted).Count(&deleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count deleted datasets: %v", err)
	}
	stats["deleted"] = deleted

	return stats, nil
}

// GetDatasetCountByType returns count of datasets grouped by type
func GetDatasetCountByType() (map[string]int64, error) {
	type TypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var results []TypeCount
	err := database.DB.Model(&database.Dataset{}).
		Select("type, COUNT(*) as count").
		Where("status != ?", consts.DatapackDeleted).
		Group("type").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count datasets by type: %v", err)
	}

	typeCounts := make(map[string]int64)
	for _, result := range results {
		typeCounts[result.Type] = result.Count
	}

	return typeCounts, nil
}

// GetDatasetTotalSize gets total size of datasets
func GetDatasetTotalSize() (int64, error) {
	var totalSize int64
	if err := database.DB.Model(&database.Dataset{}).
		Where("status != ?", consts.DatapackDeleted).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate dataset total size: %v", err)
	}
	return totalSize, nil
}

// RemoveAllLabelsFromDataset removes all labels from dataset
func RemoveAllLabelsFromDataset(datasetID int) error {
	// Get all label IDs for this dataset to update usage count
	var labelIDs []int
	if err := database.DB.Model(&database.DatasetLabel{}).
		Where("dataset_id = ?", datasetID).
		Pluck("label_id", &labelIDs).Error; err != nil {
		return fmt.Errorf("failed to get label IDs: %v", err)
	}

	// Delete all label associations
	if err := database.DB.Where("dataset_id = ?", datasetID).
		Delete(&database.DatasetLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from dataset: %v", err)
	}

	// Update usage count for all affected labels
	if len(labelIDs) > 0 {
		if err := database.DB.Model(&database.Label{}).
			Where("id IN ?", labelIDs).
			UpdateColumn("usage", gorm.Expr("usage - 1")).Error; err != nil {
			return fmt.Errorf("failed to update label usage: %v", err)
		}
	}

	return nil
}

// RemoveAllInjectionsFromDataset removes all fault injections from dataset
func RemoveAllInjectionsFromDataset(datasetID int) error {
	if err := database.DB.Where("dataset_id = ?", datasetID).
		Delete(&database.DatasetFaultInjection{}).Error; err != nil {
		return fmt.Errorf("failed to remove all injections from dataset: %v", err)
	}
	return nil
}

// GetDeletedDatasetByNameAndVersion gets deleted dataset by name and version
func GetDeletedDatasetByNameAndVersion(name, version string) (*database.Dataset, error) {
	var dataset database.Dataset
	if err := database.DB.
		Where("name = ? AND version = ? AND status = ?", name, version, consts.DatapackDeleted).
		First(&dataset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("deleted dataset '%s:%s' not found", name, version)
		}
		return nil, fmt.Errorf("failed to get deleted dataset: %v", err)
	}
	return &dataset, nil
}
