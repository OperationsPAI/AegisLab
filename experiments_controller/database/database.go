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
	); err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}

	createFaultInjectionIndexes()
	verifyAllIndexes()

	createFaultInjectionViews()
}

func createFaultInjectionIndexes() {
	indexQueries := []string{
		// 1. Create GIN index for JSONB field (supports @> and ->> operations)
		`CREATE INDEX IF NOT EXISTS idx_fault_injection_schedules_labels_gin 
         ON fault_injection_schedules USING GIN (labels)`,

		// 2. Create expression indexes for common JSON keys
		`CREATE INDEX IF NOT EXISTS idx_fault_injection_schedules_env 
         ON fault_injection_schedules ((labels ->> 'env'))`,

		`CREATE INDEX IF NOT EXISTS idx_fault_injection_schedules_batch 
         ON fault_injection_schedules ((labels ->> 'batch'))`,

		// 3. Composite expression index for multi-condition JSONB queries
		`CREATE INDEX IF NOT EXISTS idx_fault_injection_schedules_env_batch 
         ON fault_injection_schedules ((labels ->> 'env'), (labels ->> 'batch'))`,

		// 4. Composite index with time field (optimized for ListInjections queries)
		`CREATE INDEX IF NOT EXISTS idx_fault_injection_schedules_query_optimized 
         ON fault_injection_schedules ((labels ->> 'env'), (labels ->> 'batch'), benchmark, status, fault_type, created_at DESC)`,

		// 5. Dedicated time index for time range queries
		`CREATE INDEX IF NOT EXISTS idx_fault_injection_schedules_created_at 
         ON fault_injection_schedules (created_at DESC)`,
	}

	for _, query := range indexQueries {
		if err := DB.Exec(query).Error; err != nil {
			logrus.Warnf("failed to create index: %v", err)
		}
	}
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
