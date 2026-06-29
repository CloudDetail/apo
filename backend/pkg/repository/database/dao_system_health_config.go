// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

const DefaultSystemHealthSevereAbnormalScoreThreshold = 60
const DefaultSystemHealthGeneralAbnormalScoreThreshold = 80

type SystemHealthConfig struct {
	ID                            uint      `gorm:"primaryKey;autoIncrement"`
	AbnormalScoreThreshold        int       `gorm:"column:abnormal_score_threshold;not null;default:80"`
	SevereAbnormalScoreThreshold  int       `gorm:"column:severe_abnormal_score_threshold;not null;default:60"`
	GeneralAbnormalScoreThreshold int       `gorm:"column:general_abnormal_score_threshold;not null;default:80"`
	UpdatedAt                     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SystemHealthConfig) TableName() string {
	return "system_health_config"
}

func (repo *daoRepo) GetOrCreateSystemHealthConfig(ctx core.Context) (SystemHealthConfig, error) {
	var cfg SystemHealthConfig
	result := repo.GetContextDB(ctx).Order("id ASC").First(&cfg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		cfg = SystemHealthConfig{
			AbnormalScoreThreshold:        DefaultSystemHealthGeneralAbnormalScoreThreshold,
			SevereAbnormalScoreThreshold:  DefaultSystemHealthSevereAbnormalScoreThreshold,
			GeneralAbnormalScoreThreshold: DefaultSystemHealthGeneralAbnormalScoreThreshold,
		}
		return cfg, repo.GetContextDB(ctx).Create(&cfg).Error
	}
	if result.Error != nil {
		return SystemHealthConfig{}, result.Error
	}
	cfg = normalizeSystemHealthConfigModel(cfg)
	return cfg, nil
}

func (repo *daoRepo) UpdateSystemHealthConfig(ctx core.Context, severeAbnormalScoreThreshold int, generalAbnormalScoreThreshold int) (SystemHealthConfig, error) {
	cfg, err := repo.GetOrCreateSystemHealthConfig(ctx)
	if err != nil {
		return SystemHealthConfig{}, err
	}
	cfg.AbnormalScoreThreshold = generalAbnormalScoreThreshold
	cfg.SevereAbnormalScoreThreshold = severeAbnormalScoreThreshold
	cfg.GeneralAbnormalScoreThreshold = generalAbnormalScoreThreshold
	err = repo.GetContextDB(ctx).Model(&cfg).Updates(SystemHealthConfig{
		AbnormalScoreThreshold:        generalAbnormalScoreThreshold,
		SevereAbnormalScoreThreshold:  severeAbnormalScoreThreshold,
		GeneralAbnormalScoreThreshold: generalAbnormalScoreThreshold,
	}).Error
	if err != nil {
		return SystemHealthConfig{}, err
	}
	return cfg, nil
}

func normalizeSystemHealthConfigModel(cfg SystemHealthConfig) SystemHealthConfig {
	generalThreshold := cfg.GeneralAbnormalScoreThreshold
	if generalThreshold < 1 || generalThreshold > 100 {
		generalThreshold = cfg.AbnormalScoreThreshold
	}
	if generalThreshold < 1 || generalThreshold > 100 {
		generalThreshold = DefaultSystemHealthGeneralAbnormalScoreThreshold
	}

	severeThreshold := cfg.SevereAbnormalScoreThreshold
	if severeThreshold < 1 || severeThreshold >= generalThreshold {
		severeThreshold = DefaultSystemHealthSevereAbnormalScoreThreshold
	}
	if severeThreshold >= generalThreshold {
		severeThreshold = generalThreshold - 1
	}
	if severeThreshold < 1 {
		severeThreshold = 1
	}

	cfg.AbnormalScoreThreshold = generalThreshold
	cfg.SevereAbnormalScoreThreshold = severeThreshold
	cfg.GeneralAbnormalScoreThreshold = generalThreshold
	return cfg
}
