// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) initMenuItems() error {
	menuItems := []MenuItem{
		{Key: "service", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/service.svg"},
		{Key: "logs", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/log.svg"},
		{Key: "faultSite"},
		{Key: "full"},
		{Key: "trace", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/trace.svg"},
		{Key: "faultSiteTrace"},
		{Key: "fullTrace"},
		{Key: "system", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"},
		{Key: "basic", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"},
		{Key: "application", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"},
		{Key: "middleware", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"},
		{Key: "alerts", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/alert.svg"},
		{Key: "config", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/setting.svg"},
		{Key: "manage", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/system.svg"},
		{Key: "userManage"},
		{Key: "menuManage"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&MenuItem{}); err != nil {
			return err
		}
		// Menu item might include item which not support to existing
		// but the mapping between item and feature will be deleted.
		for _, menuItem := range menuItems {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				UpdateAll: true,
			}).Create(&menuItem).Error; err != nil {
				return err
			}
		}

		relations := map[string]string{
			"faultSite":      "logs",
			"full":           "logs",
			"faultSiteTrace": "trace",
			"fullTrace":      "trace",
			"userManage":     "manage",
			"menuManage":     "manage",
		}

		// update parent_id
		for childKey, parentKey := range relations {
			var parent MenuItem
			if err := tx.Where("key = ?", parentKey).First(&parent).Error; err != nil {
				return err
			}

			if err := tx.Model(&MenuItem{}).Where("key = ?", childKey).Update("parent_id", parent.ItemID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
