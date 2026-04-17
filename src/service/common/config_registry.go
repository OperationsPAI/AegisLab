package common

import (
	"context"
	"fmt"
	"sync"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ConfigHandler defines the interface for handling configuration changes.
// Implement this interface in any package and register via RegisterHandler.
type ConfigHandler interface {
	// Category returns the configuration category this handler is responsible for.
	Category() string

	// Scope returns the configuration scope (global or consumer) that this handler manages.
	Scope() consts.ConfigScope

	// Handle processes a configuration change.
	Handle(ctx context.Context, key, oldValue, newValue string) error
}

// configRegistry manages configuration handlers
type configRegistry struct {
	mu       sync.RWMutex
	handlers map[consts.ConfigScope]map[string]ConfigHandler
}

type ConfigPublisher interface {
	Publish(ctx context.Context, channel string, message any) error
}

var registryInstance = newConfigRegistry()

// RegisterHandler registers a configuration handler.
// External packages (e.g. consumer) call this to plug in their own handlers.
func RegisterHandler(handler ConfigHandler) {
	getConfigRegistry().register(handler)
}

// RegisterGlobalHandlers registers handlers for global-scope configurations.
// Safe to call multiple times — duplicate registrations are skipped.
func RegisterGlobalHandlers(publisher ConfigPublisher) {
	registry := getConfigRegistry()
	if registry.ensureRegistered(&algoConfigHandler{publisher: publisher}) {
		scope := consts.ConfigScopeGlobal
		logrus.Infof("Registered %d global config handler(s)", len(ListRegisteredConfigKeys(&scope)))
	}
}

// ListRegisteredConfigKeys returns all registered configuration keys for informational purposes (e.g. logging)
func ListRegisteredConfigKeys(scope *consts.ConfigScope) []string {
	r := getConfigRegistry()
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0)
	for s, categoryMap := range r.handlers {
		if scope != nil && s != *scope {
			continue
		}
		for category := range categoryMap {
			keys = append(keys, category)
		}
	}
	return keys
}

// PublishWrapper wraps a config update function and publishes the result to Redis.
// Exported so consumer and producer can reuse it in their own handlers.
func PublishWrapper(ctx context.Context, publisher ConfigPublisher, function func() error) error {
	updateResponse := dto.NewConfigUpdateResponse()

	defer func() {
		if publisher == nil {
			return
		}
		if err := publisher.Publish(ctx, consts.ConfigUpdateResponseChannel, updateResponse); err != nil {
			logrus.Errorf("failed to publish config update response to Redis: %v", err)
		}
	}()

	if err := function(); err != nil {
		updateResponse.Error = fmt.Sprintf("error applying config update: %v", err)
		return fmt.Errorf("error applying config update: %w", err)
	}

	updateResponse.Success = true
	return nil
}

// getConfigRegistry returns the singleton config registry instance
func getConfigRegistry() *configRegistry {
	return registryInstance
}

func newConfigRegistry() *configRegistry {
	return &configRegistry{
		handlers: make(map[consts.ConfigScope]map[string]ConfigHandler),
	}
}

func resetConfigRegistryForTest() {
	registryInstance = newConfigRegistry()
}

// handleConfigChange routes a configuration change to the appropriate handler
func handleConfigChange(ctx context.Context, db *gorm.DB, key, oldValue, newValue string) error {
	existingConfig, err := newConfigStore(db).getConfigByKey(key)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing config %s from database: %w", key, err)
	}

	r := getConfigRegistry()
	r.mu.RLock()
	handler, exists := r.handlers[existingConfig.Scope][existingConfig.Category]
	r.mu.RUnlock()

	if !exists {
		logrus.Warnf("no specific handler for config %s, using generic viper update", key)
		return config.SetViperValue(key, newValue, existingConfig.ValueType)
	}

	logrus.WithFields(logrus.Fields{
		"key":       key,
		"old_value": oldValue,
		"new_value": newValue,
	}).Info("Applying config change via registered handler")

	if err := config.SetViperValue(key, newValue, existingConfig.ValueType); err != nil {
		logrus.Warnf("failed to update viper for config %s: %v", key, err)
	}

	if err := handler.Handle(ctx, key, oldValue, newValue); err != nil {
		return fmt.Errorf("handler failed for config %s: %w", key, err)
	}

	return nil
}

// register registers a configuration handler (internal method)
func (r *configRegistry) register(handler ConfigHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	scope := handler.Scope()
	category := handler.Category()

	if _, ok := r.handlers[scope]; !ok {
		r.handlers[scope] = make(map[string]ConfigHandler)
	}
	if _, exists := r.handlers[scope][category]; exists {
		logrus.Debugf("Config handler for scope=%s category=%s already registered, overwriting",
			consts.GetConfigScopeName(scope), category)
	}

	r.handlers[scope][category] = handler
	logrus.Debugf("Registered config handler for scope=%s category=%s",
		consts.GetConfigScopeName(scope), category)
}

func (r *configRegistry) ensureRegistered(handler ConfigHandler) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	scope := handler.Scope()
	category := handler.Category()

	if _, ok := r.handlers[scope]; !ok {
		r.handlers[scope] = make(map[string]ConfigHandler)
	}
	if _, exists := r.handlers[scope][category]; exists {
		return false
	}

	r.handlers[scope][category] = handler
	logrus.Debugf("Registered config handler for scope=%s category=%s",
		consts.GetConfigScopeName(scope), category)
	return true
}

// =====================================================================
// AlgoConfigHandler - handles algo configuration (e.g. algo.detector)
// =====================================================================

type algoConfigHandler struct {
	publisher ConfigPublisher
}

func (h *algoConfigHandler) Category() string { return "algo" }

func (h *algoConfigHandler) Scope() consts.ConfigScope { return consts.ConfigScopeGlobal }

func (h *algoConfigHandler) Handle(ctx context.Context, key, oldValue, newValue string) error {
	return PublishWrapper(ctx, h.publisher, func() error {
		switch key {
		case consts.DetectorKey:
			config.SetDetectorName(newValue)
			logrus.WithFields(logrus.Fields{
				"key":       key,
				"old_value": oldValue,
				"new_value": newValue,
			}).Info("Detector algorithm name updated")
		default:
			logrus.Warnf("Unknown algo config key: %s, skipping update", key)
		}
		return nil
	})
}
