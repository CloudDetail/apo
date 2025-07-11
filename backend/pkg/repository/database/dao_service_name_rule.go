// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

type ServiceNameRule struct {
	ID          int    `gorm:"column:id;primary_key;auto_increment" json:"id"`
	Service     string `gorm:"column:service_name;type:varchar(100)" json:"serviceName"`
	ClusterId   string `gorm:"column:cluster_id;type:varchar(100)" json:"clusterId"`
}

func (ServiceNameRule) TableName() string {
	return "service_name_rule"
}

func (repo *daoRepo) CreateServiceNameRule(ctx core.Context, rule *ServiceNameRule) error {
	return repo.GetContextDB(ctx).Create(&rule).Error
}

func (repo *daoRepo) ListAllServiceNameRule(ctx core.Context) ([]ServiceNameRule, error) {
	var nameRules []ServiceNameRule
	err := repo.GetContextDB(ctx).
		Model(&ServiceNameRule{}).
		Order("cluster_id, service_name").
		Scan(&nameRules).Error
	return nameRules, err
}

func (repo *daoRepo) ServiceNameRuleExists(ctx core.Context, ruleId int) (bool, error) {
	var count int64

	query := repo.GetContextDB(ctx).
		Model(&ServiceNameRule{}).
		Where("id = ?", ruleId)
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *daoRepo) DeleteServiceNameRule(ctx core.Context, ruleId int) error {
	if err := repo.GetContextDB(ctx).Model(&ServiceNameRuleCondition{}).Where("rule_id = ?", ruleId).Delete(nil).Error; err != nil {
		return err
	}
	return repo.GetContextDB(ctx).Model(&ServiceNameRule{}).Where("id = ?", ruleId).Delete(nil).Error
}