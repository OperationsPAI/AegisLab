package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aegis/client"
	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	"github.com/sirupsen/logrus"
)

// configHandler defines the internal interface for handling configuration changes
type configHandler interface {
	// handle processes a configuration change
	// key: the full configuration key (e.g., "rate_limiting.max_concurrent_builds")
	// Returns error if the update cannot be applied
	handle(ctx context.Context, key, oldValue, newValue string) error

	// category returns the configuration category this handler is responsible for
	category() string
}

// configRegistry manages configuration handlers
type configRegistry struct {
	mu       sync.RWMutex
	handlers map[string]configHandler
}

var (
	registryInstance *configRegistry
	registryOnce     sync.Once
)

// HandleConfigChange routes a configuration change to the appropriate handler
// This is the main public API for handling config changes
func HandleConfigChange(ctx context.Context, key, oldValue, newValue string) error {
	existingConfig, err := repository.GetConfigByKey(database.DB, key, false)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing config %s from database: %w", key, err)
	}

	r := getConfigRegistry()
	r.mu.RLock()
	handler, exists := r.handlers[existingConfig.Category]
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

	if err := handler.handle(ctx, key, oldValue, newValue); err != nil {
		return fmt.Errorf("handler failed for config %s: %w", key, err)
	}

	return nil
}

// ListRegisteredConfigKeys returns all registered configuration keys
// This is useful for debugging and monitoring
func ListRegisteredConfigKeys() []string {
	r := getConfigRegistry()
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.handlers))
	for key := range r.handlers {
		keys = append(keys, key)
	}
	return keys
}

// =====================================================================
// Built-in Config Handlers Registration
// =====================================================================

// RegisterBuiltinHandlers registers all built-in configuration handlers
// This should be called during consumer initialization
func RegisterBuiltinHandlers() {
	registry := getConfigRegistry()

	// Register handlers using struct implementation
	registry.register(newChaosSystemCountHandler(GetMonitor(), k8s.GetK8sController()))
	registry.register(newRateLimitingConfigHandler(
		GetRestartPedestalRateLimiter(),
		GetBuildContainerRateLimiter(),
		GetAlgoExecutionRateLimiter(),
	))

	// Add more handlers here...

	logrus.Infof("Registered %d built-in config handlers", len(ListRegisteredConfigKeys()))
}

// getConfigRegistry returns the singleton config registry instance (internal use only)
func getConfigRegistry() *configRegistry {
	registryOnce.Do(func() {
		registryInstance = &configRegistry{
			handlers: make(map[string]configHandler),
		}
	})
	return registryInstance
}

// register registers a configuration handler (internal method)
func (r *configRegistry) register(handler configHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	category := handler.category()
	if _, exists := r.handlers[category]; exists {
		logrus.Warnf("Config handler for category %s already registered, overwriting", category)
	}

	r.handlers[category] = handler
	logrus.Debugf("Registered config handler for category: %s", category)
}

// publishWrapper wraps a config update function to publish response to Redis
func publishWrapper(ctx context.Context, function func() error) error {
	updateResponse := dto.NewConfigUpdateResponse()

	defer func() {
		if err := client.RedisPublish(ctx, consts.ConfigUpdateResponseChannel, updateResponse); err != nil {
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

// =====================================================================
// ChaosSystemCountHandler - handles injection.system.count configuration
// =====================================================================

// chaosSystemCountHandler handles chaos system configuration changes
type chaosSystemCountHandler struct {
	monitor    *monitor
	controller *k8s.Controller
}

// UpdateK8sController updates K8s controller informers based on namespace changes
func UpdateK8sController(controller *k8s.Controller, toAdd, toRemove []string) error {
	if controller == nil {
		logrus.Warn("Controller not initialized, skipping informer update")
		return nil
	}

	if len(toAdd) > 0 {
		logrus.Infof("Adding informers for active namespaces: %v", toAdd)
		if err := controller.AddNamespaceInformers(toAdd); err != nil {
			return fmt.Errorf("failed to add namespace informers: %w", err)
		}
	}

	if len(toRemove) > 0 {
		logrus.Infof("Marking namespaces as inactive: %v", toRemove)
		controller.RemoveNamespaceInformers(toRemove)
	}

	return nil
}

// newChaosSystemCountHandler creates a new chaos system count handler with dependencies injected
func newChaosSystemCountHandler(m *monitor, c *k8s.Controller) *chaosSystemCountHandler {
	return &chaosSystemCountHandler{
		monitor:    m,
		controller: c,
	}
}

func (h *chaosSystemCountHandler) category() string {
	return "injection.system.count"
}

func (h *chaosSystemCountHandler) handle(ctx context.Context, key, oldValue, newValue string) error {
	return publishWrapper(ctx, func() error {
		return config.GetChaosSystemConfigManager().Reload(h.onUpdate)
	})
}

// onUpdate is called when chaos system configuration is reloaded
func (h *chaosSystemCountHandler) onUpdate() error {
	logrus.Info("Chaos system configuration updated, refreshing namespaces...")

	if h.monitor == nil {
		logrus.Warn("Monitor not initialized, skipping namespace refresh")
		return nil
	}

	result, err := h.monitor.RefreshNamespaces()
	if err != nil {
		return fmt.Errorf("failed to refresh namespaces: %w", err)
	}

	totalChanges := len(result.Added) + len(result.Recovered) + len(result.Disabled) + len(result.Deleted)
	logrus.Infof("Namespace refresh completed: %d total changes", totalChanges)

	if len(result.Added) > 0 {
		logrus.Infof("Added namespaces: %v", result.Added)
	}
	if len(result.Recovered) > 0 {
		logrus.Infof("Recovered namespaces: %v", result.Recovered)
	}
	if len(result.Disabled) > 0 {
		logrus.Warnf("Disabled namespaces (have active locks): %v", result.Disabled)
	}
	if len(result.Deleted) > 0 {
		logrus.Infof("Deleted namespaces (no active locks): %v", result.Deleted)
	}

	namespacesToAdd := make([]string, 0, len(result.Added)+len(result.Recovered))
	namespacesToAdd = append(namespacesToAdd, result.Added...)
	namespacesToAdd = append(namespacesToAdd, result.Recovered...)

	namespacesToRemove := make([]string, 0, len(result.Disabled)+len(result.Deleted))
	namespacesToRemove = append(namespacesToRemove, result.Disabled...)
	namespacesToRemove = append(namespacesToRemove, result.Deleted...)

	return UpdateK8sController(h.controller, namespacesToAdd, namespacesToRemove)
}

// =====================================================================
// RateLimitingConfigHandler - handles rate_limiting configuration
// =====================================================================

// rateLimitingConfigHandler handles rate limiting configuration changes
type rateLimitingConfigHandler struct {
	restartLimiter *TokenBucketRateLimiter
	buildLimiter   *TokenBucketRateLimiter
	algoLimiter    *TokenBucketRateLimiter
}

// newRateLimitingConfigHandler creates a new rate limiting config handler
func newRateLimitingConfigHandler(
	restartLimiter *TokenBucketRateLimiter,
	buildLimiter *TokenBucketRateLimiter,
	algoLimiter *TokenBucketRateLimiter,
) *rateLimitingConfigHandler {
	return &rateLimitingConfigHandler{
		restartLimiter: restartLimiter,
		buildLimiter:   buildLimiter,
		algoLimiter:    algoLimiter,
	}
}

func (h *rateLimitingConfigHandler) category() string {
	return "rate_limiting"
}

func (h *rateLimitingConfigHandler) handle(ctx context.Context, key, oldValue, newValue string) error {
	return publishWrapper(ctx, func() error {
		logrus.WithFields(logrus.Fields{
			"key":       key,
			"old_value": oldValue,
			"new_value": newValue,
		}).Info("Rate limiting configuration updated, applying changes...")

		switch key {
		case "rate_limiting.max_concurrent_builds":
			if h.buildLimiter != nil {
				maxTokens := config.GetInt(consts.MaxTokensKeyBuildContainer)
				_, currentTimeout := h.buildLimiter.GetConfig()
				h.buildLimiter.UpdateConfig(maxTokens, currentTimeout)
			}

		case "rate_limiting.max_concurrent_restarts":
			if h.restartLimiter != nil {
				maxTokens := config.GetInt(consts.MaxTokensKeyRestartPedestal)
				_, currentTimeout := h.restartLimiter.GetConfig()
				h.restartLimiter.UpdateConfig(maxTokens, currentTimeout)
			}

		case "rate_limiting.max_concurrent_algo_execution":
			if h.algoLimiter != nil {
				maxTokens := config.GetInt(consts.MaxTokensKeyAlgoExecution)
				_, currentTimeout := h.algoLimiter.GetConfig()
				h.algoLimiter.UpdateConfig(maxTokens, currentTimeout)
			}

		case "rate_limiting.token_wait_timeout":
			// Update timeout for all rate limiters
			tokenWaitTimeout := config.GetInt("rate_limiting.token_wait_timeout")
			timeout := time.Duration(tokenWaitTimeout) * time.Second

			if h.restartLimiter != nil {
				maxTokens, _ := h.restartLimiter.GetConfig()
				h.restartLimiter.UpdateConfig(maxTokens, timeout)
			}

			if h.buildLimiter != nil {
				maxTokens, _ := h.buildLimiter.GetConfig()
				h.buildLimiter.UpdateConfig(maxTokens, timeout)
			}

			if h.algoLimiter != nil {
				maxTokens, _ := h.algoLimiter.GetConfig()
				h.algoLimiter.UpdateConfig(maxTokens, timeout)
			}

		default:
			logrus.Warnf("Unknown rate limiting config key: %s, skipping update", key)
			return nil
		}

		logrus.Info("Rate limiting configuration applied successfully")
		return nil
	})
}
