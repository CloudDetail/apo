// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"gorm.io/gorm"
)

const defaultCustomDashboardSQLDirectory = "./sqlscripts/default_custom_dashboards"

func (repo *daoRepo) initDefaultCustomDashboards(ctx core.Context) error {
	db := repo.GetContextDB(ctx)
	path, err := defaultCustomDashboardSQLPath(defaultCustomDashboardSQLDirectory, db.Dialector.Name())
	if err != nil {
		return err
	}
	return executeDefaultCustomDashboardSQL(db, path)
}

func defaultCustomDashboardSQLPath(directory, dialect string) (string, error) {
	switch dialect {
	case "sqlite", "mysql", "postgres":
		return filepath.Join(directory, fmt.Sprintf("default_custom_dashboards-%s.sql", dialect)), nil
	default:
		return "", fmt.Errorf("unsupported database dialect for default custom dashboards: %s", dialect)
	}
}

func executeDefaultCustomDashboardSQL(db *gorm.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read default custom dashboard SQL %s: %w", path, err)
	}
	sql, err := util.ValidateSQL(string(raw))
	if err != nil {
		return fmt.Errorf("validate default custom dashboard SQL %s: %w", path, err)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sql).Error; err != nil {
			return fmt.Errorf("execute default custom dashboard SQL %s: %w", path, err)
		}
		return nil
	})
}
