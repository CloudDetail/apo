// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import

// API represents an API endpoint definition
core "github.com/CloudDetail/apo/backend/pkg/core"

type API struct {
	ID      int    `gorm:"primary_key;auto_increment"`
	Path    string `gorm:"column:path;type:varchar(255);index:idx_path_method,unique" mapstructure:"path"`
	Method  string `gorm:"column:method;type:varchar(10);index:idx_path_method,unique" mapstructure:"method"`
	Enabled bool   `gorm:"column:enabled;default:true"`
}

func (API) TableName() string {
	return "api"
}

func (repo *daoRepo) GetAPIByPath(ctx core.Context, path string, method string) (*API, error) {
	var api API
	err := repo.GetContextDB(ctx).Where("path = ? AND method = ?", path, method).Find(&api).Error
	return &api, err
}
