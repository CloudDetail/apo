// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"

func (repo *subRepo) GetAlertSlience() ([]sc.AlertSlienceConfig, error) {
	var result []sc.AlertSlienceConfig
	err := repo.db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *subRepo) AddAlertSlience(SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.db.Create(SlienceConfig).Error
}

func (repo *subRepo) UpdateAlertSlience(SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.db.Where("id = ?", SlienceConfig.ID).Updates(SlienceConfig).Error
}

func (repo *subRepo) DeleteAlertSlience(id int) error {
	return repo.db.Delete(&sc.AlertSlienceConfig{}, "id = ? ", id).Error
}
