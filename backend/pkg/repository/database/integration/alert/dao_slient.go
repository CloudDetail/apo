// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
)

func (repo *subRepo) GetAlertSlience(ctx core.Context) ([]sc.AlertSlienceConfig, error) {
	var result []sc.AlertSlienceConfig
	err := repo.GetContextDB(ctx).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *subRepo) AddAlertSlience(ctx core.Context, SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.GetContextDB(ctx).Create(SlienceConfig).Error
}

func (repo *subRepo) UpdateAlertSlience(ctx core.Context, SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.GetContextDB(ctx).Where("id = ?", SlienceConfig.ID).Updates(SlienceConfig).Error
}

func (repo *subRepo) DeleteAlertSlience(ctx core.Context, id int) error {
	return repo.GetContextDB(ctx).Delete(&sc.AlertSlienceConfig{}, "id = ? ", id).Error
}
