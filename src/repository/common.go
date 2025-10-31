package repository

import (
	"aegis/consts"
	"aegis/database"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	commonOmitFields = "active_name"
)

type ModelConstraint interface {
	database.Container | database.ContainerVersion | database.HelmConfig |
		database.Project |
		database.Resource | database.User | database.Role | database.Permission |
		database.UserContainer | database.UserProject | database.UserRole | database.UserPermission | database.RolePermission
}

func createModel[T ModelConstraint](db *gorm.DB, model *T) error {
	if err := db.Create(model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return consts.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func findModel[T ModelConstraint](db *gorm.DB, condition string, value ...any) (*T, error) {
	var result T
	if err := db.
		Where(condition, value...).
		First(&result).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrNotFound
		}

		return nil, fmt.Errorf("failed to query model with condition '%s': %w", condition, err)
	}
	return &result, nil
}

func updateModel[T ModelConstraint](db *gorm.DB, model *T) error {
	if err := db.Save(model).Error; err != nil {
		return err
	}
	return nil
}
