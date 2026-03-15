package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aegis/client"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/repository"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// scopePrefix maps configuration scopes to their etcd key prefix.
// Scopes absent from the map (e.g. producer) have no etcd representation.
var scopePrefix = map[consts.ConfigScope]string{
	consts.ConfigScopeProducer: consts.ConfigEtcdProducerPrefix,
	consts.ConfigScopeConsumer: consts.ConfigEtcdConsumerPrefix,
	consts.ConfigScopeGlobal:   consts.ConfigEtcdGlobalPrefix,
}

// configUpdateListener listens for configuration update events from etcd.
// It supports incremental scope activation via EnsureScope — each scope is
// loaded and watched independently, making it safe for both, producer-only
// and consumer-only modes.
type configUpdateListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	active map[consts.ConfigScope]bool // scopes already loaded + watched
}

var (
	configListenerInstance *configUpdateListener
	configListenerOnce     sync.Once
)

// GetConfigUpdateListener returns the singleton instance of configUpdateListener
func GetConfigUpdateListener(ctx context.Context) *configUpdateListener {
	configListenerOnce.Do(func() {
		listenerCtx, cancel := context.WithCancel(ctx)
		configListenerInstance = &configUpdateListener{
			ctx:    listenerCtx,
			cancel: cancel,
			active: make(map[consts.ConfigScope]bool),
		}

		go func() {
			<-ctx.Done()
			logrus.Info("Parent context cancelled, stopping config update listener...")
			configListenerInstance.Stop()
		}()
	})
	return configListenerInstance
}

// EnsureScope loads initial config values from etcd and starts a watcher for
// the given scope. The call is idempotent — invoking it multiple times for the
// same scope is a safe no-op. Scopes without an etcd prefix (e.g. producer)
// are silently skipped.
func (l *configUpdateListener) EnsureScope(scope consts.ConfigScope) error {
	prefix, ok := scopePrefix[scope]
	if !ok {
		logrus.Debugf("Scope %s has no etcd prefix, skipping listener setup",
			consts.GetConfigScopeName(scope))
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.active[scope] {
		return nil
	}

	scopeName := consts.GetConfigScopeName(scope)

	// Load initial values from etcd into viper
	if err := l.loadScopeFromEtcd(scope, prefix, scopeName); err != nil {
		return fmt.Errorf("failed to load %s configs from etcd: %w", scopeName, err)
	}

	// Start a dedicated watcher goroutine for this scope
	go l.watchPrefix(prefix, scopeName)

	l.active[scope] = true
	logrus.Infof("Config listener active for scope %s (prefix=%s)", scopeName, prefix)
	return nil
}

// Stop cancels the listener context, stopping all watcher goroutines.
func (l *configUpdateListener) Stop() {
	l.cancel()
	logrus.Info("Config update listener stopped")
}

// loadScopeFromEtcd loads all configs for a given scope from etcd into viper.
// Falls back to MySQL defaults only if config doesn't exist in etcd.
func (l *configUpdateListener) loadScopeFromEtcd(scope consts.ConfigScope, prefix, scopeName string) error {
	configMetadata, err := repository.ListConfigByScope(database.DB, scope)
	if err != nil {
		return fmt.Errorf("failed to list %s config metadata from database: %w", scopeName, err)
	}

	loadedCount := 0
	initializedCount := 0

	for _, meta := range configMetadata {
		etcdKey := fmt.Sprintf("%s%s", prefix, meta.Key)

		// Try to get current value from etcd first
		etcdValue, err := client.EtcdGet(l.ctx, etcdKey)
		if err != nil {
			logrus.Errorf("Failed to get config %s from etcd: %v", meta.Key, err)
			continue
		}

		var valueToLoad string
		if etcdValue == "" {
			// Config doesn't exist in etcd, initialize it with MySQL default value
			if err := client.EtcdPut(l.ctx, etcdKey, meta.DefaultValue, 0); err != nil {
				logrus.Errorf("Failed to initialize config %s in etcd: %v", meta.Key, err)
				continue
			}

			valueToLoad = meta.DefaultValue
			initializedCount++
			logrus.Infof("Initialized config %s in etcd with default value from MySQL", meta.Key)
		} else {
			valueToLoad = etcdValue
		}

		// Load config to Viper (local memory cache)
		if err := config.SetViperValue(meta.Key, valueToLoad, meta.ValueType); err != nil {
			logrus.Errorf("Failed to load config %s to Viper: %v", meta.Key, err)
			continue
		}
		loadedCount++
	}

	logrus.Infof("Loaded %d/%d %s configs from etcd to Viper (initialized %d new configs)",
		loadedCount, len(configMetadata), scopeName, initializedCount)

	return nil
}

// watchPrefix watches a single etcd prefix for configuration changes.
// Each scope gets its own goroutine calling this method.
func (l *configUpdateListener) watchPrefix(prefix, scopeName string) {
	watchChan := client.EtcdWatch(l.ctx, prefix, true)
	logrus.Infof("Started watching etcd prefix %s for %s config changes", prefix, scopeName)

	for {
		select {
		case <-l.ctx.Done():
			logrus.Infof("Config watcher for %s stopped (context cancelled)", scopeName)
			return

		case watchResp, ok := <-watchChan:
			if !ok {
				logrus.Warnf("etcd %s watch channel closed, restarting...", scopeName)
				time.Sleep(1 * time.Second)
				watchChan = client.EtcdWatch(l.ctx, prefix, true)
				continue
			}
			if watchResp.Canceled {
				logrus.Warnf("etcd %s watch was canceled, restarting...", scopeName)
				time.Sleep(1 * time.Second)
				watchChan = client.EtcdWatch(l.ctx, prefix, true)
				continue
			}
			if err := watchResp.Err(); err != nil {
				logrus.Errorf("etcd %s watch error: %v", scopeName, err)
				time.Sleep(1 * time.Second)
				watchChan = client.EtcdWatch(l.ctx, prefix, true)
				continue
			}
			for _, event := range watchResp.Events {
				l.handleEtcdEvent(event, prefix)
			}
		}
	}
}

// handleEtcdEvent handles a single etcd event from a given prefix
func (l *configUpdateListener) handleEtcdEvent(event *clientv3.Event, prefix string) {
	key := string(event.Kv.Key)
	newValue := string(event.Kv.Value)

	// Extract config key (remove prefix)
	if len(key) <= len(prefix) {
		logrus.Warnf("Invalid etcd key: %s", key)
		return
	}
	configKey := key[len(prefix):]

	var oldValue string
	if event.PrevKv != nil {
		oldValue = string(event.PrevKv.Value)
	}

	logrus.WithFields(logrus.Fields{
		"type":      event.Type,
		"key":       configKey,
		"old_value": oldValue,
		"new_value": newValue,
	}).Info("received config change from etcd")

	// Apply config change via registry
	if err := handleConfigChange(l.ctx, configKey, oldValue, newValue); err != nil {
		logrus.Errorf("failed to apply config update for %s: %v", configKey, err)
		return
	}

	logrus.Infof("successfully applied config change for %s", configKey)
}
