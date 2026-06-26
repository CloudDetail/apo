// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

const DefaultSystemHealthAbnormalScoreThreshold = 80

type SystemHealthConfig struct {
	ID                     uint      `gorm:"primaryKey;autoIncrement"`
	AbnormalScoreThreshold int       `gorm:"column:abnormal_score_threshold;not null;default:80"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SystemHealthConfig) TableName() string {
	return "system_health_config"
}

func (repo *daoRepo) GetOrCreateSystemHealthConfig(ctx core.Context) (SystemHealthConfig, error) {
	var cfg SystemHealthConfig
	result := repo.GetContextDB(ctx).Order("id ASC").First(&cfg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		cfg = SystemHealthConfig{AbnormalScoreThreshold: DefaultSystemHealthAbnormalScoreThreshold}
		return cfg, repo.GetContextDB(ctx).Create(&cfg).Error
	}
	if result.Error != nil {
		return SystemHealthConfig{}, result.Error
	}
	if cfg.AbnormalScoreThreshold < 1 || cfg.AbnormalScoreThreshold > 100 {
		cfg.AbnormalScoreThreshold = DefaultSystemHealthAbnormalScoreThreshold
	}
	return cfg, nil
}

func (repo *daoRepo) UpdateSystemHealthConfig(ctx core.Context, abnormalScoreThreshold int) (SystemHealthConfig, error) {
	cfg, err := repo.GetOrCreateSystemHealthConfig(ctx)
	if err != nil {
		return SystemHealthConfig{}, err
	}
	cfg.AbnormalScoreThreshold = abnormalScoreThreshold
	err = repo.GetContextDB(ctx).Model(&cfg).Updates(SystemHealthConfig{
		AbnormalScoreThreshold: abnormalScoreThreshold,
	}).Error
	if err != nil {
		return SystemHealthConfig{}, err
	}
	return cfg, nil
}
