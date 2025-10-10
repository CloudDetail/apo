// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

func (repo *appRepo) CreateLogo(logo *model.Logo) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id != 0").Delete(&model.Logo{}).Error; err != nil {
			return err
		}

		return tx.Create(logo).Error
	})
}

func (repo *appRepo) GetLogo() (*model.Logo, error) {
	logo := &model.Logo{}
	err := repo.db.Model(&model.Logo{}).Find(logo).Error
	if err != nil {
		return nil, err
	}

	return logo, nil
}
