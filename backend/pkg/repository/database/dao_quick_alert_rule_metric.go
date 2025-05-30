// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// AlertMetricsData provide PQL corresponding to user selectable metrics
type AlertMetricsData struct {
	Id int `json:"-" gorm:"primaryKey;autoIncrement"`

	Name   string `json:"name" gorm:"not null;type:varchar(255);column:name"`
	NameEN string `json:"-" gorm:"type:varchar(255);column:name_en"`
	PQL    string `json:"pql" gorm:"not null;type:varchar(5000);column:pql"`
	Unit   string `json:"unit" gorm:"not null;type:varchar(255);column:unit"`
	Group  string `json:"group" gorm:"not null;type:varchar(255);column:group"`
}

func (a *AlertMetricsData) TableName() string {
	return "quick_alert_rule_metric"
}

// ListQuickMutationMetric list of all quick metrics
func (repo *daoRepo) ListQuickAlertRuleMetric(ctx core.Context) ([]AlertMetricsData, error) {
	var quickAlertMetrics []AlertMetricsData
	err := repo.GetContextDB(ctx).Model(&AlertMetricsData{}).
		Select(getQuickAlertRuleNameField(ctx), "pql", "unit", "group").
		Scan(&quickAlertMetrics).
		Error
	return quickAlertMetrics, err
}

func getQuickAlertRuleNameField(ctx core.Context) string {
	var lang = code.LANG_EN
	if ctx != nil {
		lang = ctx.LANG()
	}
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `name_en AS "name"`
	}
	return `"name"`
}
