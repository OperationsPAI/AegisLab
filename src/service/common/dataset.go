package common

import (
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"fmt"
)

// mapRefsToDatasetVersions maps dataset refs to their corresponding dataset versions
func MapRefsToDatasetVersions(refs []*dto.DatasetRef, userID int) (map[*dto.DatasetRef]database.DatasetVersion, error) {
	versions, err := getUniqueVersionsForDatasetRefs(refs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get dataset versions: %w", err)
	}

	flatMap := make(map[string][]database.DatasetVersion)
	hierarchicalMap := make(map[string]map[string]database.DatasetVersion)

	for _, version := range versions {
		datasetName := version.Dataset.Name
		versionName := version.Name

		flatMap[datasetName] = append(flatMap[datasetName], version)

		if _, exists := hierarchicalMap[datasetName]; !exists {
			hierarchicalMap[datasetName] = make(map[string]database.DatasetVersion)
		}
		hierarchicalMap[datasetName][versionName] = version
	}

	results := make(map[*dto.DatasetRef]database.DatasetVersion, len(refs))
	for _, ref := range refs {
		var result database.DatasetVersion
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

// getUniqueVersionsForDatasetrefs retrieves unique dataset versions for the given dataset refs
func getUniqueVersionsForDatasetRefs(refs []*dto.DatasetRef, userID int) ([]database.DatasetVersion, error) {
	datasetNamesSet := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if ref.Name != "" {
			datasetNamesSet[ref.Name] = struct{}{}
		}
	}

	if len(datasetNamesSet) == 0 {
		return []database.DatasetVersion{}, nil
	}

	requiredNames := make([]string, 0, len(datasetNamesSet))
	for name := range datasetNamesSet {
		requiredNames = append(requiredNames, name)
	}

	versions, err := repository.BatchGetDatasetVersions(database.DB, requiredNames, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get dataset versions: %w", err)
	}

	return versions, nil
}
