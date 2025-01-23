// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"os"

	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

func (repo *subRepo) initDefaultAlertTagMapping(sqlScript string) error {
	if err := repo.db.AutoMigrate(
		&alert.AlertSource{},
		&alert.TargetTag{},
		&alert.AlertEnrichRule{},
		&alert.AlertEnrichCondition{},
		&alert.AlertEnrichSchemaTarget{},
		&alert.Cluster{},
		&alert.AlertSource2Cluster{},
	); err != nil {
		return err
	}

	var count int64
	if err := repo.db.Model(&alert.TargetTag{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if sqlScript == "" {
		sqlScript = "./sqlscripts/default_alert_tag_mapping.sql"
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
