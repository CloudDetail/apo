package database

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) initRole() error {
	roles := []Role{
		{RoleName: model.ROLE_ADMIN},
		{RoleName: model.ROLE_MANAGER},
		{RoleName: model.ROLE_VIEWER},
		{RoleName: model.ROLE_ANONYMOS},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Role{}); err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "role_name"}},
			DoNothing: true,
		}).Create(roles).Error
	})
}
