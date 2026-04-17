package labelmodule

import (
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) BatchDelete(_ context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	return s.repo.Transaction(func(tx *gorm.DB) error {
		labels, err := s.repo.ListLabelsByID(tx, ids)
		if err != nil {
			return fmt.Errorf("failed to list labels by IDs: %w", err)
		}
		if len(labels) == 0 {
			return fmt.Errorf("no labels found for the provided IDs")
		}
		if len(labels) != len(ids) {
			return fmt.Errorf("some labels not found for the provided IDs")
		}

		labelMap := make(map[int]*model.Label, len(labels))
		for _, label := range labels {
			labelMap[label.ID] = &label
		}

		containerCountMap, err := s.removeContainersFromLabels(tx, ids)
		if err != nil {
			return fmt.Errorf("failed to delete container-label associations: %v", err)
		}
		datasetCountMap, err := s.removeDatasetsFromLabels(tx, ids)
		if err != nil {
			return fmt.Errorf("failed to delete dataset-label associations: %v", err)
		}
		projectCountMap, err := s.removeProjectsFromLabels(tx, ids)
		if err != nil {
			return fmt.Errorf("failed to delete project-label associations: %v", err)
		}
		injectionCountMap, err := s.removeInjectionsFromLabels(tx, ids)
		if err != nil {
			return fmt.Errorf("failed to delete injection-label associations: %v", err)
		}
		executionCountMap, err := s.removeExecutionsFromLabels(tx, ids)
		if err != nil {
			return fmt.Errorf("failed to delete execution-label associations: %v", err)
		}

		toUpdatedLabels := make([]model.Label, 0, len(ids))
		for labelID, label := range labelMap {
			totalDecrement := int64(0)
			totalDecrement += containerCountMap[labelID]
			totalDecrement += datasetCountMap[labelID]
			totalDecrement += projectCountMap[labelID]
			totalDecrement += injectionCountMap[labelID]
			totalDecrement += executionCountMap[labelID]
			label.Usage = max(label.Usage-int(totalDecrement), 0)
			toUpdatedLabels = append(toUpdatedLabels, *label)
		}

		if err := s.repo.BatchUpdateLabels(tx, toUpdatedLabels); err != nil {
			return fmt.Errorf("failed to update label usages: %v", err)
		}
		if err := s.repo.BatchDeleteLabels(tx, ids); err != nil {
			return fmt.Errorf("failed to batch delete labels: %v", err)
		}
		return nil
	})
}

func (s *Service) Create(_ context.Context, req *CreateLabelReq) (*LabelResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("label validation failed: %w", err)
	}

	label := req.ConvertToLabel()
	var createdLabel *model.Label
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		item, err := s.createLabelCore(tx, label)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
		createdLabel = item
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewLabelResp(createdLabel), nil
}

func (s *Service) Delete(_ context.Context, id int) error {
	return s.repo.Transaction(func(tx *gorm.DB) error {
		label, err := s.repo.GetLabelByID(tx, id)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: label with id %d not found", consts.ErrNotFound, id)
			}
			return fmt.Errorf("failed to get label: %v", err)
		}

		containerRows, err := s.repo.RemoveContainersFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete container-label associations: %v", err)
		}
		datasetRows, err := s.repo.RemoveDatasetsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete dataset-label associations: %v", err)
		}
		projectRows, err := s.repo.RemoveProjectsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete project-label associations: %v", err)
		}
		injectionRows, err := s.repo.RemoveInjectionsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete injection-label associations: %v", err)
		}
		executionRows, err := s.repo.RemoveExecutionsFromLabel(tx, label.ID)
		if err != nil {
			return fmt.Errorf("failed to delete execution-label associations: %v", err)
		}

		totalRows := int(containerRows + datasetRows + projectRows + injectionRows + executionRows)
		if err := s.repo.BatchDecreaseLabelUsages(tx, []int{label.ID}, totalRows); err != nil {
			return fmt.Errorf("failed to decrease label usage: %v", err)
		}

		rows, err := s.repo.DeleteLabel(tx, id)
		if err != nil {
			return fmt.Errorf("failed to delete label: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: label id %d not found", consts.ErrNotFound, id)
		}
		return nil
	})
}

func (s *Service) GetDetail(_ context.Context, id int) (*LabelDetailResp, error) {
	label, err := s.repo.GetLabelByID(s.repo.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: label with ID %d not found", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}
	return NewLabelDetailResp(label), nil
}

func (s *Service) List(_ context.Context, req *ListLabelReq) (*dto.ListResp[LabelResp], error) {
	limit, offset := req.ToGormParams()
	filterOptions := req.ToFilterOptions()
	labels, total, err := s.repo.ListLabels(limit, offset, filterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}
	items := make([]LabelResp, 0, len(labels))
	for i := range labels {
		items = append(items, *NewLabelResp(&labels[i]))
	}
	return &dto.ListResp[LabelResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) Update(_ context.Context, req *UpdateLabelReq, id int) (*LabelResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedLabel *model.Label
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		existingLabel, err := s.repo.GetLabelByID(tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: label with ID %d not found", consts.ErrNotFound, id)
			}
			return fmt.Errorf("failed to get label: %w", err)
		}

		req.PatchLabelModel(existingLabel)
		if err := s.repo.UpdateLabel(tx, existingLabel); err != nil {
			return fmt.Errorf("failed to update label: %w", err)
		}
		updatedLabel = existingLabel
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewLabelResp(updatedLabel), nil
}

type labelRemovalOps struct {
	countFunc  func(*gorm.DB, []int) (map[int]int64, error)
	removeFunc func(*gorm.DB, []int) (int64, error)
	entityName string
}

func (s *Service) createLabelCore(db *gorm.DB, label *model.Label) (*model.Label, error) {
	return CreateLabelCore(db, label)
}

func (s *Service) removeAssociationsFromLabels(db *gorm.DB, labelIDs []int, ops labelRemovalOps) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return nil, nil
	}
	countsMap, err := ops.countFunc(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s-label counts: %w", ops.entityName, err)
	}
	if len(countsMap) == 0 {
		return nil, nil
	}
	rows, err := ops.removeFunc(db, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to remove %ss from labels: %w", ops.entityName, err)
	}
	if rows == 0 {
		return nil, nil
	}
	return countsMap, nil
}

func (s *Service) removeContainersFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return s.removeAssociationsFromLabels(db, labelIDs, labelRemovalOps{
		countFunc:  s.repo.ListContainerLabelCounts,
		removeFunc: s.repo.RemoveContainersFromLabels,
		entityName: "container",
	})
}

func (s *Service) removeDatasetsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return s.removeAssociationsFromLabels(db, labelIDs, labelRemovalOps{
		countFunc:  s.repo.ListDatasetLabelCounts,
		removeFunc: s.repo.RemoveDatasetsFromLabels,
		entityName: "dataset",
	})
}

func (s *Service) removeProjectsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return s.removeAssociationsFromLabels(db, labelIDs, labelRemovalOps{
		countFunc:  s.repo.ListProjectLabelCounts,
		removeFunc: s.repo.RemoveProjectsFromLabels,
		entityName: "project",
	})
}

func (s *Service) removeInjectionsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return s.removeAssociationsFromLabels(db, labelIDs, labelRemovalOps{
		countFunc:  s.repo.ListInjectionLabelCounts,
		removeFunc: s.repo.RemoveInjectionsFromLabels,
		entityName: "injection",
	})
}

func (s *Service) removeExecutionsFromLabels(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	return s.removeAssociationsFromLabels(db, labelIDs, labelRemovalOps{
		countFunc:  s.repo.ListExecutionLabelCounts,
		removeFunc: s.repo.RemoveExecutionsFromLabels,
		entityName: "execution",
	})
}
