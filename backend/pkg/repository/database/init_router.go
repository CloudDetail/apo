// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"gorm.io/gorm"
)

var validRouters = []Router{
	{RouterTo: "/service", HideTimeSelector: false},
	{RouterTo: "/logs/fault-site", HideTimeSelector: true},
	{RouterTo: "/logs/full", HideTimeSelector: false},
	{RouterTo: "/system-dashboard", HideTimeSelector: false},
	{RouterTo: "/basic-dashboard", HideTimeSelector: false},
	{RouterTo: "/application-dashboard", HideTimeSelector: false},
	{RouterTo: "/middleware-dashboard", HideTimeSelector: false},
	{RouterTo: "/alerts/rule", HideTimeSelector: true},
	{RouterTo: "/alerts/notify", HideTimeSelector: true},
	{RouterTo: "/integration/data", HideTimeSelector: true},
	{RouterTo: "/integration/alerts", HideTimeSelector: true},
	{RouterTo: "/config", HideTimeSelector: true},
	{RouterTo: "/system/user-manage", HideTimeSelector: true},
	{RouterTo: "/system/menu-manage", HideTimeSelector: true},
	{RouterTo: "/trace/fault-site", HideTimeSelector: true},
	{RouterTo: "/trace/full", HideTimeSelector: true},
	{RouterTo: "/system/data-group", HideTimeSelector: true},
	{RouterTo: "/system/config", HideTimeSelector: true},
	{RouterTo: "/system/team", HideTimeSelector: true},
	{RouterTo: "/alerts/events", HideTimeSelector: false},
	{RouterTo: "/workflows", HideTimeSelector: true},
	{RouterTo: "/system/role-manage", HideTimeSelector: true},
	{RouterTo: "/service/info", HideTimeSelector: false},
	{RouterTo: "/integration/data/settings", HideTimeSelector: true},
	{RouterTo: "/user", HideTimeSelector: true},
}

// initRouterData TODO Add mapping of router to feature when permission control is required
func (repo *daoRepo) initRouterData() error {
	return repo.Admin().Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Router{}); err != nil {
			return err
		}

		var existingRouter []Router
		if err := tx.Where("custom = ?", false).Find(&existingRouter).Error; err != nil {
			return err
		}

		existingMap := make(map[string]Router)
		for _, router := range existingRouter {
			existingMap[router.RouterTo] = router
		}

		var toAdd []Router
		var toUpdate []Router
		toDelete := make([]int, 0)

		for _, router := range validRouters {
			if existing, exists := existingMap[router.RouterTo]; exists {
				if existing.HideTimeSelector != router.HideTimeSelector {
					existing.HideTimeSelector = router.HideTimeSelector
					toUpdate = append(toUpdate, existing)
				}
				delete(existingMap, router.RouterTo)
			} else {
				toAdd = append(toAdd, router)
			}
		}

		for _, router := range existingMap {
			toDelete = append(toDelete, router.RouterID)
		}

		if len(toAdd) > 0 {
			if err := tx.Create(&toAdd).Error; err != nil {
				return err
			}
		}

		if len(toUpdate) > 0 {
			for _, router := range toUpdate {
				if err := tx.Model(&Router{}).
					Where("router_id = ?", router.RouterID).
					Updates(map[string]interface{}{
						"hide_time_selector": router.HideTimeSelector,
					}).Error; err != nil {
					return err
				}
			}
		}

		if len(toDelete) > 0 {
			if err := tx.Model(&Router{}).Where("router_id IN ?", toDelete).Delete(nil).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
