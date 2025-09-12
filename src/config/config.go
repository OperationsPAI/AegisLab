package config

import (
	"fmt"
	"os"
	"sort"

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
	logrus.Info(viper.AllSettings())
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

func GetValidBenchmarkMap() map[string]struct{} {
	benchmarks := GetStringSlice("injection.benchmark")
	if len(benchmarks) == 0 {
		logrus.Warn("No benchmarks configured, using default 'clickhouse'")
		benchmarks = []string{"clickhouse"}
	}

	benchmarkMap := make(map[string]struct{}, len(benchmarks))
	for _, benchmark := range benchmarks {
		if benchmark == "" {
			logrus.Warn("Empty benchmark name found, skipping")
			continue
		}

		benchmarkMap[benchmark] = struct{}{}
	}

	return benchmarkMap
}

func GetNsConfigMap() (map[string]map[string]any, error) {
	m := GetMap("injection.namespace_config")
	nsConfigMap := make(map[string]map[string]any, len(m))
	for ns, c := range m {
		config, ok := c.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid namespace config for %s", ns)
		}

		nsConfigMap[ns] = config
	}

	return nsConfigMap, nil
}

func GetNsCountMap() (map[string]int, error) {
	nsConfigMap, err := GetNsConfigMap()
	if err != nil {
		return nil, err
	}

	nsCountMap := make(map[string]int, len(nsConfigMap))
	for ns, config := range nsConfigMap {
		value, exists := config["count"]
		if !exists {
			return nil, fmt.Errorf("namespace %s does not have a count field", ns)
		}

		vInt, ok := value.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid namespace value for %s", ns)
		}

		nsCountMap[ns] = int(vInt)
	}

	return nsCountMap, nil
}

func GetNsPrefixs() []string {
	m := GetMap("injection.namespace_config")
	nsPrefixs := make([]string, 0, len(m))
	for ns := range m {
		nsPrefixs = append(nsPrefixs, ns)
	}

	sort.Strings(nsPrefixs)
	return nsPrefixs
}

func GetAllNamespaces() ([]string, error) {
	nsCountMap, err := GetNsCountMap()
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(nsCountMap))
	for ns, count := range nsCountMap {
		for idx := range count {
			namespaces = append(namespaces, fmt.Sprintf("%s%d", ns, idx))
		}
	}

	return namespaces, nil
}
