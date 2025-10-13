package utils

import (
	"aegis/config"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

func GetValidBenchmarkMap() map[string]struct{} {
	benchmarks := config.GetStringSlice("injection.benchmark")
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
	m := config.GetMap("injection.namespace_config")
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
	m := config.GetMap("injection.namespace_config")
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

func CheckNsPrefixExists(nsPrefix string) bool {
	m := config.GetMap("injection.namespace_config")
	_, exists := m[nsPrefix]
	return exists
}
