package common

import (
	"aegis/utils"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var templateVarRegex = regexp.MustCompile(`{{\s*\.([a-zA-Z0-9_]+)\s*}}`)

// extractTemplateVars extracts all variable names used in the template string
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

// renderTemplate renders the template string by replacing variables with values from the context structure
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
