package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"aegis/config"

	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

type databaseConfig struct {
	host     string
	port     int
	user     string
	password string
	database string
}

func (d *databaseConfig) ToDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.user, d.password, d.host, d.port, d.database)
}

// Global DB object
var DB *gorm.DB

func InitDB() {
	var err error

	mysqlConfig := &databaseConfig{
		host:     config.GetString("database.mysql_host"),
		port:     config.GetInt("database.mysql_port"),
		user:     config.GetString("database.mysql_user"),
		password: config.GetString("database.mysql_password"),
		database: config.GetString("database.mysql_db"),
	}

	connectWithRetry(mysqlConfig)

	if err = DB.AutoMigrate(
		// Core entities
		&Container{},
		&ContainerVersion{},
		&HelmConfig{},
		&ParameterConfig{},
		&Dataset{},
		&DatasetVersion{},
		&Project{},
		&Label{},
		&User{},
		&Role{},
		&Permission{},
		&Resource{},
		&AuditLog{},

		// Many-to-many relationship tables
		&ContainerLabel{},
		&DatasetLabel{},
		&ProjectLabel{},
		&ContainerVersionEnvVar{},
		&HelmConfigValue{},
		&DatasetVersionInjection{},
		&TaskLabel{},

		&UserContainer{},
		&UserDataset{},
		&UserProject{},
		&UserRole{},
		&RolePermission{},
		&UserPermission{},

		// Business entities
		&Task{},
		&FaultInjection{},
		&Execution{},
		&DetectorResult{},
		&GranularityResult{},
	); err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}

	createDetectorViews()
}

func connectWithRetry(dbConfig *databaseConfig) {
	maxRetries := 3
	retryDelay := 10 * time.Second

	var err error
	for i := 0; i <= maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dbConfig.ToDSN()), &gorm.Config{
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logger.Warn,
					IgnoreRecordNotFoundError: true,
					Colorful:                  true,
				}),
			TranslateError: true,
		})
		if err == nil {
			logrus.Info("Successfully connected to the database")
			if err := DB.Use(tracing.NewPlugin()); err != nil {
				panic(err)
			}

			break
		}

		logrus.Errorf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries+1, err)
		if i < maxRetries {
			logrus.Infof("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		logrus.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries+1, err)
	}
}
