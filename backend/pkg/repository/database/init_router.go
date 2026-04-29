// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"os"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

var validRouters = []Router{
	{RouterTo: "/knowledge-bases", HideTimeSelector: true},
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
	// {RouterTo: "/", HideTimeSelector: true},
	{RouterTo: "/system/team", HideTimeSelector: true},
	{RouterTo: "/alerts/events", HideTimeSelector: false},
	{RouterTo: "/alerts/events/detail/:alertID/:eventID", HideTimeSelector: false},
	{RouterTo: "/workflows", HideTimeSelector: true},
	{RouterTo: "/system/role-manage", HideTimeSelector: true},
	{RouterTo: "/service/info", HideTimeSelector: false},
	{RouterTo: "/integration/data/settings", HideTimeSelector: true},
	{RouterTo: "/user", HideTimeSelector: true},
	{RouterTo: "/demo/service-config", HideTimeSelector: true},
	{RouterTo: "/mysql-dashboard", HideTimeSelector: false},
	{RouterTo: "/healthy-service", HideTimeSelector: false},
	{RouterTo: "/system/data-group", HideTimeSelector: true},
	{RouterTo: "/report", HideTimeSelector: true},
	{RouterTo: "/dataplane", HideTimeSelector: true},
	{RouterTo: "/agent", HideTimeSelector: true},
	{RouterTo: "/probe-installer", HideTimeSelector: true},
	{RouterTo: "/alert-analyze", HideTimeSelector: false},
	{RouterTo: "/alert-analyze/info", HideTimeSelector: false},
	{RouterTo: "/cause/report/:traceId/:spanId", HideTimeSelector: true},
	{RouterTo: "/cause/error-report/:traceId", HideTimeSelector: true},
	{RouterTo: "/alert-analyze/chatAI/:service/:endpoint", HideTimeSelector: false},
	{RouterTo: "/probe-management", HideTimeSelector: false},
	{RouterTo: "/probe-management/settings", HideTimeSelector: false},
}

var sreRouters = []Router{
	{RouterTo: "/agent", HideTimeSelector: true},
	{RouterTo: "/knowledge-bases", HideTimeSelector: true},
	{RouterTo: "/probe-management", HideTimeSelector: true},
	{RouterTo: "/integration/alerts", HideTimeSelector: true},
	{RouterTo: "/dataplane", HideTimeSelector: true},
	{RouterTo: "/alerts/notify", HideTimeSelector: true},
	{RouterTo: "/alerts/events", HideTimeSelector: false},
	{RouterTo: "/alerts/events/detail/:alertID/:eventID", HideTimeSelector: false},
	{RouterTo: "/system/data-group", HideTimeSelector: true},
	{RouterTo: "/system/team", HideTimeSelector: true},
	{RouterTo: "/system/user-manage", HideTimeSelector: true},
	{RouterTo: "/system/menu-manage", HideTimeSelector: true},
	{RouterTo: "/system/role-manage", HideTimeSelector: true},
	{RouterTo: "/cause/report/:traceId/:spanId", HideTimeSelector: true},
	{RouterTo: "/cause/error-report/:traceId", HideTimeSelector: true},
	{RouterTo: "/report", HideTimeSelector: true},
	{RouterTo: "/probe-management/settings", HideTimeSelector: false},
}

// initRouterData TODO Add mapping of router to feature when permission control is required
func (repo *daoRepo) initRouterData(ctx core.Context) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
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

		var toAdd, toUpdate []Router
		toDelete := make([]int, 0)
		productMode := os.Getenv("PRODUCT_MODE")
		productRouters := validRouters
		if productMode == "syncause" {
			productRouters = sreRouters
		}

		for _, router := range productRouters {
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
