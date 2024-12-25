// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"os"
)

// createMenuItems Auto migrate table and execute init sql script.
// It will not execute if target table has record.
// Make sure sql script exists.
func (repo *daoRepo) initSql(model interface{}, sqlScript string) error {
	if err := repo.db.AutoMigrate(&model); err != nil {
		return err
	}

	var count int64
	repo.db.Model(&model).Count(&count)
	if count > 0 {
		return nil
	}

	if _, err := os.Stat(sqlScript); err == nil {
		sql, err := os.ReadFile(sqlScript)
		if err != nil {
			return err
		}
		if err := repo.db.Exec(string(sql)).Error; err != nil {
			return err
		}
	}
	return nil
}
