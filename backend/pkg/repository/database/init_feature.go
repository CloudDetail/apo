// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

var validFeatures = []Feature{
	{FeatureName: "服务概览"},
	{FeatureName: "工作流"},
	{FeatureName: "日志检索"}, {FeatureName: "故障现场日志"}, {FeatureName: "全量日志"},
	{FeatureName: "链路追踪"}, {FeatureName: "故障现场链路"}, {FeatureName: "全量链路"},
	{FeatureName: "全局资源大盘"},
	{FeatureName: "应用基础设施大盘"},
	{FeatureName: "应用指标大盘"},
	{FeatureName: "中间件大盘"},
	{FeatureName: "告警管理"}, {FeatureName: "告警规则"}, {FeatureName: "告警通知"}, {FeatureName: "告警事件"},
	{FeatureName: "接入中心"}, {FeatureName: "数据接入"}, {FeatureName: "告警接入"},
	{FeatureName: "配置中心"},
	{FeatureName: "系统管理"}, {FeatureName: "用户管理"}, {FeatureName: "菜单管理"}, {FeatureName: "系统配置"},
	{FeatureName: "数据组管理"}, {FeatureName: "团队管理"}, {FeatureName: "角色管理"},
}

func (repo *daoRepo) initFeature(ctx core.Context) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		var existingFeatures []Feature
		if err := tx.Where("custom = ?", false).Find(&existingFeatures).Error; err != nil {
			return err
		}

		existingFeatureMap := make(map[string]int)
		for _, feature := range existingFeatures {
			existingFeatureMap[feature.FeatureName] = feature.FeatureID
		}

		newFeatureMap := make(map[string]struct{})
		for _, feature := range validFeatures {
			newFeatureMap[feature.FeatureName] = struct{}{}
		}

		for _, feature := range validFeatures {
			// Add new feature.
			if _, exists := existingFeatureMap[feature.FeatureName]; exists {
				continue
			}
			if err := tx.Create(&feature).Error; err != nil {
				return err
			}
		}

		// remove feature which not support to exist
		for featureName, featureID := range existingFeatureMap {
			if _, exists := newFeatureMap[featureName]; !exists {
				if err := tx.Where("feature_id = ?", featureID).Delete(&Feature{}).Error; err != nil {
					return err
				}
			}
		}

		// Add parent_id relationships
		parentChildMapping := map[string][]string{
			"日志检索": {"故障现场日志", "全量日志"},
			"链路追踪": {"故障现场链路", "全量链路"},
			"系统管理": {"用户管理", "菜单管理", "数据组管理", "系统配置", "团队管理", "角色管理"},
			"告警管理": {"告警规则", "告警通知", "告警事件", "告警事件详情"},
			"接入中心": {"数据接入", "告警接入"},
		}
		for parentName, childNames := range parentChildMapping {
			var parent Feature
			if err := tx.Where("feature_name = ?", parentName).First(&parent).Error; err != nil {
				return err
			}
			for _, childName := range childNames {
				if err := tx.Model(&Feature{}).Where("feature_name = ?", childName).Update("parent_id", parent.FeatureID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
