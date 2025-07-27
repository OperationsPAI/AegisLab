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

	createFaultInjectionIndexes()
	verifyAllIndexes()

	createExecutionResultViews()
	createFaultInjectionViews()
	createDetectorViews()
}

func createFaultInjectionIndexes() {
	// 注意：原有的 JSONB labels 索引已移除，因为现在使用统一的 Label 表和关系表
	// 如果需要为新的标签系统创建特殊索引，可以在这里添加
}

func verifyAllIndexes() {
	tables := []string{
		"fault_injection_schedules",
	}

	for _, table := range tables {
		verifyIndexes(table)
	}
}

func verifyIndexes(tableName string) {
	var indexes []struct {
		IndexName string `gorm:"column:indexname"`
		TableName string `gorm:"column:tablename"`
	}

	query := `
        SELECT indexname, tablename 
        FROM pg_indexes 
        WHERE tablename = ? 
        ORDER BY indexname
    `

	if err := DB.Raw(query, tableName).Scan(&indexes).Error; err != nil {
		logrus.Errorf("Failed to verify indexes: %v", err)
		return
	}

	logrus.Infof("Found %d indexes on fault_injection_schedules:", len(indexes))
	for _, idx := range indexes {
		logrus.Infof("  - %s", idx.IndexName)
	}
}
