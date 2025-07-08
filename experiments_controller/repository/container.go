package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"gorm.io/gorm"
)

func CreateContainer(container *database.Container) error {
	var existingContainer database.Container

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("type = ? AND name = ? AND image = ? AND tag = ?", container.Type, container.Name, container.Image, container.Tag).
			First(&existingContainer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(container).Error
			}

			return err
		}

		return tx.Model(&existingContainer).Update("updated_at", tx.NowFunc()).Error
	})
	if err != nil {
		return fmt.Errorf("failed to create or update container: %v", err)
	}

	return nil
}

func GetContaineInfo(opts *dto.GetContainerFilterOptions) (*database.Container, error) {
	query := database.DB.Where("name = ?", opts.Name)

	if opts != nil {
		if opts.Type != "" {
			query = query.Where("type = ?", opts.Type)
		}

		if opts.Image != "" {
			query = query.Where("image = ?", opts.Image)
		}

		if opts.Image != "" && opts.Tag != "" {
			query = query.Where("tag = ?", opts.Tag)
		}
	}

	var record database.Container
	if err := query.
		Order("updated_at DESC").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("container info '%s' not found", opts.Name)
		}

		return nil, fmt.Errorf("failed to query container info: %v", err)
	}

	return &record, nil
}

func ListContainers(opts *dto.ListContainersFilterOptions) ([]database.Container, error) {
	query := database.DB.Order("created_at DESC")

	if opts != nil {
		if opts.Status != nil {
			query = query.Where("status = ?", *opts.Status)
		}

		if opts.Type != "" {
			query = query.Where("type = ?", opts.Type)
		}

		if len(opts.Names) > 0 {
			query = query.Where("name IN ?", opts.Names)
		}
	}

	var containers []database.Container
	if err := query.Find(&containers).Error; err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	return containers, nil
}
