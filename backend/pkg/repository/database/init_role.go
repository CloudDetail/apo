// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

func (repo *daoRepo) initRole() error {
	roles := []string{model.ROLE_ADMIN, model.ROLE_MANAGER, model.ROLE_VIEWER, model.ROLE_ANONYMOS}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Role{}); err != nil {
			return err
		}
		for _, roleName := range roles {
			var count int64
			if err := tx.Model(&Role{}).Where("role_name = ?", roleName).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				role := Role{RoleName: roleName}
				if err := tx.Create(&role).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
