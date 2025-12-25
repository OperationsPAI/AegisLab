package common

import (
	"aegis/consts"
	"aegis/database"
	"aegis/repository"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"gorm.io/gorm"
)

// configTypeConstraints defines which metadata fields are applicable for each value type
type configTypeConstraints struct {
	supportsMinMax  bool
	supportsPattern bool
	supportsOptions bool
}

// configTypeRules defines validation rules for each config value type
var configTypeRules = map[consts.ConfigValueType]configTypeConstraints{
	consts.ConfigValueTypeBool: {
		supportsMinMax:  false,
		supportsPattern: false,
		supportsOptions: true,
	},
	consts.ConfigValueTypeInt: {
		supportsMinMax:  true,
		supportsPattern: false,
		supportsOptions: true,
	},
	consts.ConfigValueTypeFloat: {
		supportsMinMax:  true,
		supportsPattern: false,
		supportsOptions: true,
	},
	consts.ConfigValueTypeString: {
		supportsMinMax:  false,
		supportsPattern: true,
		supportsOptions: true,
	},
	consts.ConfigValueTypeStringArray: {
		supportsMinMax:  false,
		supportsPattern: false,
		supportsOptions: false,
	},
}

// CreateConfig creates a new configuration with history tracking
func CreateConfig(db *gorm.DB, config *database.DynamicConfig) error {
	if err := repository.CreateConfig(db, config); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: configuration with key '%s' already exists", consts.ErrAlreadyExists, config.Key)
		}
		return fmt.Errorf("failed to create config: %w", err)
	}
	return nil
}

// ValidateConfig validates a configuration against its type and constraints
func ValidateConfig(cfg *database.DynamicConfig) error {
	// Validate metadata constraints
	if err := validateConfigMetadataConstraints(cfg); err != nil {
		return err
	}

	// Validate the value itself
	switch cfg.ValueType {
	case consts.ConfigValueTypeBool:
		if _, err := strconv.ParseBool(cfg.Value); err != nil {
			return fmt.Errorf("invalid boolean value: %s", cfg.Value)
		}

	case consts.ConfigValueTypeInt:
		intVal, err := strconv.ParseInt(cfg.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", cfg.Value)
		}
		if cfg.MinValue != nil && float64(intVal) < *cfg.MinValue {
			return fmt.Errorf("value %d is below minimum %v", intVal, *cfg.MinValue)
		}
		if cfg.MaxValue != nil && float64(intVal) > *cfg.MaxValue {
			return fmt.Errorf("value %d exceeds maximum %v", intVal, *cfg.MaxValue)
		}

	case consts.ConfigValueTypeFloat:
		floatVal, err := strconv.ParseFloat(cfg.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", cfg.Value)
		}
		if cfg.MinValue != nil && floatVal < *cfg.MinValue {
			return fmt.Errorf("value %f is below minimum %v", floatVal, *cfg.MinValue)
		}
		if cfg.MaxValue != nil && floatVal > *cfg.MaxValue {
			return fmt.Errorf("value %f exceeds maximum %v", floatVal, *cfg.MaxValue)
		}

	case consts.ConfigValueTypeString:
		if cfg.Pattern != "" {
			matched, err := regexp.MatchString(cfg.Pattern, cfg.Value)
			if err != nil {
				return fmt.Errorf("invalid regex pattern: %w", err)
			}
			if !matched {
				return fmt.Errorf("value does not match required pattern")
			}
		}

	case consts.ConfigValueTypeStringArray:
		var strArray []string
		if err := json.Unmarshal([]byte(cfg.Value), &strArray); err != nil {
			return fmt.Errorf("invalid string array format: %w", err)
		}
	}

	// Validate against allowed options if defined
	if cfg.Options != "" {
		if err := validateConfigOptions(cfg); err != nil {
			return err
		}
	}

	return nil
}

// validateConfigMetadataConstraints validates that metadata fields are appropriate for the value type
func validateConfigMetadataConstraints(cfg *database.DynamicConfig) error {
	rules, exists := configTypeRules[cfg.ValueType]
	if !exists {
		return fmt.Errorf("unknown value type: %d", cfg.ValueType)
	}

	// Check MinValue/MaxValue constraints
	if !rules.supportsMinMax {
		if cfg.MinValue != nil {
			return fmt.Errorf("min_value is not applicable for %s type", consts.GetDynamicConfigTypeName(cfg.ValueType))
		}
		if cfg.MaxValue != nil {
			return fmt.Errorf("max_value is not applicable for %s type", consts.GetDynamicConfigTypeName(cfg.ValueType))
		}
	}

	// Check Pattern constraints
	if !rules.supportsPattern && cfg.Pattern != "" {
		return fmt.Errorf("pattern is not applicable for %s type", consts.GetDynamicConfigTypeName(cfg.ValueType))
	}

	// Check Options constraints
	if !rules.supportsOptions && cfg.Options != "" {
		return fmt.Errorf("options is not applicable for %s type", consts.GetDynamicConfigTypeName(cfg.ValueType))
	}

	return nil
}

// validateConfigOptions validates the config value against allowed options based on value type
func validateConfigOptions(cfg *database.DynamicConfig) error {
	switch cfg.ValueType {
	case consts.ConfigValueTypeString:
		var allowedOptions []string
		if err := json.Unmarshal([]byte(cfg.Options), &allowedOptions); err != nil {
			return fmt.Errorf("invalid options format (expected []string): %w", err)
		}
		if !slices.Contains(allowedOptions, cfg.Value) {
			return fmt.Errorf("value '%s' is not in allowed options: %v", cfg.Value, allowedOptions)
		}

	case consts.ConfigValueTypeInt:
		var allowedOptions []int64
		if err := json.Unmarshal([]byte(cfg.Options), &allowedOptions); err != nil {
			return fmt.Errorf("invalid options format (expected []int): %w", err)
		}
		intVal, _ := strconv.ParseInt(cfg.Value, 10, 64)
		if !slices.Contains(allowedOptions, intVal) {
			return fmt.Errorf("value '%d' is not in allowed options: %v", intVal, allowedOptions)
		}

	case consts.ConfigValueTypeFloat:
		var allowedOptions []float64
		if err := json.Unmarshal([]byte(cfg.Options), &allowedOptions); err != nil {
			return fmt.Errorf("invalid options format (expected []float64): %w", err)
		}
		floatVal, _ := strconv.ParseFloat(cfg.Value, 64)
		if !slices.Contains(allowedOptions, floatVal) {
			return fmt.Errorf("value '%f' is not in allowed options: %v", floatVal, allowedOptions)
		}

	case consts.ConfigValueTypeBool:
		var allowedOptions []bool
		if err := json.Unmarshal([]byte(cfg.Options), &allowedOptions); err != nil {
			return fmt.Errorf("invalid options format (expected []bool): %w", err)
		}
		boolVal, _ := strconv.ParseBool(cfg.Value)
		if !slices.Contains(allowedOptions, boolVal) {
			return fmt.Errorf("value '%v' is not in allowed options: %v", boolVal, allowedOptions)
		}

	case consts.ConfigValueTypeStringArray:
		return fmt.Errorf("options field is not applicable for string array type")
	}

	return nil
}
