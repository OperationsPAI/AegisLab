package initialization

import (
	"aegis/config"
	"aegis/consts"
	"aegis/service/common"
	"context"

	"github.com/sirupsen/logrus"
)

func registerHandlers(ctx context.Context, scope consts.ConfigScope, handlerFunc func()) {
	// Register global-scope handlers (idempotent via sync.Once)
	common.RegisterGlobalHandlers()
	if handlerFunc != nil {
		handlerFunc()
	}

	// Ensure etcd listener covers required scopes (each call is idempotent)
	listener := common.GetConfigUpdateListener(ctx)

	if err := listener.EnsureScope(consts.ConfigScopeGlobal); err != nil {
		logrus.Fatalf("Failed to activate global config listener: %v", err)
	}

	if scope == consts.ConfigScopeConsumer {
		if err := listener.EnsureScope(consts.ConfigScopeConsumer); err != nil {
			logrus.Fatalf("Failed to activate consumer config listener: %v", err)
		}
	}

	logrus.Infof("Config handlers registered for scope %s, %d total handler(s)",
		consts.GetConfigScopeName(scope), len(common.ListRegisteredConfigKeys(nil)))

	// Sync atomic vars from viper (listener has loaded configs from etcd)
	config.SetDetectorName(config.GetString(consts.DetectorKey))
	logrus.Infof("Global detector name initialized: %s", config.GetDetectorName())
}
