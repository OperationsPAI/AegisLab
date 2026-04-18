package initialization

import (
	"context"
	"fmt"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	k8sinfra "aegis/infra/k8s"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	"aegis/service/common"
	"aegis/service/consumer"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeConsumer(
	ctx context.Context,
	db *gorm.DB,
	controller *k8sinfra.Controller,
	monitor consumer.NamespaceMonitor,
	publisher *redisinfra.Gateway,
	listener *common.ConfigUpdateListener,
	restartLimiter *consumer.TokenBucketRateLimiter,
	buildLimiter *consumer.TokenBucketRateLimiter,
	algoLimiter *consumer.TokenBucketRateLimiter,
) error {
	consumerData, err := newConfigDataWithDB(db, consts.ConfigScopeConsumer)
	if err != nil {
		return fmt.Errorf("failed to load consumer config metadata: %w", err)
	}

	if len(consumerData.configs) == 0 {
		logrus.Info("Seeding initial system data for consumer...")
		if err := initializeConsumer(db); err != nil {
			return fmt.Errorf("failed to initialize system data for consumer: %w", err)
		}
		logrus.Info("Successfully seeded initial system data for consumer")
	} else {
		logrus.Info("Initial system data for consumer already seeded, skipping initialization")
	}

	common.RegisterGlobalHandlers(publisher)
	consumer.RegisterConsumerHandlers(controller, monitor, publisher, restartLimiter, buildLimiter, algoLimiter)
	if err := activateConfigScope(consumerData.scope, listener); err != nil {
		return err
	}

	// Namespace/bootstrap informer initialization can take noticeably longer than
	// the Fx startup deadline when the local cluster is cold or slow. Run it in
	// the background so consumer/both startup does not fail with
	// "context deadline exceeded" during local debugging.
	if monitor == nil {
		logrus.Warn("Monitor not initialized, skipping namespace initialization")
		return nil
	}

	monitor.SetContext(ctx)
	go func() {
		logrus.Info("Initializing namespaces on startup...")

		initialized, err := monitor.InitializeNamespaces()
		if err != nil {
			logrus.Errorf("Failed to initialize namespaces: %v", err)
			return
		}

		if len(initialized) == 0 {
			logrus.Warn("No namespaces to initialize on startup")
			return
		}

		logrus.Infof("Initialized namespaces on startup: %v", initialized)
		if err := consumer.UpdateK8sController(controller, initialized, []string{}); err != nil {
			logrus.Errorf("Failed to update k8s controller: %v", err)
		}
	}()

	return nil
}

func initializeConsumer(db *gorm.DB) error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	return withOptimizedDBSettings(db, func() error {
		err := db.Transaction(func(tx *gorm.DB) error {
			if _, err := initializeDynamicConfigs(tx, initialData); err != nil {
				return fmt.Errorf("failed to initialize dynamic configs for consumer: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to initialize consumer data: %w", err)
		}

		return nil
	})
}

func initializeDynamicConfigs(tx *gorm.DB, data *InitialData) ([]model.DynamicConfig, error) {
	var configs []model.DynamicConfig
	for _, configData := range data.DynamicConfigs {
		cfg := configData.ConvertToDBDynamicConfig()
		if err := common.ValidateConfigMetadataConstraints(cfg); err != nil {
			return nil, fmt.Errorf("invalid config value for key %s: %w", configData.Key, err)
		}

		if err := common.CreateConfig(tx, cfg); err != nil {
			return nil, fmt.Errorf("failed to create dynamic config %s: %w", configData.Key, err)
		}

		configs = append(configs, *cfg)
	}

	return configs, nil
}
