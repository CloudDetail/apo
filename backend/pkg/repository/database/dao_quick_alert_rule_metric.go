// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"strings"
)

// AlertMetricsData provide PQL corresponding to user selectable metrics
type AlertMetricsData struct {
	Id int `json:"-" gorm:"primaryKey;autoIncrement"`

	Name   string `json:"name" gorm:"not null;type:varchar(100);column:name"`
	NameEN string `json:"-" gorm:"type:varchar(100);column:name_en"`
	PQL    string `json:"pql" gorm:"not null;type:varchar(5000);column:pql"`
	Unit   string `json:"unit" gorm:"not null;type:varchar(100);column:unit"`
	Group  string `json:"group" gorm:"not null;type:varchar(100);column:group"`
}

func (a *AlertMetricsData) TableName() string {
	return "quick_alert_rule_metric"
}

// ListQuickMutationMetric list of all quick metrics
func (repo *daoRepo) ListQuickAlertRuleMetric(lang string) ([]AlertMetricsData, error) {
	var quickAlertMetrics []AlertMetricsData
	err := repo.db.Model(&AlertMetricsData{}).
		Select(getQuickAlertRuleNameField(lang), "pql", "unit", "group").
		Scan(&quickAlertMetrics).
		Error
	return quickAlertMetrics, err
}

func getQuickAlertRuleNameField(lang string) string {
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `name_en AS "name"`
	}
	// if strings.HasPrefix(lang, "zh") { // zh_CN,zh
	// 	return "name"
	// }
	return `"name"`
}
