// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"gorm.io/gorm"
)

func (repo *daoRepo) CreateDingTalkReceiver(ctx core.Context, dingTalkConfig *amconfig.DingTalkConfig) error {
	return repo.GetContextDB(ctx).Create(dingTalkConfig).Error
}

func (repo *daoRepo) GetDingTalkReceiver(ctx core.Context, uuid string) (amconfig.DingTalkConfig, error) {
	config := amconfig.DingTalkConfig{}
	err := repo.GetContextDB(ctx).Select("url, secret").Where("uuid = ?", uuid).First(&config).Error
	return config, err
}

func (repo *daoRepo) GetDingTalkReceiverByAlertName(ctx core.Context, configFile string, alertName string, page, pageSize int) ([]*amconfig.DingTalkConfig, int64, error) {
	var dingTalkConfigs []*amconfig.DingTalkConfig
	offset := (page - 1) * pageSize

	query := repo.GetContextDB(ctx).Select("alert_name, url, secret").Where("config_file = ?", configFile)
	countQuery := repo.GetContextDB(ctx).Model(&amconfig.DingTalkConfig{}).Select("*").Where("config_file = ?", configFile)

	if len(alertName) > 0 {
		query = query.Where("alert_name = ?", alertName)
		countQuery = countQuery.Where("alert_name = ?", alertName)
	}

	err := query.Offset(offset).Limit(pageSize).Find(&dingTalkConfigs).Error
	var count int64
	countQuery.Count(&count)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, count, err
	}

	return dingTalkConfigs, count, nil
}

func (repo *daoRepo) UpdateDingTalkReceiver(ctx core.Context, dingTalkConfig *amconfig.DingTalkConfig, oldName string) error {
	return repo.GetContextDB(ctx).Where("config_file = ? AND alert_name = ?", dingTalkConfig.ConfigFile, oldName).Updates(dingTalkConfig).Error
}

func (repo *daoRepo) DeleteDingTalkReceiver(ctx core.Context, configFile, alertName string) error {
	return repo.GetContextDB(ctx).Where("config_file = ? AND alert_name = ?", configFile, alertName).Delete(&amconfig.DingTalkConfig{}).Error
}
