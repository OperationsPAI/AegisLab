package initialization

import (
	"encoding/json"
	"fmt"
	"os"

	"aegis/database"

	"github.com/sirupsen/logrus"
)

func loadInitialDataFromFile(filePath string) (*InitialData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial data file: %w", err)
	}

	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initial data: %w", err)
	}

	return &initialData, nil
}

func withOptimizedDBSettings(fn func() error) error {
	if err := database.DB.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		logrus.Warnf("Failed to disable foreign key checks: %v", err)
	}
	if err := database.DB.Exec("SET UNIQUE_CHECKS=0").Error; err != nil {
		logrus.Warnf("Failed to disable unique checks: %v", err)
	}

	defer func() {
		if err := database.DB.Exec("SET FOREIGN_KEY_CHECKS=1").Error; err != nil {
			logrus.Errorf("Failed to re-enable foreign key checks: %v", err)
		}
		if err := database.DB.Exec("SET UNIQUE_CHECKS=1").Error; err != nil {
			logrus.Errorf("Failed to re-enable unique checks: %v", err)
		}
	}()

	return fn()
}
