// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// initRouterData TODO Add mapping of router to feature when permission control is required
func (repo *daoRepo) initRouterData() error {
	routers := []Router{
		{RouterTo: "/service", HideTimeSelector: false, MenuItemKey: "service"},
		{RouterTo: "/logs/fault-site", HideTimeSelector: true, MenuItemKey: "faultSite"},
		{RouterTo: "/logs/full", HideTimeSelector: false, MenuItemKey: "full"},
		{RouterTo: "/system-dashboard", HideTimeSelector: false, MenuItemKey: "system"},
		{RouterTo: "/basic-dashboard", HideTimeSelector: false, MenuItemKey: "basic"},
		{RouterTo: "/application-dashboard", HideTimeSelector: false, MenuItemKey: "application"},
		{RouterTo: "/middleware-dashboard", HideTimeSelector: false, MenuItemKey: "middleware"},
		{RouterTo: "/alerts", HideTimeSelector: true, MenuItemKey: "alerts"},
		{RouterTo: "/config", HideTimeSelector: true, MenuItemKey: "config"},
		{RouterTo: "/system/user-manage", HideTimeSelector: true, MenuItemKey: "userManage"},
		{RouterTo: "/system/menu-manage", HideTimeSelector: false, MenuItemKey: "menuManage"},
		{RouterTo: "/trace/fault-site", HideTimeSelector: true, MenuItemKey: "faultSiteTrace"},
		{RouterTo: "/trace/full", HideTimeSelector: true, MenuItemKey: "fullTrace"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Router{}); err != nil {
			return err
		}
		for _, router := range routers {
			var menuItem MenuItem
			if err := tx.Where("key = ?", router.MenuItemKey).Find(&menuItem).Error; err != nil {
				return err
			}

			// menu item doesn't exist
			if menuItem.ItemID == 0 {
				continue
			}
			router.MenuItemID = menuItem.ItemID
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "menu_item_id"}},
				UpdateAll: true,
			}).Create(&router).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
