package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BatchDeleteLabels deletes multiple labels and their associations in a transaction
func BatchDeleteLabels(labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		labels, err := repository.ListLabelsByID(tx, labelIDs)
		if err != nil {
			return fmt.Errorf("failed to list labels by IDs: %w", err)
		}

		if len(labels) == 0 {
			return fmt.Errorf("no labels found for the provided IDs")
		}
		if len(labels) != len(labelIDs) {
			return fmt.Errorf("some labels not found for the provided IDs")
		}

		labelMap := make(map[int]*database.Label, len(labels))
		for _, label := range labels {
			labelMap[label.ID] = &label
		}

		containerCountMap, err := removeContainersFromLabels(tx, labelIDs)
		if err != nil {
			return fmt.Errorf("failed to delete container-label associations: %v", err)
		}

		datasetCountMap, err := removeDatasetsFromLabels(tx, labelIDs)
		if err != nil {
			return fmt.Errorf("failed to delete dataset-label associations: %v", err)
		}

		projectCountMap, err := removeProjectsFromLabels(tx, labelIDs)
		if err != nil {
			return fmt.Errorf("failed to delete project-label associations: %v", err)
		}

		injectionCountMap, err := removeInjectionsFromLabels(tx, labelIDs)
		if err != nil {
			return fmt.Errorf("failed to delete injection-label associations: %v", err)
		}

		executionCountMap, err := removeExecutionsFromLabels(tx, labelIDs)
		if err != nil {
			return fmt.Errorf("failed to delete execution-label associations: %v", err)
		}

		toUpdatedLabels := make([]database.Label, 0, len(labelIDs))
		for labelID, label := range labelMap {
			totalDecrement := int64(0)

			if count, exists := containerCountMap[labelID]; exists {
				totalDecrement += count
			}
			if count, exists := datasetCountMap[labelID]; exists {
				totalDecrement += count
			}
			if count, exists := projectCountMap[labelID]; exists {
				totalDecrement += count
			}
			if count, exists := injectionCountMap[labelID]; exists {
				totalDecrement += count
			}
			if count, exists := executionCountMap[labelID]; exists {
				totalDecrement += count
			}

			label.Usage = label.Usage - int(totalDecrement)
			if label.Usage < 0 {
				label.Usage = 0
			}

			toUpdatedLabels = append(toUpdatedLabels, *label)
		}

		if err := repository.BatchUpdateLabels(tx, toUpdatedLabels); err != nil {
			return fmt.Errorf("failed to update label usages: %v", err)
		}

		if err := repository.BatchDeleteLabels(tx, labelIDs); err != nil {
			return fmt.Errorf("failed to batch delete labels: %v", err)
		}

		return nil
	})
}

// CreateLabel creates a new label or reactivates an existing deleted one
func CreateLabel(req *dto.CreateLabelReq) (*dto.LabelResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("label validation failed: %w", err)
	}

	label := req.ConvertToLabel()

	var createdLabel *database.Label
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		label, err := CreateLabelCore(tx, label)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}

		createdLabel = label
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewLabelResp(createdLabel), nil
}

// CreateLabelCore performs the core logic of creating a label within a transaction
func CreateLabelCore(db *gorm.DB, label *database.Label) (*database.Label, error) {
	existingLabel, err := repository.GetLabelByKeyAndValue(db, label.Key, label.Value)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing label: %w", err)
	}

	if existingLabel == nil {
		if err := repository.CreateLabel(db, label); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return nil, fmt.Errorf("%w: label with key %s and value %s already exists", consts.ErrAlreadyExists, label.Key, label.Value)
			}
			return nil, fmt.Errorf("failed to create label: %w", err)
		}

		return label, nil
	}

	existingLabel.Category = label.Category
	existingLabel.Description = label.Description
	existingLabel.Color = label.Color
	existingLabel.Status = consts.CommonEnabled

	if err := repository.UpdateLabel(db, existingLabel); err != nil {
		return nil, fmt.Errorf("failed to update existing label: %w", err)
	}

	return existingLabel, nil
}

// DeleteLabel deletes a label by its ID
func DeleteLabel(labelID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		label, err := repository.GetLabelByID(tx, labelID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: label with id %d not found", consts.ErrNotFound, labelID)
			}
			return fmt.Errorf("failed to get label: %v", err)
		}

		// Delete all related  associations
		containerRows, err := repository.RemoveContainersFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete container-label associations: %v", err)
		}

		datasetRows, err := repository.RemoveDatasetsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete dataset-label associations: %v", err)
		}

		projectRows, err := repository.RemoveProjectsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete project-label associations: %v", err)
		}

		injectionRows, err := repository.RemoveInjectionsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete injection-label associations: %v", err)
		}

		executionRows, err := repository.RemoveExecutionsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete execution-label associations: %v", err)
		}

		totalRows := int(containerRows + datasetRows + projectRows + injectionRows + executionRows)
		if err := repository.BatchDecreaseLabelUsages(tx, []int{label.ID}, totalRows); err != nil {
			return fmt.Errorf("failed to decrease label usage: %v", err)
		}

		// Delete the label itself
		rows, err := repository.DeleteLabel(tx, labelID)
		if err != nil {
			return fmt.Errorf("failed to delete label: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: label id %d not found", consts.ErrNotFound, labelID)
		}

		return nil
	})
}

// GetLabelDetail retrieves detailed information about a label by its ID
func GetLabelDetail(labelID int) (*dto.LabelDetailResp, error) {
	label, err := repository.GetLabelByID(database.DB, labelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: label with ID %d not found", consts.ErrNotFound, labelID)
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return dto.NewLabelDetailResp(label), nil
}

// ListLabels lists labels based on the provided filters
func ListLabels(req *dto.ListLabelReq) (*dto.ListResp[dto.LabelResp], error) {
	limit, offset := req.ToGormParams()
	fitlerOptions := req.ToFilterOptions()

	labels, total, err := repository.ListLabels(database.DB, limit, offset, fitlerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	labelResps := make([]dto.LabelResp, 0, len(labels))
	for i := range labels {
		labelResps = append(labelResps, *dto.NewLabelResp(&labels[i]))
	}

	resp := dto.ListResp[dto.LabelResp]{
		Items:      labelResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateLabel updates an existing label's details
func UpdateLabel(req *dto.UpdateLabelReq, labelID int) (*dto.LabelResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedLabel *database.Label

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingLabel, err := repository.GetLabelByID(tx, labelID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: label with ID %d not found", consts.ErrNotFound, labelID)
			}
			return fmt.Errorf("failed to get label: %w", err)
		}

		req.PatchLabelModel(existingLabel)

		if err := repository.UpdateLabel(tx, existingLabel); err != nil {
			return fmt.Errorf("failed to update label: %w", err)
		}

		updatedLabel = existingLabel
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewLabelResp(updatedLabel), nil
}

// checkLabelKeyValue checks if a label with the specified key and value exists in the provided label slice
func checkLabelKeyValue(labels []database.Label, key, value string) bool {
	for _, label := range labels {
		if label.Key == key && label.Value == value {
			return true
		}
	}
	return false
}

// removeContainersFromLabels removes container associations from multiple labels and returns the total usage count removed
func removeContainersFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return nil, nil
	}

	countsMap, err := repository.ListContainerLabelCounts(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get container-label counts: %w", err)
	}
	if len(countsMap) == 0 {
		return nil, nil
	}

	rows, err := repository.RemoveContainersFromLabels(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to remove containers from labels: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	return countsMap, nil
}

// removeDatasetsFromLabels removes dataset associations from multiple labels and returns the count map
func removeDatasetsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return nil, nil
	}

	countsMap, err := repository.ListDatasetLabelCounts(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset-label counts: %w", err)
	}
	if len(countsMap) == 0 {
		return nil, nil
	}

	rows, err := repository.RemoveDatasetsFromLabels(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to remove datasets from labels: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	return countsMap, nil
}

// removeProjectsFromLabels removes project associations from multiple labels and returns the count map
func removeProjectsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return nil, nil
	}

	countsMap, err := repository.ListProjectLabelCounts(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get project-label counts: %w", err)
	}
	if len(countsMap) == 0 {
		return nil, nil
	}

	rows, err := repository.RemoveProjectsFromLabels(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to remove projects from labels: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	return countsMap, nil
}

// removeInjectionsFromLabels removes injection associations from multiple labels and returns the count map
func removeInjectionsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return nil, nil
	}

	countsMap, err := repository.ListInjectionLabelCounts(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get injection-label counts: %w", err)
	}
	if len(countsMap) == 0 {
		return nil, nil
	}

	rows, err := repository.RemoveInjectionsFromLabels(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to remove injections from labels: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	return countsMap, nil
}

// removeExecutionsFromLabels removes execution associations from multiple labels and returns the count map
func removeExecutionsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return nil, nil
	}

	countsMap, err := repository.ListExecutionLabelCounts(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution-label counts: %w", err)
	}
	if len(countsMap) == 0 {
		return nil, nil
	}

	rows, err := repository.RemoveExecutionsFromLabels(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to remove executions from labels: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	return countsMap, nil
}
