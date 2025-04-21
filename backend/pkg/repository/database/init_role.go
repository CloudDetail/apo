// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) initRole() error {
	roles := []Role{
		{RoleName: model.ROLE_ADMIN},
		{RoleName: model.ROLE_VIEWER},
		{RoleName: model.ROLE_ANONYMOS},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "role_name"}},
			DoNothing: true,
		}).Create(roles).Error
	})
}
