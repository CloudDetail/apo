// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"slices"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
)

type ServiceNameRuleCondition struct {
	ID        int    `gorm:"column:id;primary_key;auto_increment" json:"id"`
	RuleID    int    `gorm:"column:rule_id;" json:"ruleId"`
	Key       string `gorm:"column:key;type:varchar(20)" json:"key"`
	MatchType string `gorm:"column:match_type;type:varchar(20)" json:"matchType"`
	Value     string `gorm:"column:value;type:varchar(1000)" json:"value"`
}

func (condition *ServiceNameRuleCondition) Match(labels map[string]string) bool {
	labelValue, ok := labels[condition.Key]
	if !ok {
		return false
	}
	switch condition.MatchType {
		case "equals":
			return labelValue == condition.Value
		case "startsWith":
			return strings.HasPrefix(labelValue, condition.Value)
		case "endsWith":
			return strings.HasSuffix(labelValue, condition.Value)
		case "contains":
			return strings.Contains(labelValue, condition.Value)
		case "has":
			return slices.Contains(strings.Split(labelValue, ","), condition.Value)
		default:
			return false
	}
}

func (ServiceNameRuleCondition) TableName() string {
	return "service_name_rule_condition"
}

func (repo *daoRepo) UpsertServiceNameRuleCondition(ctx core.Context, condition *ServiceNameRuleCondition) error {
	if condition.ID > 0 {
		return repo.GetContextDB(ctx).Model(&ServiceNameRuleCondition{}).Where("id = ?", condition.ID).Updates(condition).Error
	}
	return repo.GetContextDB(ctx).Create(condition).Error
}

func (repo *daoRepo) ListAllServiceNameRuleCondition(ctx core.Context) ([]ServiceNameRuleCondition, error) {
	var conditions []ServiceNameRuleCondition
	err := repo.GetContextDB(ctx).
		Model(&ServiceNameRuleCondition{}).
		Order("rule_id ASC").
		Scan(&conditions).Error
	return conditions, err
}
