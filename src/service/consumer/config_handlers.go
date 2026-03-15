package consumer

import (
	"context"
	"fmt"
	"time"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/service/common"

	"github.com/sirupsen/logrus"
)

// RegisterConsumerHandlers registers all consumer-scoped configuration handlers.
// Should be called during consumer initialization, after RegisterGlobalHandlers.
func RegisterConsumerHandlers() {
	scope := consts.ConfigScopeConsumer
	common.RegisterHandler(newChaosSystemCountHandler(GetMonitor(), k8s.GetK8sController()))
	common.RegisterHandler(newRateLimitingConfigHandler(
		GetRestartPedestalRateLimiter(),
		GetBuildContainerRateLimiter(),
		GetAlgoExecutionRateLimiter(),
	))
	logrus.Infof("Registered consumer config handlers: %v", common.ListRegisteredConfigKeys(&scope))
}

// UpdateK8sController updates K8s controller informers based on namespace changes.
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

// =====================================================================
// ChaosSystemCountHandler - handles injection.system.count configuration
// =====================================================================

type chaosSystemCountHandler struct {
	monitor    *monitor
	controller *k8s.Controller
}

func newChaosSystemCountHandler(m *monitor, c *k8s.Controller) *chaosSystemCountHandler {
	return &chaosSystemCountHandler{monitor: m, controller: c}
}

func (h *chaosSystemCountHandler) Category() string          { return "injection.system.count" }
func (h *chaosSystemCountHandler) Scope() consts.ConfigScope { return consts.ConfigScopeConsumer }

func (h *chaosSystemCountHandler) Handle(ctx context.Context, key, oldValue, newValue string) error {
	return common.PublishWrapper(ctx, func() error {
		return config.GetChaosSystemConfigManager().Reload(h.onUpdate)
	})
}

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

type rateLimitingConfigHandler struct {
	restartLimiter *TokenBucketRateLimiter
	buildLimiter   *TokenBucketRateLimiter
	algoLimiter    *TokenBucketRateLimiter
}

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

func (h *rateLimitingConfigHandler) Category() string          { return "rate_limiting" }
func (h *rateLimitingConfigHandler) Scope() consts.ConfigScope { return consts.ConfigScopeConsumer }

func (h *rateLimitingConfigHandler) Handle(ctx context.Context, key, oldValue, newValue string) error {
	return common.PublishWrapper(ctx, func() error {
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
		}

		logrus.Info("Rate limiting configuration applied successfully")
		return nil
	})
}
