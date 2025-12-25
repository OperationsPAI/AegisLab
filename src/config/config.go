package config

import (
	"aegis/consts"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init Initialize configuration
func Init(configPath string) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	viper.SetConfigName("config." + env)
	viper.SetConfigType("toml")

	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	viper.AddConfigPath("$HOME/.rcabench")
	viper.AddConfigPath("/etc/rcabench")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		configFile := viper.ConfigFileUsed()
		content, readErr := os.ReadFile(configFile)

		if readErr != nil {
			logrus.Errorf("Failed to read config file content: %v", readErr)
		} else {
			logrus.Errorf("Config file original content:\n%s", string(content))
		}

		if parseErr, ok := err.(*viper.ConfigParseError); ok {
			logrus.Fatalf("Config file parsing failed: %v\nDetails: %v", parseErr, parseErr.Error())
		} else {
			logrus.Fatalf("Failed to read config file: %v", err)
		}
	}

	logrus.Printf("Config file loaded successfully: %v; configPath: %v, ", viper.ConfigFileUsed(), configPath)

	// Automatically bind environment variables
	viper.AutomaticEnv()

	// Validate configuration
	if err := validate(); err != nil {
		logrus.Fatalf("Configuration validation failed: %v", err)
	}
	logrus.Info("Configuration validation passed")
}

// Get Get configuration item value
func Get(key string) any {
	return viper.Get(key)
}

// GetString Get string type configuration item
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt Get integer type configuration item
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool Get boolean type configuration item
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetFloat64 Get float64 type configuration item
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetStringSlice Get string slice type configuration item
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetIntSlice Get integer slice type configuration item
func GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}

// GetMap Get map type configuration item
func GetMap(key string) map[string]any {
	return viper.GetStringMap(key)
}

// GetList Get any list type configuration item
func GetList(key string) []any {
	value := viper.Get(key)
	if value == nil {
		return nil
	}
	if list, ok := value.([]any); ok {
		return list
	}
	return nil
}

// SetViperValue sets a value in viper based on the value type
func SetViperValue(key, value string, valueType consts.ConfigValueType) error {
	switch valueType {
	case consts.ConfigValueTypeString:
		viper.Set(key, value)

	case consts.ConfigValueTypeBool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for %s: %w", key, err)
		}
		viper.Set(key, boolVal)

	case consts.ConfigValueTypeInt:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for %s: %w", key, err)
		}
		viper.Set(key, intVal)

	case consts.ConfigValueTypeFloat:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value for %s: %w", key, err)
		}
		viper.Set(key, floatVal)

	case consts.ConfigValueTypeStringArray:
		// Parse JSON array
		var strSlice []string
		if err := json.Unmarshal([]byte(value), &strSlice); err != nil {
			// Fallback to comma-separated values
			strSlice = strings.Split(value, ",")
			for i := range strSlice {
				strSlice[i] = strings.TrimSpace(strSlice[i])
			}
		}
		viper.Set(key, strSlice)

	default:
		return fmt.Errorf("unsupported value type %d for key %s", valueType, key)
	}

	return nil
}

// validate validates the configuration
func validate() error {
	// Required fields validation
	requiredFields := []string{
		"name",
		"version",
		"port",
		"workspace",
	}

	for _, field := range requiredFields {
		if !viper.IsSet(field) {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// Validate port range
	port := viper.GetInt("port")
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1-65535)", port)
	}

	// Database configuration
	mysqlFields := []string{
		"database.mysql.host",
		"database.mysql.port",
		"database.mysql.user",
		"database.mysql.password",
		"database.mysql.db",
	}
	for _, field := range mysqlFields {
		if !viper.IsSet(field) {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// Redis configuration
	if !viper.IsSet("redis.host") {
		return fmt.Errorf("required field 'redis.host' is missing")
	}

	// Jaeger configuration
	if !viper.IsSet("jaeger.endpoint") {
		return fmt.Errorf("required field 'jaeger.endpoint' is missing")
	}

	// Kubernetes configuration
	k8sFields := []string{
		"k8s.namespace",
		"k8s.job.service_account.name",
	}
	for _, field := range k8sFields {
		if !viper.IsSet(field) {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// BuildKit configuration
	if !viper.IsSet("buildkit.address") {
		return fmt.Errorf("required field 'buildkit.address' is missing")
	}

	logrus.Debug("All required configuration fields are present and valid")
	return nil
}
