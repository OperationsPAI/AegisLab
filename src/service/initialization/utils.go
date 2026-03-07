package initialization

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"aegis/database"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert/yaml"
)

func loadInitialDataFromFile(filePath string) (*InitialData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial data file: %w", err)
	}

	var initialData InitialData

	ext := filepath.Ext(filePath)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &initialData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal initial data: %w", err)
		}
	case ".yaml":
		if err := yaml.Unmarshal(data, &initialData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal initial data: %w", err)
		}
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
