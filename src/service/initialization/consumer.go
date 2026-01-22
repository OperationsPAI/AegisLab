package initialization

import (
	"context"
	"fmt"
	"path/filepath"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/repository"
	"aegis/service/common"
	"aegis/service/consumer"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var consumerData *ConsumerData

func InitConcurrencyLock(ctx context.Context) {
	if err := repository.InitConcurrencyLock(ctx); err != nil {
		logrus.Fatalf("error setting concurrency lock to 0: %v", err)
	}
}

func InitializeConsumer(ctx context.Context) {
	configs, err := repository.ListExistingConfigs(database.DB)
	if err != nil {
		logrus.Fatalf("Failed to check existing dynamic configs: %v", err)
	}

	consumerData = &ConsumerData{
		configs: configs,
	}

	if len(consumerData.configs) == 0 {
		logrus.Info("Seeding initial system data for consumer...")
		if err := initializeConsumer(); err != nil {
			logrus.Fatalf("Failed to initialize system data for consumer: %v", err)
		}
		logrus.Info("Successfully seeded initial system data for consumer")
	} else {
		logrus.Info("Initial system data for consumer already seeded, skipping initialization")
	}

	// Register built-in config handlers
	consumer.RegisterBuiltinHandlers()

	// Start config update listener
	listener := consumer.GetConfigUpdateListener(ctx)

	if err := listener.Start(); err != nil {
		logrus.Fatalf("Failed to start config update listener: %v", err)
	}
	logrus.Infof("Config update listener started successfully, watching %d registered handlers",
		len(consumer.ListRegisteredConfigKeys()))

	// Initialize namespaces on startup - critical after restart to re-initialize CRD informers
	logrus.Info("Initializing namespaces on startup...")
	monitor := consumer.GetMonitor()
	if monitor == nil {
		logrus.Warn("Monitor not initialized, skipping namespace initialization")
	} else {
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
		if err := consumer.UpdateK8sController(k8s.GetK8sController(), initialized, []string{}); err != nil {
			logrus.Errorf("Failed to update k8s controller: %v", err)
			return
		}
	}
}

func initializeConsumer() error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	return withOptimizedDBSettings(func() error {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := initializeDynamicConfigs(tx, initialData); err != nil {
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

func initializeDynamicConfigs(tx *gorm.DB, data *InitialData) error {
	var configs []database.DynamicConfig
	for _, configData := range data.DynamicConfigs {
		cfg := configData.ConvertToDBDynamicConfig()
		if err := common.ValidateConfigMetadataConstraints(cfg); err != nil {
			return fmt.Errorf("invalid config value for key %s: %w", configData.Key, err)
		}

		if err := common.CreateConfig(tx, cfg); err != nil {
			return fmt.Errorf("failed to create dynamic config %s: %w", configData.Key, err)
		}

		configs = append(configs, *cfg)
	}

	consumerData.configs = configs
	return nil
}
