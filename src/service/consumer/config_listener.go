package consumer

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

// configUpdateListener listens for configuration update events from etcd
type configUpdateListener struct {
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
	etcdPrefix string
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
			ctx:        listenerCtx,
			cancel:     cancel,
			etcdPrefix: consts.ConfigEtcdPrefix,
		}

		go func() {
			<-ctx.Done()
			logrus.Info("Parent context cancelled, stopping config update listener...")
			configListenerInstance.Stop()
		}()
	})
	return configListenerInstance
}

// Start starts listening for configuration update events from etcd
func (l *configUpdateListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Load initial configs from etcd (source of truth) on first start
	// Falls back to MySQL defaults if etcd values don't exist
	if err := l.loadConfigsFromEtcd(); err != nil {
		logrus.Errorf("Failed to load initial configs from etcd: %v", err)
	}

	l.isRunning = true
	go l.watchEtcd()
	logrus.Info("Config update listener started, watching etcd for changes")
	return nil
}

// Stop stops the listener
func (l *configUpdateListener) Stop() {
	if !l.isRunning {
		return
	}

	l.cancel()
	l.isRunning = false
	logrus.Info("Config update listener stopped")
}

// loadConfigsFromEtcd loads all consumer configs from etcd (the source of truth) on startup.
//
// Falls back to MySQL defaults only if config doesn't exist in etcd
func (l *configUpdateListener) loadConfigsFromEtcd() error {
	configMetadata, err := repository.ListConfigByScope(database.DB, consts.ConfigScopeConsumer)
	if err != nil {
		return fmt.Errorf("failed to list consumer config metadata from database: %w", err)
	}

	loadedCount := 0
	initializedCount := 0

	for _, meta := range configMetadata {
		etcdKey := fmt.Sprintf("%s%s", l.etcdPrefix, meta.Key)

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

	logrus.Infof("Loaded %d/%d consumer configs from etcd to Viper (initialized %d new configs)",
		loadedCount, len(configMetadata), initializedCount)

	return nil
}

// watchEtcd watches etcd for configuration changes
func (l *configUpdateListener) watchEtcd() {
	watchChan := client.EtcdWatch(l.ctx, l.etcdPrefix, true)
	logrus.WithField("prefix", l.etcdPrefix).Info("Started watching etcd for config changes")

	for {
		select {
		case <-l.ctx.Done():
			logrus.Info("Config watcher context cancelled")
			return

		case watchResp, ok := <-watchChan:
			if !ok {
				logrus.Warn("etcd watch channel closed, stopping watcher")
				return
			}

			if watchResp.Canceled {
				logrus.Warn("etcd watch was canceled, restarting...")
				time.Sleep(1 * time.Second)
				watchChan = client.EtcdWatch(l.ctx, l.etcdPrefix, true)
				continue
			}

			if err := watchResp.Err(); err != nil {
				logrus.Errorf("etcd watch error: %v", err)
				time.Sleep(1 * time.Second)
				watchChan = client.EtcdWatch(l.ctx, l.etcdPrefix, true)
				continue
			}

			for _, event := range watchResp.Events {
				l.handleEtcdEvent(event)
			}
		}
	}
}

// handleEtcdEvent handles a single etcd event
func (l *configUpdateListener) handleEtcdEvent(event *clientv3.Event) {
	key := string(event.Kv.Key)
	newValue := string(event.Kv.Value)

	// Extract config key (remove prefix)
	if len(key) <= len(l.etcdPrefix) {
		logrus.Warnf("Invalid etcd key: %s", key)
		return
	}
	configKey := key[len(l.etcdPrefix):]

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
	if err := HandleConfigChange(l.ctx, configKey, oldValue, newValue); err != nil {
		logrus.Errorf("failed to apply config update for %s: %v", configKey, err)
		return
	}

	logrus.Infof("successfully applied config change for %s", configKey)
}
