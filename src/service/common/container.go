package common

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/repository"
	"aegis/utils"
	"fmt"

	"gorm.io/gorm"
)

func ListContainerVersionEnvVarsWithDB(db *gorm.DB, specs []dto.ParameterSpec, version *model.ContainerVersion) ([]dto.ParameterItem, error) {
	return listParameterItemsWithDB(db, specs, repository.ListContainerVersionEnvVars, version.ID, version)
}

func ListHelmConfigValuesWithDB(db *gorm.DB, specs []dto.ParameterSpec, cfg *model.HelmConfig) ([]dto.ParameterItem, error) {
	return listParameterItemsWithDB(db, specs, repository.ListHelmConfigValues, cfg.ID, cfg.ContainerVersion)
}

func MapRefsToContainerVersionsWithDB(db *gorm.DB, refs []*dto.ContainerRef, containerType consts.ContainerType, userID int) (map[*dto.ContainerRef]model.ContainerVersion, error) {
	versions, err := getUniqueVersionsForContainerRefsWithDB(db, refs, containerType, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get container versions: %w", err)
	}

	flatMap := make(map[string][]model.ContainerVersion)
	hierarchicalMap := make(map[string]map[string]model.ContainerVersion)

	for _, version := range versions {
		containerName := version.Container.Name
		versionName := version.Name

		flatMap[containerName] = append(flatMap[containerName], version)

		if _, exists := hierarchicalMap[containerName]; !exists {
			hierarchicalMap[containerName] = make(map[string]model.ContainerVersion)
		}
		hierarchicalMap[containerName][versionName] = version
	}

	results := make(map[*dto.ContainerRef]model.ContainerVersion, len(refs))
	for _, ref := range refs {
		var result model.ContainerVersion
		containerTypeName := consts.GetContainerTypeName(containerType)
		if ref.Version != "" {
			if _, exists := hierarchicalMap[ref.Name]; !exists {
				availableContainers := getAvailableContainerNames(hierarchicalMap)
				if len(availableContainers) == 0 {
					// Check if container exists with different type
					exists, actualType, err := repository.CheckContainerExistsWithDifferentType(db, ref.Name, containerType, userID)
					if err != nil {
						return nil, fmt.Errorf("failed to check container type: %w", err)
					}
					if exists {
						return nil, fmt.Errorf("%s container '%s' not found: container exists but has type '%s', not '%s'",
							containerTypeName, ref.Name, consts.GetContainerTypeName(actualType), containerTypeName)
					}
					return nil, fmt.Errorf("%s container '%s' not found: no %s containers available in database for user %d",
						containerTypeName, ref.Name, containerTypeName, userID)
				}
				return nil, fmt.Errorf("%s container '%s' not found (available containers: %v)", containerTypeName, ref.Name, availableContainers)
			}

			if _, exists := hierarchicalMap[ref.Name][ref.Version]; !exists {
				return nil, fmt.Errorf("%s container version not found: %s:%s (available versions for %s: %v)", containerTypeName, ref.Name, ref.Version, ref.Name, getAvailableVersions(hierarchicalMap, ref.Name))
			}

			result = hierarchicalMap[ref.Name][ref.Version]
		} else {
			if _, exists := flatMap[ref.Name]; !exists {
				availableContainers := getAvailableContainerNames(hierarchicalMap)
				if len(availableContainers) == 0 {
					// Check if container exists with different type
					exists, actualType, err := repository.CheckContainerExistsWithDifferentType(db, ref.Name, containerType, userID)
					if err != nil {
						return nil, fmt.Errorf("failed to check container type: %w", err)
					}
					if exists {
						return nil, fmt.Errorf("%s container '%s' not found: container exists but has type '%s', not '%s'",
							containerTypeName, ref.Name, consts.GetContainerTypeName(actualType), containerTypeName)
					}
					return nil, fmt.Errorf("%s container '%s' not found: no %s containers available in database for user %d",
						containerTypeName, ref.Name, containerTypeName, userID)
				}
				return nil, fmt.Errorf("%s container '%s' not found (available containers: %v)", containerTypeName, ref.Name, availableContainers)
			}
			result = flatMap[ref.Name][0]
		}

		results[ref] = result
	}

	return results, nil
}

func getUniqueVersionsForContainerRefsWithDB(db *gorm.DB, refs []*dto.ContainerRef, containerType consts.ContainerType, userID int) ([]model.ContainerVersion, error) {
	containerNamesSet := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if ref.Name != "" {
			containerNamesSet[ref.Name] = struct{}{}
		}
	}

	if len(containerNamesSet) == 0 {
		return []model.ContainerVersion{}, nil
	}

	requiredNames := make([]string, 0, len(containerNamesSet))
	for name := range containerNamesSet {
		requiredNames = append(requiredNames, name)
	}

	versions, err := repository.BatchGetContainerVersions(db, containerType, requiredNames, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get container versions: %w", err)
	}

	return versions, nil
}

func listParameterItemsWithDB(db *gorm.DB, specs []dto.ParameterSpec, fetcher repository.ParameterConfigFetcher, resourceID int, contextCfg any) ([]dto.ParameterItem, error) {
	keys := make([]string, 0, len(specs))
	for _, item := range specs {
		keys = append(keys, item.Key)
	}

	paramConfigs, err := fetcher(db, keys, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %w", err)
	}

	if len(paramConfigs) == 0 && len(specs) > 0 {
		return nil, fmt.Errorf("no configurations found for the provided specs")
	}

	paramConfigMap := make(map[string]model.ParameterConfig, len(paramConfigs))
	for _, config := range paramConfigs {
		paramConfigMap[config.Key] = config
	}

	processedParamConfigs := make(map[string]struct{})

	items := make([]dto.ParameterItem, 0, len(specs))
	for _, spec := range specs {
		config, exists := paramConfigMap[spec.Key]
		if !exists {
			return nil, fmt.Errorf("configuration not found for key: %s", spec.Key)
		}

		processedParamConfigs[spec.Key] = struct{}{}

		item, err := processParameterConfig(config, spec.Value, contextCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to process parameter config for key %s: %w", spec.Key, err)
		}
		if item != nil {
			items = append(items, *item)
		}
	}

	for _, paramConfigMap := range paramConfigMap {
		if _, processed := processedParamConfigs[paramConfigMap.Key]; !processed {
			item, err := processParameterConfig(paramConfigMap, nil, contextCfg)
			if err != nil {
				return nil, fmt.Errorf("failed to process parameter config for key %s: %w", paramConfigMap.Key, err)
			}
			if item != nil {
				items = append(items, *item)
			}
		}
	}

	return items, nil
}

// processParameterConfig processes a single parameter configuration and returns the corresponding parameter item
func processParameterConfig(config model.ParameterConfig, userValue any, contextCfg any) (*dto.ParameterItem, error) {
	switch config.Type {
	case consts.ParameterTypeFixed:
		finalValue := userValue
		if finalValue == nil {
			if config.Required && config.DefaultValue == nil {
				return nil, fmt.Errorf("required fixed parameter %s is missing a value and has no default", config.Key)
			} else if config.DefaultValue != nil {
				convertedValue, err := utils.ConvertStringToSimpleType(*config.DefaultValue)
				if err != nil {
					return nil, fmt.Errorf("failed to convert default value for parameter %s: %w", config.Key, err)
				}
				finalValue = convertedValue
			}
		}

		return &dto.ParameterItem{
			Key:   config.Key,
			Value: finalValue,
		}, nil

	case consts.ParameterTypeDynamic:
		if config.TemplateString == nil || *config.TemplateString == "" {
			return nil, fmt.Errorf("dynamic parameter %s is missing a template string", config.Key)
		}

		templateVars := extractTemplateVars(*config.TemplateString)
		if len(templateVars) == 0 {
			return &dto.ParameterItem{
				Key:            config.Key,
				TemplateString: *config.TemplateString,
			}, nil
		}

		renderedValue, err := renderTemplate(*config.TemplateString, templateVars, contextCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to render dynamic parameter %s: %w", config.Key, err)
		}
		if config.Required && renderedValue == "" {
			return nil, fmt.Errorf("required dynamic parameter %s rendered to an empty string", config.Key)
		}
		if renderedValue != "" {
			return &dto.ParameterItem{
				Key:   config.Key,
				Value: renderedValue,
			}, nil
		}

		return nil, nil
	default:
		return nil, fmt.Errorf("unknown parameter type for key %s", config.Key)
	}
}

// getAvailableContainerNames returns a list of available container names from the hierarchical map
func getAvailableContainerNames(hierarchicalMap map[string]map[string]model.ContainerVersion) []string {
	names := make([]string, 0, len(hierarchicalMap))
	for name := range hierarchicalMap {
		names = append(names, name)
	}
	return names
}

// getAvailableVersions returns a list of available versions for a specific container
func getAvailableVersions(hierarchicalMap map[string]map[string]model.ContainerVersion, containerName string) []string {
	if versions, exists := hierarchicalMap[containerName]; exists {
		versionNames := make([]string, 0, len(versions))
		for versionName := range versions {
			versionNames = append(versionNames, versionName)
		}
		return versionNames
	}
	return []string{}
}
