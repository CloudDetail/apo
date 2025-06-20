// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var validMenuItemMappings = []struct {
	MenuItem
	RouterKey string
}{
	{MenuItem: MenuItem{Key: "service", Order: 1}, RouterKey: "/service"},
	{MenuItem: MenuItem{Key: "workflows", Order: 3}, RouterKey: "/workflows"},
	{MenuItem: MenuItem{Key: "alerts", Order: 4}},
	{MenuItem: MenuItem{Key: "logs", Order: 5}},
	{MenuItem: MenuItem{Key: "faultSite", Order: 7}, RouterKey: "/logs/fault-site"},
	{MenuItem: MenuItem{Key: "full", Order: 9}, RouterKey: "/logs/full"},
	{MenuItem: MenuItem{Key: "trace", Order: 11}},
	{MenuItem: MenuItem{Key: "faultSiteTrace", Order: 13}, RouterKey: "/trace/fault-site"},
	{MenuItem: MenuItem{Key: "fullTrace", Order: 15}, RouterKey: "/trace/full"},
	{MenuItem: MenuItem{Key: "system", Order: 17}, RouterKey: "/system-dashboard"},
	{MenuItem: MenuItem{Key: "basic", Order: 19}, RouterKey: "/basic-dashboard"},
	{MenuItem: MenuItem{Key: "application", Order: 21}, RouterKey: "/application-dashboard"},
	{MenuItem: MenuItem{Key: "middleware", Order: 23}, RouterKey: "/middleware-dashboard"},
	{MenuItem: MenuItem{Key: "alertsRule", Order: 27}, RouterKey: "/alerts/rule"},
	{MenuItem: MenuItem{Key: "alertsNotify", Order: 29}, RouterKey: "/alerts/notify"},
	{MenuItem: MenuItem{Key: "alertEvents", Order: 31}, RouterKey: "/alerts/events"},
	{MenuItem: MenuItem{Key: "integration", Order: 33}},
	{MenuItem: MenuItem{Key: "dataIntegration", Order: 35}, RouterKey: "/integration/data"},
	{MenuItem: MenuItem{Key: "alertsIntegration", Order: 37}, RouterKey: "/integration/alerts"},
	{MenuItem: MenuItem{Key: "config", Order: 39}, RouterKey: "/config"},
	{MenuItem: MenuItem{Key: "manage", Order: 41}},
	{MenuItem: MenuItem{Key: "userManage", Order: 43}, RouterKey: "/system/user-manage"},
	{MenuItem: MenuItem{Key: "menuManage", Order: 45}, RouterKey: "/system/menu-manage"},
	{MenuItem: MenuItem{Key: "dataGroup", Order: 49}, RouterKey: "/system/data-group"},
	{MenuItem: MenuItem{Key: "team", Order: 51}, RouterKey: "/system/team"},
	{MenuItem: MenuItem{Key: "role", Order: 52}, RouterKey: "/system/role-manage"},
}

func (repo *daoRepo) initMenuItems(ctx core.Context) error {

	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Menu item might include item which not support to existing
		// but the mapping between item and feature will be deleted
		// because once a menu was deleted, the feature should also be deleted.
		for i, menuItem := range validMenuItemMappings {
			if len(menuItem.RouterKey) == 0 {
				continue
			}
			var routerID int
			err := tx.Model(&Router{}).Select("router_id").Where("router_to = ?", menuItem.RouterKey).First(&routerID).Error
			if err != nil {
				return err
			}
			validMenuItemMappings[i].RouterID = routerID
		}

		menuItems := make([]MenuItem, len(validMenuItemMappings))
		for i := range validMenuItemMappings {
			menuItems[i] = validMenuItemMappings[i].MenuItem
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
			"dataIntegration":   "integration",
			"alertsIntegration": "integration",
			"dataGroup":         "manage",
			"team":              "manage",
			"alertEvents":       "alerts",
			"role":              "manage",
		}

		// update parent_id
		for childKey, parentKey := range relations {
			var parent MenuItem
			if err := tx.Where(`"key" = ?`, parentKey).First(&parent).Error; err != nil {
				return err
			}

			if err := tx.Model(&MenuItem{}).Where(`"key" = ?`, childKey).Update("parent_id", parent.ItemID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
