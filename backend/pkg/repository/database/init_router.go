package database

import (
	"gorm.io/gorm"
)

// initRouterData TODO Add mapping of router to feature when permission control is required
func (repo *daoRepo) initRouterData() error {
	routers := []Router{
		{RouterTo: "/service", HideTimeSelector: false},
		{RouterTo: "/logs/fault-site", HideTimeSelector: true},
		{RouterTo: "/logs/full", HideTimeSelector: false},
		{RouterTo: "/system-dashboard", HideTimeSelector: false},
		{RouterTo: "/basic-dashboard", HideTimeSelector: false},
		{RouterTo: "/application-dashboard", HideTimeSelector: false},
		{RouterTo: "/middleware-dashboard", HideTimeSelector: false},
		{RouterTo: "/alerts", HideTimeSelector: true},
		{RouterTo: "/config", HideTimeSelector: true},
		{RouterTo: "/system/user-manage", HideTimeSelector: true},
		{RouterTo: "/system/menu-manage", HideTimeSelector: false},
		{RouterTo: "/trace/fault-site", HideTimeSelector: true},
		{RouterTo: "/trace/full", HideTimeSelector: true},
		{RouterTo: "/system/added/", HideTimeSelector: false},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Router{}); err != nil {
			return err
		}

		var existingRouter, toAdd []Router
		var toDelete []int

		if err := tx.Where("custom = ?", false).Find(&existingRouter).Error; err != nil {
			return err
		}

		existingMap := make(map[string]Router)
		for _, router := range existingRouter {
			existingMap[router.RouterTo] = router
		}

		for _, router := range routers {
			if _, exists := existingMap[router.RouterTo]; !exists {
				toAdd = append(toAdd, router)
			} else {
				delete(existingMap, router.RouterTo)
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

		if len(toDelete) > 0 {
			err := tx.Model(&Router{}).Where("router_id in ?", toDelete).Delete(nil).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
