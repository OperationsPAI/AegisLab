package config

import (
	"fmt"
	"os"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init 初始化配置
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
			logrus.Errorf("读取配置文件内容失败: %v", readErr)
		} else {
			logrus.Errorf("配置文件原始内容:\n%s", string(content))
		}

		if parseErr, ok := err.(*viper.ConfigParseError); ok {
			logrus.Fatalf("配置文件解析失败: %v\n详细信息: %v", parseErr, parseErr.Error())
		} else {
			logrus.Fatalf("读取配置文件失败: %v", err)
		}
	}

	logrus.Printf("配置文件加载成功: %v; configPath: %v, ", viper.ConfigFileUsed(), configPath)

	// 自动绑定环境变量
	viper.AutomaticEnv()
	logrus.Info(viper.AllSettings())
}

// Get 获取配置项的值
func Get(key string) any {
	return viper.Get(key)
}

// GetString 获取字符串类型的配置项
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt 获取整数类型的配置项
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool 获取布尔类型的配置项
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetFloat64 获取浮点数类型的配置项
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetStringSlice 获取字符串列表类型的配置项
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetIntSlice 获取整数列表类型的配置项
func GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}

// GetMap 获取映射类型的配置项
func GetMap(key string) map[string]any {
	return viper.GetStringMap(key)
}

// GetList 获取任意列表类型的配置项
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
