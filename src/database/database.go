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

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Timezone string
}

func NewDatabaseConfig(databaseType string) *DatabaseConfig {
	return &DatabaseConfig{
		Type:     databaseType,
		Host:     config.GetString(fmt.Sprintf("database.%s.host", databaseType)),
		Port:     config.GetInt(fmt.Sprintf("database.%s.port", databaseType)),
		User:     config.GetString(fmt.Sprintf("database.%s.user", databaseType)),
		Password: config.GetString(fmt.Sprintf("database.%s.password", databaseType)),
		Database: config.GetString(fmt.Sprintf("database.%s.db", databaseType)),
		Timezone: config.GetString(fmt.Sprintf("database.%s.timezone", databaseType)),
	}
}

func (d *DatabaseConfig) ToDSN() (string, error) {
	if d.Type != "mysql" {
		return "", fmt.Errorf("unsupported database type: %s", d.Type)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Database)
	return dsn, nil
}

// Global DB object
var DB *gorm.DB

func InitDB() {
	var err error

	mysqlConfig := NewDatabaseConfig("mysql")

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

		// Business entities
		&Task{},
		&FaultInjection{},
		&Execution{},
		&DetectorResult{},
		&GranularityResult{},

		// Many-to-many relationship tables
		&ContainerLabel{},
		&DatasetLabel{},
		&ProjectLabel{},
		&ContainerVersionEnvVar{},
		&HelmConfigValue{},
		&DatasetVersionInjection{},
		&FaultInjectionLabel{},
		&ExecutionInjectionLabel{},
		&ConfigLabel{},

		&UserContainer{},
		&UserDataset{},
		&UserProject{},
		&UserRole{},
		&RolePermission{},
		&UserPermission{},

		// Dynamic configuration entities
		&DynamicConfig{},
		&ConfigHistory{},
	); err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}

	createDetectorViews()
}

func connectWithRetry(dbConfig *DatabaseConfig) {
	maxRetries := 3
	retryDelay := 10 * time.Second

	dsn, err := dbConfig.ToDSN()
	if err != nil {
		logrus.Fatalf("Failed to construct DSN: %v", err)
	}

	for i := 0; i <= maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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
