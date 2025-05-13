// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (repo *subRepo) GetAlertSlience(ctx_core core.Context,) ([]sc.AlertSlienceConfig, error) {
	var result []sc.AlertSlienceConfig
	err := repo.db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *subRepo) AddAlertSlience(ctx_core core.Context, SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.db.Create(SlienceConfig).Error
}

func (repo *subRepo) UpdateAlertSlience(ctx_core core.Context, SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.db.Where("id = ?", SlienceConfig.ID).Updates(SlienceConfig).Error
}

func (repo *subRepo) DeleteAlertSlience(ctx_core core.Context, id int) error {
	return repo.db.Delete(&sc.AlertSlienceConfig{}, "id = ? ", id).Error
}
