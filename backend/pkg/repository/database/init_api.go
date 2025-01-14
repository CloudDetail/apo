// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"github.com/spf13/viper"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) initApi() error {
	var apis []API
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./sqlscripts/api.yml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("apis", &apis); err != nil {
		return err
	}

	return repo.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}, {Name: "method"}},
		UpdateAll: true,
	}).Create(&apis).Error
}
