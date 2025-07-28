package database

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// FixExecutionResultForeignKey 修复 execution_results 表的外键约束
func FixExecutionResultForeignKey() error {
	logrus.Info("开始修复 execution_results 表的外键约束...")

	// 1. 删除旧的外键约束（如果存在）
	if err := dropOldForeignKey(); err != nil {
		logrus.Warnf("删除旧外键约束时出现警告: %v", err)
	}

	// 2. 添加新的外键约束
	if err := addNewForeignKey(); err != nil {
		return fmt.Errorf("添加新外键约束失败: %v", err)
	}

	logrus.Info("execution_results 表的外键约束修复完成")
	return nil
}

// dropOldForeignKey 删除旧的外键约束
func dropOldForeignKey() error {
	// 检查是否存在旧的外键约束
	var constraintName string
	err := DB.Raw(`
		SELECT conname 
		FROM pg_constraint 
		WHERE conrelid = 'execution_results'::regclass 
		AND contype = 'f' 
		AND conname LIKE '%dataset%'
	`).Scan(&constraintName).Error

	if err != nil {
		return fmt.Errorf("查询外键约束失败: %v", err)
	}

	if constraintName != "" {
		// 删除旧的外键约束
		err = DB.Exec(fmt.Sprintf("ALTER TABLE execution_results DROP CONSTRAINT IF EXISTS %s", constraintName)).Error
		if err != nil {
			return fmt.Errorf("删除外键约束失败: %v", err)
		}
		logrus.Infof("删除了旧的外键约束: %s", constraintName)
	}

	return nil
}

// addNewForeignKey 添加新的外键约束
func addNewForeignKey() error {
	// 添加新的外键约束，指向 fault_injection_schedules 表
	err := DB.Exec(`
		ALTER TABLE execution_results 
		ADD CONSTRAINT fk_execution_results_dataset 
		FOREIGN KEY (dataset_id) 
		REFERENCES fault_injection_schedules(id) 
		ON DELETE CASCADE
	`).Error

	if err != nil {
		return fmt.Errorf("添加外键约束失败: %v", err)
	}

	logrus.Info("成功添加了新的外键约束: fk_execution_results_dataset")
	return nil
}
