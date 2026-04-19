package common

import (
	"fmt"

	"aegis/consts"
	"aegis/model"

	"gorm.io/gorm"
)

type configStore struct {
	db *gorm.DB
}

func newConfigStore(db *gorm.DB) *configStore {
	return &configStore{db: db}
}

func (s *configStore) getConfigByKey(key string) (*model.DynamicConfig, error) {
	var cfg model.DynamicConfig
	if err := s.db.Where("config_key = ?", key).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("failed to find config with key %s: %w", key, err)
	}
	return &cfg, nil
}

func (s *configStore) listConfigsByScope(scope consts.ConfigScope) ([]model.DynamicConfig, error) {
	var configs []model.DynamicConfig
	if err := s.db.Where("scope = ?", scope).Order("config_key ASC").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list configs by scope %s: %w", consts.GetConfigScopeName(scope), err)
	}
	return configs, nil
}
