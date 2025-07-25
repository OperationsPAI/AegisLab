package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
)

func GetProject(column, param string) (*database.Project, error) {
	var record database.Project
	if err := database.DB.
		Where(fmt.Sprintf("%s = ?", column), param).
		First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}
