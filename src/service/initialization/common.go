package initialization

import (
	"aegis/config"
	"aegis/consts"
	"aegis/service/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func activateConfigScope(scope consts.ConfigScope, listener *common.ConfigUpdateListener) error {
	if err := listener.EnsureScope(consts.ConfigScopeGlobal); err != nil {
		return fmt.Errorf("failed to activate global config listener: %w", err)
	}

	if scope == consts.ConfigScopeConsumer {
		if err := listener.EnsureScope(consts.ConfigScopeConsumer); err != nil {
			return fmt.Errorf("failed to activate consumer config listener: %w", err)
		}
	}

	logrus.Infof("Config handlers registered for scope %s, %d total handler(s)",
		consts.GetConfigScopeName(scope), len(common.ListRegisteredConfigKeys(nil)))

	config.SetDetectorName(config.GetString(consts.DetectorKey))
	logrus.Infof("Global detector name initialized: %s", config.GetDetectorName())
	return nil
}
