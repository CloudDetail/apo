// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) initMenuItems() error {
	menuItemsMapping := []struct {
		MenuItem
		RouterKey string
	}{
		{MenuItem: MenuItem{Key: "service", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/service.svg"}, RouterKey: "/service"},
		{MenuItem: MenuItem{Key: "logs", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/log.svg"}},
		{MenuItem: MenuItem{Key: "faultSite"}, RouterKey: "/logs/fault-site"},
		{MenuItem: MenuItem{Key: "full"}, RouterKey: "/logs/full"},
		{MenuItem: MenuItem{Key: "trace", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/trace.svg"}},
		{MenuItem: MenuItem{Key: "faultSiteTrace"}, RouterKey: "/trace/fault-site"},
		{MenuItem: MenuItem{Key: "fullTrace"}, RouterKey: "/trace/full"},
		{MenuItem: MenuItem{Key: "system", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"}, RouterKey: "/system-dashboard"},
		{MenuItem: MenuItem{Key: "basic", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"}, RouterKey: "/basic-dashboard"},
		{MenuItem: MenuItem{Key: "application", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"}, RouterKey: "/application-dashboard"},
		{MenuItem: MenuItem{Key: "middleware", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg"}, RouterKey: "/middleware-dashboard"},
		{MenuItem: MenuItem{Key: "alerts", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/alert.svg"}},
		{MenuItem: MenuItem{Key: "alertsRule"}, RouterKey: "/alerts/rule"},
		{MenuItem: MenuItem{Key: "alertsNotify"}, RouterKey: "/alerts/notify"},
		{MenuItem: MenuItem{Key: "config", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/setting.svg"}, RouterKey: "/config"},
		{MenuItem: MenuItem{Key: "manage", Icon: "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/system.svg"}},
		{MenuItem: MenuItem{Key: "userManage"}, RouterKey: "/system/user-manage"},
		{MenuItem: MenuItem{Key: "menuManage"}, RouterKey: "/system/menu-manage"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&MenuItem{}); err != nil {
			return err
		}
		// Menu item might include item which not support to existing
		// but the mapping between item and feature will be deleted
		// because once a menu was deleted, the feature should also be deleted.
		for i, menuItem := range menuItemsMapping {
			if len(menuItem.RouterKey) == 0 {
				continue
			}
			var routerID int
			err := tx.Model(&Router{}).Select("router_id").Where("router_to = ?", menuItem.RouterKey).First(&routerID).Error
			if err != nil {
				return err
			}
			menuItemsMapping[i].RouterID = routerID
		}

		menuItems := make([]MenuItem, len(menuItemsMapping))
		for i := range menuItemsMapping {
			menuItems[i] = menuItemsMapping[i].MenuItem
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			UpdateAll: true,
		}).Create(&menuItems).Error; err != nil {
			return err
		}

		relations := map[string]string{
			"faultSite":      "logs",
			"full":           "logs",
			"faultSiteTrace": "trace",
			"fullTrace":      "trace",
			"userManage":     "manage",
			"menuManage":     "manage",
			"alertsRule":     "alerts",
			"alertsNotify":   "alerts",
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
