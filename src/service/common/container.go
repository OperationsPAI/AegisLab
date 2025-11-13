package common

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"fmt"
)

// MapRefsToContainerVersions maps container refs to their corresponding container versions
func MapRefsToContainerVersions(refs []*dto.ContainerRef, containerType consts.ContainerType, userID int) (map[*dto.ContainerRef]database.ContainerVersion, error) {
	versions, err := getUniqueVersionsForContainerRefs(refs, containerType, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get container versions: %w", err)
	}

	flatMap := make(map[string][]database.ContainerVersion)
	hierarchicalMap := make(map[string]map[string]database.ContainerVersion)

	for _, version := range versions {
		containerName := version.Container.Name
		versionName := version.Name

		flatMap[containerName] = append(flatMap[containerName], version)

		if _, exists := hierarchicalMap[containerName]; !exists {
			hierarchicalMap[containerName] = make(map[string]database.ContainerVersion)
		}
		hierarchicalMap[containerName][versionName] = version
	}

	results := make(map[*dto.ContainerRef]database.ContainerVersion, len(refs))
	for _, ref := range refs {
		var result database.ContainerVersion
		if ref.Version != "" {
			if _, exists := hierarchicalMap[ref.Name]; !exists {
				return nil, fmt.Errorf("container not found: %s", ref.Name)
			}

			if _, exists := hierarchicalMap[ref.Name][ref.Version]; !exists {
				return nil, fmt.Errorf("container version not found: %s:%s", ref.Name, ref.Version)
			}

			result = hierarchicalMap[ref.Name][ref.Version]
		} else {
			if _, exists := flatMap[ref.Name]; !exists {
				return nil, fmt.Errorf("container not found: %s", ref.Name)
			}
			result = flatMap[ref.Name][0]
		}

		results[ref] = result
	}

	return results, nil
}

// getUniqueVersionsForContainerRefs retrieves unique container versions for the given container refs
func getUniqueVersionsForContainerRefs(refs []*dto.ContainerRef, containerType consts.ContainerType, userID int) ([]database.ContainerVersion, error) {
	containerNamesSet := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if ref.Name != "" {
			containerNamesSet[ref.Name] = struct{}{}
		}
	}

	if len(containerNamesSet) == 0 {
		return []database.ContainerVersion{}, nil
	}

	requiredNames := make([]string, 0, len(containerNamesSet))
	for name := range containerNamesSet {
		requiredNames = append(requiredNames, name)
	}

	versions, err := repository.BatchGetContainerVersions(database.DB, containerType, requiredNames, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get container versions: %w", err)
	}

	return versions, nil
}
