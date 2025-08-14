// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"os"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type menuMapping struct {
	MenuItem
	RouterKey string
}

var validMenuItemMappings = []menuMapping{
	{MenuItem: MenuItem{Key: "agent", Order: 0}, RouterKey: "/agent"},
	{MenuItem: MenuItem{Key: "service", Order: 1}, RouterKey: "/service"},
	{MenuItem: MenuItem{Key: "workflows", Order: 3}, RouterKey: "/workflows"},
	{MenuItem: MenuItem{Key: "alerts", Order: 4}},
	{MenuItem: MenuItem{Key: "logs", Order: 5}},
	{MenuItem: MenuItem{Key: "analysis", Order: 6}, RouterKey: "/alert-analyze"},
	{MenuItem: MenuItem{Key: "faultSite", Order: 7}, RouterKey: "/logs/fault-site"},
	{MenuItem: MenuItem{Key: "full", Order: 9}, RouterKey: "/logs/full"},
	{MenuItem: MenuItem{Key: "trace", Order: 11}},
	{MenuItem: MenuItem{Key: "faultSiteTrace", Order: 13}, RouterKey: "/trace/fault-site"},
	{MenuItem: MenuItem{Key: "fullTrace", Order: 15}, RouterKey: "/trace/full"},
	{MenuItem: MenuItem{Key: "system", Order: 17}, RouterKey: "/system-dashboard"},
	{MenuItem: MenuItem{Key: "basic", Order: 19}, RouterKey: "/basic-dashboard"},
	{MenuItem: MenuItem{Key: "application", Order: 21}, RouterKey: "/application-dashboard"},
	{MenuItem: MenuItem{Key: "middleware", Order: 23}, RouterKey: "/middleware-dashboard"},
	{MenuItem: MenuItem{Key: "mysql", Order: 24}, RouterKey: "/mysql-dashboard"},
	{MenuItem: MenuItem{Key: "alertsRule", Order: 27}, RouterKey: "/alerts/rule"},
	{MenuItem: MenuItem{Key: "alertsNotify", Order: 29}, RouterKey: "/alerts/notify"},
	{MenuItem: MenuItem{Key: "alertEvents", Order: 31}, RouterKey: "/alerts/events"},
	{MenuItem: MenuItem{Key: "healthy", Order: 32}, RouterKey: "/healthy-service"},
	{MenuItem: MenuItem{Key: "alertEventDetail", Order: 32}, RouterKey: "/alerts/events/detail/:alertID/:eventID"},
	{MenuItem: MenuItem{Key: "integration", Order: 33}},
	{MenuItem: MenuItem{Key: "dataIntegration", Order: 35}, RouterKey: "/integration/data"},
	{MenuItem: MenuItem{Key: "alertsIntegration", Order: 37}, RouterKey: "/integration/alerts"},
	{MenuItem: MenuItem{Key: "dataplane", Order: 38}, RouterKey: "/dataplane"},
	{MenuItem: MenuItem{Key: "config", Order: 39}, RouterKey: "/config"},
	{MenuItem: MenuItem{Key: "manage", Order: 41}},
	{MenuItem: MenuItem{Key: "userManage", Order: 43}, RouterKey: "/system/user-manage"},
	{MenuItem: MenuItem{Key: "menuManage", Order: 45}, RouterKey: "/system/menu-manage"},
	{MenuItem: MenuItem{Key: "team", Order: 51}, RouterKey: "/system/team"},
	{MenuItem: MenuItem{Key: "role", Order: 52}, RouterKey: "/system/role-manage"},
	{MenuItem: MenuItem{Key: "dataGroup", Order: 49}, RouterKey: "/system/data-group"},
}

var sreMenuItemMappings = []menuMapping{
	{MenuItem: MenuItem{Key: "agent", Order: 0}, RouterKey: "/agent"},
	{MenuItem: MenuItem{Key: "integration", Order: 33}},
	{MenuItem: MenuItem{Key: "dataIntegration", Order: 35}, RouterKey: "/probe-management"},
	{MenuItem: MenuItem{Key: "alertsIntegration", Order: 37}, RouterKey: "/integration/alerts"},
	{MenuItem: MenuItem{Key: "dataplane", Order: 38}, RouterKey: "/dataplane"},
	{MenuItem: MenuItem{Key: "alerts", Order: 4}},
	{MenuItem: MenuItem{Key: "alertsNotify", Order: 29}, RouterKey: "/alerts/notify"},
	{MenuItem: MenuItem{Key: "alertEvents", Order: 31}, RouterKey: "/alerts/events"},
	{MenuItem: MenuItem{Key: "manage", Order: 41}},
	{MenuItem: MenuItem{Key: "userManage", Order: 43}, RouterKey: "/system/user-manage"},
	{MenuItem: MenuItem{Key: "menuManage", Order: 45}, RouterKey: "/system/menu-manage"},
	{MenuItem: MenuItem{Key: "team", Order: 51}, RouterKey: "/system/team"},
	{MenuItem: MenuItem{Key: "role", Order: 52}, RouterKey: "/system/role-manage"},
	{MenuItem: MenuItem{Key: "dataGroup", Order: 49}, RouterKey: "/system/data-group"},
}

func (repo *daoRepo) initMenuItems(ctx core.Context) error {

	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		productMode := os.Getenv("PRODUCT_MODE")
		productMenus := validMenuItemMappings
		if productMode == "sre_agent" {
			productMenus = sreMenuItemMappings
		}
		// Menu item might include item which not support to existing
		// but the mapping between item and feature will be deleted
		// because once a menu was deleted, the feature should also be deleted.
		for i, menuItem := range productMenus {
			if len(menuItem.RouterKey) == 0 {
				continue
			}
			var routerID int
			err := tx.Model(&Router{}).Select("router_id").Where("router_to = ?", menuItem.RouterKey).First(&routerID).Error
			if err != nil {
				return err
			}
			productMenus[i].RouterID = routerID
		}

		menuItems := make([]MenuItem, len(productMenus))
		for i := range productMenus {
			menuItems[i] = productMenus[i].MenuItem
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
			"dataIntegration":   "integration",
			"alertsIntegration": "integration",
			"dataplane":         "integration",
			"dataGroup":         "manage",
			"team":              "manage",
			"alertEvents":       "alerts",
			"role":              "manage",
		}

		// update parent_id
		for childKey, parentKey := range relations {
			var parent MenuItem
			if err := tx.Where(`"key" = ?`, parentKey).Find(&parent).Error; err != nil {
				return err
			}
			if parent.ItemID == 0 {
				continue
			}
			if err := tx.Model(&MenuItem{}).Where(`"key" = ?`, childKey).Update("parent_id", parent.ItemID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
