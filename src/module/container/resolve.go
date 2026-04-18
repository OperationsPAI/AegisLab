package containermodule

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var templateVarRegex = regexp.MustCompile(`{{\s*\.([a-zA-Z0-9_]+)\s*}}`)

func (r *Repository) ListContainerVersionEnvVars(specs []dto.ParameterSpec, version *model.ContainerVersion) ([]dto.ParameterItem, error) {
	return listParameterItemsWithDB(r, specs, r.listContainerVersionEnvVars, version.ID, version)
}

func (r *Repository) ListHelmConfigValues(specs []dto.ParameterSpec, cfg *model.HelmConfig) ([]dto.ParameterItem, error) {
	return listParameterItemsWithDB(r, specs, r.listHelmConfigValues, cfg.ID, cfg.ContainerVersion)
}

func (r *Repository) ResolveContainerVersions(refs []*dto.ContainerRef, containerType consts.ContainerType, userID int) (map[*dto.ContainerRef]model.ContainerVersion, error) {
	versions, err := getUniqueVersionsForContainerRefs(r, refs, containerType, userID)
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
					exists, actualType, err := r.checkContainerExistsWithDifferentType(ref.Name, containerType, userID)
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
					exists, actualType, err := r.checkContainerExistsWithDifferentType(ref.Name, containerType, userID)
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

func getUniqueVersionsForContainerRefs(repo *Repository, refs []*dto.ContainerRef, containerType consts.ContainerType, userID int) ([]model.ContainerVersion, error) {
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
	return repo.batchGetContainerVersions(containerType, requiredNames, userID)
}

func listParameterItemsWithDB(repo *Repository, specs []dto.ParameterSpec, fetcher func([]string, int) ([]model.ParameterConfig, error), resourceID int, contextCfg any) ([]dto.ParameterItem, error) {
	keys := make([]string, 0, len(specs))
	for _, item := range specs {
		keys = append(keys, item.Key)
	}

	paramConfigs, err := fetcher(keys, resourceID)
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

	for _, paramConfig := range paramConfigMap {
		if _, processed := processedParamConfigs[paramConfig.Key]; processed {
			continue
		}
		item, err := processParameterConfig(paramConfig, nil, contextCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to process parameter config for key %s: %w", paramConfig.Key, err)
		}
		if item != nil {
			items = append(items, *item)
		}
	}

	return items, nil
}

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
		return &dto.ParameterItem{Key: config.Key, Value: finalValue}, nil
	case consts.ParameterTypeDynamic:
		if config.TemplateString == nil || *config.TemplateString == "" {
			return nil, fmt.Errorf("dynamic parameter %s is missing a template string", config.Key)
		}
		templateVars := extractTemplateVars(*config.TemplateString)
		if len(templateVars) == 0 {
			return &dto.ParameterItem{Key: config.Key, TemplateString: *config.TemplateString}, nil
		}

		renderedValue, err := renderTemplate(*config.TemplateString, templateVars, contextCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to render dynamic parameter %s: %w", config.Key, err)
		}
		if config.Required && renderedValue == "" {
			return nil, fmt.Errorf("required dynamic parameter %s rendered to an empty string", config.Key)
		}
		if renderedValue != "" {
			return &dto.ParameterItem{Key: config.Key, Value: renderedValue}, nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported parameter type: %v", config.Type)
	}
}

func extractTemplateVars(templateString string) []string {
	matches := templateVarRegex.FindAllStringSubmatch(templateString, -1)
	if matches == nil {
		return nil
	}

	variables := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}
	return variables
}

func renderTemplate(templateStr string, vars []string, context any) (string, error) {
	contextValue := reflect.ValueOf(context)
	if contextValue.Kind() == reflect.Ptr {
		contextValue = contextValue.Elem()
	}

	renderedString := templateStr
	contextType := contextValue.Type()
	for _, varName := range vars {
		fieldValue := contextValue.FieldByName(varName)
		if !fieldValue.IsValid() {
			return "", fmt.Errorf("variable '%s' not found in context structure", varName)
		}

		fieldType, found := contextType.FieldByName(varName)
		if !found || fieldType.PkgPath != "" {
			return "", fmt.Errorf("variable '%s' is not an exported field in context", varName)
		}

		strValue, err := utils.ConvertSimpleTypeToString(fieldValue.Interface())
		if err != nil {
			return "", fmt.Errorf("failed to convert context value for %s: %w", varName, err)
		}

		renderedString = strings.ReplaceAll(renderedString, fmt.Sprintf("{{ .%s }}", varName), strValue)
		renderedString = strings.ReplaceAll(renderedString, fmt.Sprintf("{{.%s}}", varName), strValue)
	}
	return renderedString, nil
}

func getAvailableContainerNames(versions map[string]map[string]model.ContainerVersion) []string {
	names := make([]string, 0, len(versions))
	for name := range versions {
		names = append(names, name)
	}
	return names
}

func getAvailableVersions(versions map[string]map[string]model.ContainerVersion, containerName string) []string {
	items := versions[containerName]
	results := make([]string, 0, len(items))
	for version := range items {
		results = append(results, version)
	}
	return results
}
