// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import "os"

// AlertMetricsData provide PQL corresponding to user selectable metrics
type AlertMetricsData struct {
	Id int `json:"-" gorm:"primaryKey;autoIncrement"`

	Name  string `json:"name" gorm:"not null;type:varchar(100);column:name"`
	PQL   string `json:"pql" gorm:"not null;type:varchar(5000);column:pql"`
	Unit  string `json:"unit" gorm:"not null;type:varchar(100);column:unit"`
	Group string `json:"group" gorm:"not null;type:varchar(100);column:group"`
}

func (a *AlertMetricsData) TableName() string {
	return "quick_alert_rule_metric"
}

// ListQuickMutationMetric list of all quick metrics
func (repo *daoRepo) ListQuickAlertRuleMetric() ([]AlertMetricsData, error) {
	var quickAlertMetrics []AlertMetricsData
	err := repo.db.Find(&quickAlertMetrics).Error
	return quickAlertMetrics, err
}

func (repo *daoRepo) InitPredefinedQuickAlertRuleMetric(sqlScript string) error {
	if err := repo.db.AutoMigrate(&AlertMetricsData{}); err != nil {
		return err
	}
	var count int64
	if err := repo.db.Model(&AlertMetricsData{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if sqlScript == "" {
		sqlScript = "./sqlscripts/default_quick_alert_rule_metric.sql"
	}

	if _, err := os.Stat(sqlScript); err == nil {
		// Read the file and execute the initialization script
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
