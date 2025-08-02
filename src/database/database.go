package database

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
		&Label{},

		&Container{},
		&ContainerLabel{},
		&Dataset{},
		&DatasetLabel{},
		&Task{},
		&ExecutionResult{},
		&ExecutionResultLabel{},
		&GranularityResult{},
		&Detector{},
		&FaultInjectionSchedule{},
		&DatasetFaultInjection{},
		&FaultInjectionLabel{},

		&Project{},
		&ProjectContanier{},
		&ProjectDataset{},
		&ProjectLabel{},
		&User{},
		&Role{},
		&Permission{},
		&Resource{},
		&UserProject{},
		&UserRole{},
		&RolePermission{},
		&UserPermission{},
	); err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}

	createExecutionResultViews()
	createFaultInjectionViews()
	createDetectorViews()
}

func connectWithRetry(dbConfig *databaseConfig) {
	maxRetries := 3
	retryDelay := 10 * time.Second

	var err error
	for i := 0; i <= maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dbConfig.ToDSN()), &gorm.Config{})
		if err == nil {
			logrus.Info("Successfully connected to the database.")
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
