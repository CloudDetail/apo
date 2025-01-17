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
		{MenuItem: MenuItem{Key: "service", Order: 1}, RouterKey: "/service"},
		{MenuItem: MenuItem{Key: "logs", Order: 2}},
		{MenuItem: MenuItem{Key: "faultSite", Order: 3}, RouterKey: "/logs/fault-site"},
		{MenuItem: MenuItem{Key: "full", Order: 4}, RouterKey: "/logs/full"},
		{MenuItem: MenuItem{Key: "trace", Order: 5}},
		{MenuItem: MenuItem{Key: "faultSiteTrace", Order: 6}, RouterKey: "/trace/fault-site"},
		{MenuItem: MenuItem{Key: "fullTrace", Order: 7}, RouterKey: "/trace/full"},
		{MenuItem: MenuItem{Key: "system", Order: 8}, RouterKey: "/system-dashboard"},
		{MenuItem: MenuItem{Key: "basic", Order: 9}, RouterKey: "/basic-dashboard"},
		{MenuItem: MenuItem{Key: "application", Order: 10}, RouterKey: "/application-dashboard"},
		{MenuItem: MenuItem{Key: "middleware", Order: 11}, RouterKey: "/middleware-dashboard"},
		{MenuItem: MenuItem{Key: "alerts", Order: 15}},
		{MenuItem: MenuItem{Key: "alertsRule", Order: 16}, RouterKey: "/alerts/rule"},
		{MenuItem: MenuItem{Key: "alertsNotify", Order: 17}, RouterKey: "/alerts/notify"},
		{MenuItem: MenuItem{Key: "integration", Order: 20}},
		{MenuItem: MenuItem{Key: "alertsIntegration", Order: 21}, RouterKey: "/integration/alerts"},
		{MenuItem: MenuItem{Key: "config", Order: 25}, RouterKey: "/config"},
		{MenuItem: MenuItem{Key: "manage", Order: 30}},
		{MenuItem: MenuItem{Key: "userManage", Order: 31}, RouterKey: "/system/user-manage"},
		{MenuItem: MenuItem{Key: "menuManage", Order: 32}, RouterKey: "/system/menu-manage"},
		{MenuItem: MenuItem{Key: "systemConfig", Order: 33}, RouterKey: "/system/config"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
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
			"faultSite":         "logs",
			"full":              "logs",
			"faultSiteTrace":    "trace",
			"fullTrace":         "trace",
			"userManage":        "manage",
			"menuManage":        "manage",
			"alertsRule":        "alerts",
			"alertsNotify":      "alerts",
			"systemConfig":      "manage",
			"alertsIntegration": "integration",
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
