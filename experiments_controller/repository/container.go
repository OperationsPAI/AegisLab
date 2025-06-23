package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"gorm.io/gorm"
)

func GetContaineInfo(name string, cType consts.ContainerType) (*database.Container, error) {
	var record database.Container
	if err := database.DB.
		Where("name = ? AND type = ? AND status = ?", name, cType, true).
		Order("created_at DESC").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("container info '%s' not found", name)
		}

		return nil, fmt.Errorf("failed to query container info: %v", err)
	}

	return &record, nil
}

func ListContainers(opts *dto.FilterContainerOptions) ([]database.Container, error) {
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
