// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"gorm.io/gorm"
)

func (repo *daoRepo) initFeatureMenuItems() error {
	featureMenuMapping := []struct {
		FeatureName string
		MenuKey     string
	}{
		{"服务概览", "service"},
		{"故障现场日志", "faultSite"},
		{"全量日志", "full"},
		{"故障现场链路", "faultSiteTrace"},
		{"全量链路", "fullTrace"},
		{"全局资源大盘", "system"},
		{"应用基础设施大盘", "basic"},
		{"应用指标大盘", "application"},
		{"中间件大盘", "middleware"},
		{"告警规则", "alerts"},
		{"配置中心", "config"},
		{"用户管理", "userManage"},
		{"菜单管理", "menuManage"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&FeatureMenuItem{}); err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM feature_menu_item WHERE feature_id NOT IN (SELECT feature_id FROM feature)").Error; err != nil {
			return err
		}

		for _, mapping := range featureMenuMapping {
			var feature Feature
			if err := tx.Where("feature_name = ?", mapping.FeatureName).First(&feature).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}

			var menuItem MenuItem
			if err := tx.Where("key = ?", mapping.MenuKey).First(&menuItem).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}

			var count int64
			if err := tx.Model(&FeatureMenuItem{}).
				Where("feature_id = ? AND menu_item_id = ?", feature.FeatureID, menuItem.ItemID).
				Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				featureMenuItem := FeatureMenuItem{
					FeatureID:  feature.FeatureID,
					MenuItemID: menuItem.ItemID,
				}
				if err := tx.Create(&featureMenuItem).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
