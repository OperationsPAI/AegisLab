package common

import (
	"aegis/dto"
	"aegis/model"
	"aegis/repository"
	"fmt"

	"gorm.io/gorm"
)

// mapRefsToDatasetVersions maps dataset refs to their corresponding dataset versions
func MapRefsToDatasetVersionsWithDB(db *gorm.DB, refs []*dto.DatasetRef, userID int) (map[*dto.DatasetRef]model.DatasetVersion, error) {
	versions, err := getUniqueVersionsForDatasetRefsWithDB(db, refs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get dataset versions: %w", err)
	}

	flatMap := make(map[string][]model.DatasetVersion)
	hierarchicalMap := make(map[string]map[string]model.DatasetVersion)

	for _, version := range versions {
		datasetName := version.Dataset.Name
		versionName := version.Name

		flatMap[datasetName] = append(flatMap[datasetName], version)

		if _, exists := hierarchicalMap[datasetName]; !exists {
			hierarchicalMap[datasetName] = make(map[string]model.DatasetVersion)
		}
		hierarchicalMap[datasetName][versionName] = version
	}

	results := make(map[*dto.DatasetRef]model.DatasetVersion, len(refs))
	for _, ref := range refs {
		var result model.DatasetVersion
		if ref.Version != "" {
			if _, exists := hierarchicalMap[ref.Name]; !exists {
				return nil, fmt.Errorf("dataset not found: %s", ref.Name)
			}

			if _, exists := hierarchicalMap[ref.Name][ref.Version]; !exists {
				return nil, fmt.Errorf("dataset version not found: %s:%s", ref.Name, ref.Version)
			}

			result = hierarchicalMap[ref.Name][ref.Version]
		} else {
			if _, exists := flatMap[ref.Name]; !exists {
				return nil, fmt.Errorf("dataset not found: %s", ref.Name)
			}
			result = flatMap[ref.Name][0]
		}

		results[ref] = result
	}

	return results, nil
}

func getUniqueVersionsForDatasetRefsWithDB(db *gorm.DB, refs []*dto.DatasetRef, userID int) ([]model.DatasetVersion, error) {
	datasetNamesSet := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if ref.Name != "" {
			datasetNamesSet[ref.Name] = struct{}{}
		}
	}

	if len(datasetNamesSet) == 0 {
		return []model.DatasetVersion{}, nil
	}

	requiredNames := make([]string, 0, len(datasetNamesSet))
	for name := range datasetNamesSet {
		requiredNames = append(requiredNames, name)
	}

	versions, err := repository.BatchGetDatasetVersions(db, requiredNames, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get dataset versions: %w", err)
	}

	return versions, nil
}
