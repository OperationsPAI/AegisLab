package repository

import (
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
)

func FindTaskByID(id string) (*dto.TaskItem, error) {
	var result database.Task
	if err := database.DB.Where("tasks.id = ?", id).First(&result).Error; err != nil {
		return nil, err
	}

	var taskItem dto.TaskItem
	if err := taskItem.Convert(result); err != nil {
		return nil, err
	}

	return &taskItem, nil
}
