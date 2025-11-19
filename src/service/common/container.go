package common

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"
	"fmt"
)

// ListContainerVersionEnvVars retrieves and validates environment variables for a container version based on provided specs
func ListContainerVersionEnvVars(specs []dto.ParameterSpec, version *database.ContainerVersion) ([]dto.ParameterItem, error) {
	return listParameterItems(specs, repository.ListContainerVersionEnvVars, version.ID, version)
}

// ListHelmConfigValues retrieves and validates Helm values based on provided specs and Helm configuration
func ListHelmConfigValues(specs []dto.ParameterSpec, cfg *database.HelmConfig) ([]dto.ParameterItem, error) {
	return listParameterItems(specs, repository.ListHelmConfigValues, cfg.ID, cfg.ContainerVersion)
}

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

// listParameterItems retrieves and validates parameter items based on provided specs and a parameter config fetcher
func listParameterItems(specs []dto.ParameterSpec, fetcher repository.ParameterConfigFetcher, resourceID int, contextCfg any) ([]dto.ParameterItem, error) {
	keys := make([]string, 0, len(specs))
	for _, item := range specs {
		keys = append(keys, item.Key)
	}

	paramConfigs, err := fetcher(database.DB, keys, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %w", err)
	}

	if len(paramConfigs) == 0 && len(specs) > 0 {
		return nil, fmt.Errorf("no configurations found for the provided specs")
	}

	paramConfigMap := make(map[string]database.ParameterConfig, len(paramConfigs))
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
func processParameterConfig(config database.ParameterConfig, userValue any, contextCfg any) (*dto.ParameterItem, error) {
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

		templateVars := utils.ExtractTemplateVars(*config.TemplateString)
		if len(templateVars) == 0 {
			return &dto.ParameterItem{
				Key:            config.Key,
				TemplateString: *config.TemplateString,
			}, nil
		}

		renderedValue, err := utils.RenderTemplate(*config.TemplateString, templateVars, contextCfg)
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
