package database

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// 全局 DB 对象
var DB *gorm.DB

func InitDB() {
	var err error
	pgUser := config.GetString("database.postgres_user")
	pgPassword := config.GetString("database.postgres_password")
	pgHost := config.GetString("database.postgres_host")
	pgPort := config.GetString("database.postgres_port")
	pgDBName := config.GetString("database.postgres_db")

	pgDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", pgHost, pgUser, pgPassword, pgDBName, pgPort)

	maxRetries := 3
	retryDelay := 10 * time.Second

	for i := 0; i <= maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(pgDSN), &gorm.Config{})
		if err == nil {
			logrus.Info("Successfully connected to the database.")
			if err := DB.Use(tracing.NewPlugin()); err != nil {
				panic(err)
			}
			break // Connection successful, exit loop
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

	if err = DB.AutoMigrate(
		&Container{},
		&Project{},
		&Task{},
		&FaultInjectionSchedule{},
		&ExecutionResult{},
		&GranularityResult{},
		&Detector{},
		&Dataset{},
		&Label{},
		&DatasetFaultInjection{},
		&DatasetLabel{},
		&FaultInjectionLabel{},
		&ContainerLabel{},
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
