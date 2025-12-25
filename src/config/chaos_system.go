package config

import (
	"fmt"
	"maps"
	"regexp"
	"sync"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

const (
	ConfigKeyChaosSystem = "injection.system"
)

type ChaosSystemConfig struct {
	System         chaos.SystemType
	Count          int    `mapstructure:"count"`
	NsPattern      string `mapstructure:"ns_pattern"`
	ExtractPattern string `mapstructure:"extract_pattern"`
}

type chaosSystemConfigManager struct {
	configs   map[chaos.SystemType]ChaosSystemConfig
	mu        sync.RWMutex
	callbacks []func() error
}

var (
	managerInstance *chaosSystemConfigManager
	managerOnce     sync.Once
)

// GetChaosSystemConfigManager returns the singleton instance of SystemConfigManager
func GetChaosSystemConfigManager() *chaosSystemConfigManager {
	managerOnce.Do(func() {
		managerInstance = &chaosSystemConfigManager{
			configs: make(map[chaos.SystemType]ChaosSystemConfig),
		}
		if err := managerInstance.load(); err != nil {
			logrus.Fatalf("failed to load chaos system config: %v", err)
		}
	})
	return managerInstance
}

// Get returns the configuration for a specific system
func (m *chaosSystemConfigManager) Get(system chaos.SystemType) (ChaosSystemConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, exists := m.configs[system]
	return cfg, exists
}

// GetAll returns all system configurations
func (m *chaosSystemConfigManager) GetAll() map[chaos.SystemType]ChaosSystemConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[chaos.SystemType]ChaosSystemConfig, len(m.configs))
	maps.Copy(result, m.configs)
	return result
}

// Reload reloads system configurations from config
func (m *chaosSystemConfigManager) Reload() error {
	if err := m.load(); err != nil {
		return err
	}

	// Execute all registered callbacks
	m.mu.RLock()
	callbacks := m.callbacks
	m.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(); err != nil {
			return fmt.Errorf("callback execution failed: %w", err)
		}
	}

	return nil
}

// RegisterCallback registers a callback function to be invoked when config is reloaded
func (m *chaosSystemConfigManager) RegisterCallback(callback func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, callback)
}

func (m *chaosSystemConfigManager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := GetMap(ConfigKeyChaosSystem)
	systemConfigMap := make(map[chaos.SystemType]ChaosSystemConfig, len(cfg))

	for sys, c := range cfg {
		var sysCfg ChaosSystemConfig
		if err := mapstructure.Decode(c, &sysCfg); err != nil {
			return fmt.Errorf("failed to decode config for system %s: %w", sys, err)
		}

		system := chaos.SystemType(sys)
		if !system.IsValid() {
			return fmt.Errorf("invalid system type: %s", sys)
		}

		sysCfg.System = system
		systemConfigMap[system] = sysCfg
	}

	m.configs = systemConfigMap
	return nil
}

// ExtractNsPrefixAndNumber extracts the number from a namespace string
// using the system-specific extract pattern from dynamic config
func (s *ChaosSystemConfig) ExtractNsNumber(namespace string) (int, error) {
	if s.ExtractPattern == "" {
		return 0, fmt.Errorf("extract pattern not defined for system %s", s.System)
	}

	re, err := regexp.Compile(s.ExtractPattern)
	if err != nil {
		return 0, fmt.Errorf("invalid extract pattern for system %s: %w", s.System, err)
	}

	matches := re.FindStringSubmatch(namespace)
	if len(matches) >= 3 {
		var number int
		_, err = fmt.Sscanf(matches[2], "%d", &number)
		if err != nil {
			return 0, fmt.Errorf("failed to parse number from namespace '%s': %w", namespace, err)
		}
		return number, nil
	}

	return 0, fmt.Errorf("namespace '%s' does not match extract pattern for system %s", namespace, s.System)
}

// GetAllNamespaces generates a list of all namespaces based on the system count map
func GetAllNamespaces() ([]string, error) {
	manager := GetChaosSystemConfigManager()

	systemConfigMap := manager.GetAll()
	namespaces := make([]string, 0)
	for _, cfg := range systemConfigMap {
		template := convertPatternToTemplate(cfg.NsPattern)
		if template == "" {
			return nil, fmt.Errorf("failed to convert ns_pattern to template: %s", cfg.NsPattern)
		}

		// Generate namespaces from 0 to count-1
		for idx := 0; idx < cfg.Count; idx++ {
			ns := fmt.Sprintf(template, idx)
			namespaces = append(namespaces, ns)
		}
	}

	return namespaces, nil
}

// convertPatternToTemplate converts a regex pattern to a sprintf template
// Convert ns_pattern to a generation template
//
// e.g., "^ts\d+$" -> "ts%d"
//
// e.g., "^app-\d+$" -> "app-%d"
//
// e.g., "^test_\d+_suffix$" -> "test_%d_suffix"
func convertPatternToTemplate(pattern string) string {
	// Remove anchors ^ and $
	template := pattern
	if len(template) > 0 && template[0] == '^' {
		template = template[1:]
	}
	if len(template) > 0 && template[len(template)-1] == '$' {
		template = template[:len(template)-1]
	}

	// Replace \d+ with %d
	template = regexp.MustCompile(`\\d\+`).ReplaceAllString(template, "%d")

	return template
}
