package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreateDataset 创建数据集
func CreateDataset(dataset *database.Dataset) error {
	if err := database.DB.Create(dataset).Error; err != nil {
		return fmt.Errorf("failed to create dataset: %v", err)
	}
	return nil
}

// GetDatasetByID 根据ID获取数据集
func GetDatasetByID(id int) (*database.Dataset, error) {
	var dataset database.Dataset
	if err := database.DB.Preload("Project").First(&dataset, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get dataset: %v", err)
	}
	return &dataset, nil
}

// GetDatasetByNameAndVersion 根据名称和版本获取数据集
func GetDatasetByNameAndVersion(name, version string) (*database.Dataset, error) {
	var dataset database.Dataset
	if err := database.DB.Preload("Project").
		Where("name = ? AND version = ?", name, version).
		First(&dataset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset '%s:%s' not found", name, version)
		}
		return nil, fmt.Errorf("failed to get dataset: %v", err)
	}
	return &dataset, nil
}

// UpdateDataset 更新数据集信息
func UpdateDataset(dataset *database.Dataset) error {
	if err := database.DB.Save(dataset).Error; err != nil {
		return fmt.Errorf("failed to update dataset: %v", err)
	}
	return nil
}

// DeleteDataset 软删除数据集（设置状态为-1）
func DeleteDataset(id int) error {
	if err := database.DB.Model(&database.Dataset{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete dataset: %v", err)
	}
	return nil
}

// ListDatasets 获取数据集列表
func ListDatasets(page, pageSize int, projectID *int, datasetType string, status *int, isPublic *bool) ([]database.Dataset, int64, error) {
	var datasets []database.Dataset
	var total int64

	query := database.DB.Model(&database.Dataset{}).Preload("Project")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if datasetType != "" {
		query = query.Where("type = ?", datasetType)
	}

	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count datasets: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&datasets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list datasets: %v", err)
	}

	return datasets, total, nil
}

// GetDatasetVersions 获取数据集的所有版本
func GetDatasetVersions(name string) ([]database.Dataset, error) {
	var datasets []database.Dataset
	if err := database.DB.Where("name = ? AND status != -1", name).
		Order("created_at DESC").Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset versions: %v", err)
	}
	return datasets, nil
}

// GetDatasetFaultInjections 获取数据集关联的故障注入
func GetDatasetFaultInjections(datasetID int) ([]database.DatasetFaultInjection, error) {
	var relations []database.DatasetFaultInjection
	if err := database.DB.Preload("FaultInjectionSchedule").
		Where("dataset_id = ?", datasetID).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset fault injections: %v", err)
	}
	return relations, nil
}

// AddDatasetToFaultInjection 将数据集与故障注入关联
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

// RemoveDatasetFromFaultInjection 移除数据集与故障注入的关联
func RemoveDatasetFromFaultInjection(datasetID, faultInjectionID int) error {
	if err := database.DB.Where("dataset_id = ? AND fault_injection_id = ?", datasetID, faultInjectionID).
		Delete(&database.DatasetFaultInjection{}).Error; err != nil {
		return fmt.Errorf("failed to remove dataset from fault injection: %v", err)
	}
	return nil
}

// GetDatasetLabels 获取数据集的标签
func GetDatasetLabels(datasetID int) ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Table("labels").
		Joins("JOIN dataset_labels ON labels.id = dataset_labels.label_id").
		Where("dataset_labels.dataset_id = ?", datasetID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset labels: %v", err)
	}
	return labels, nil
}

// AddLabelToDataset 给数据集添加标签
func AddLabelToDataset(datasetID, labelID int) error {
	datasetLabel := &database.DatasetLabel{
		DatasetID: datasetID,
		LabelID:   labelID,
	}

	if err := database.DB.Create(datasetLabel).Error; err != nil {
		return fmt.Errorf("failed to add label to dataset: %v", err)
	}

	// 增加标签使用次数
	if err := database.DB.Model(&database.Label{}).Where("id = ?", labelID).
		UpdateColumn("usage", gorm.Expr("usage + 1")).Error; err != nil {
		return fmt.Errorf("failed to update label usage: %v", err)
	}

	return nil
}

// RemoveLabelFromDataset 移除数据集的标签
func RemoveLabelFromDataset(datasetID, labelID int) error {
	if err := database.DB.Where("dataset_id = ? AND label_id = ?", datasetID, labelID).
		Delete(&database.DatasetLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove label from dataset: %v", err)
	}

	// 减少标签使用次数
	if err := database.DB.Model(&database.Label{}).Where("id = ?", labelID).
		UpdateColumn("usage", gorm.Expr("usage - 1")).Error; err != nil {
		return fmt.Errorf("failed to update label usage: %v", err)
	}

	return nil
}

// SearchDatasetsByLabels 根据标签搜索数据集
func SearchDatasetsByLabels(labelKeys []string, labelValues []string) ([]database.Dataset, error) {
	var datasets []database.Dataset

	query := database.DB.Model(&database.Dataset{}).
		Joins("JOIN dataset_labels ON datasets.id = dataset_labels.dataset_id").
		Joins("JOIN labels ON dataset_labels.label_id = labels.id")

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
	if err := database.DB.Model(&database.Dataset{}).Where("status = 1").Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active datasets: %v", err)
	}
	stats["active"] = active

	// Disabled datasets
	var disabled int64
	if err := database.DB.Model(&database.Dataset{}).Where("status = 0").Count(&disabled).Error; err != nil {
		return nil, fmt.Errorf("failed to count disabled datasets: %v", err)
	}
	stats["disabled"] = disabled

	// Deleted datasets
	var deleted int64
	if err := database.DB.Model(&database.Dataset{}).Where("status = -1").Count(&deleted).Error; err != nil {
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
		Where("status != -1").
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

// GetDatasetTotalSize 获取数据集总存储大小
func GetDatasetTotalSize() (int64, error) {
	var totalSize int64
	if err := database.DB.Model(&database.Dataset{}).
		Where("status != -1").
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate dataset total size: %v", err)
	}
	return totalSize, nil
}
