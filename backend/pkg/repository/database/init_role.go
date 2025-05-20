// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var validRoles = []Role{
	{RoleName: model.ROLE_ADMIN},
	{RoleName: model.ROLE_VIEWER},
	{RoleName: model.ROLE_ANONYMOS},
}

func (repo *daoRepo) initRole(ctx core.Context) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "role_name"}},
			DoNothing: true,
		}).Create(validRoles).Error
	})
}
