// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"os"

	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

func (repo *subRepo) initDefaultExternalAlertTagMapping(sqlScript string) error {
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
		// 默认的初始化脚本
		sqlScript = "./sqlscripts/default_alert_tag_mapping.sql"
	}

	// 检查初始化脚本是否存在
	if _, err := os.Stat(sqlScript); err == nil {
		// 读取文件并执行初始化脚本
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
