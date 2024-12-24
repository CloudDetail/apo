package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *daoRepo) initInsertPages() error {
	pages := []InsertPage{
		{Url: "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", Type: "grafana", MenuItemKey: "application"},
		{Url: "grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5", Type: "grafana", MenuItemKey: "basic"},
		{Url: "grafana/dashboards/f/edwu5b9rkv94wb/", Type: "grafana", MenuItemKey: "middleware"},
		{Url: "grafana/d/k8s_views_global/e99b86-e7bea4-e680bb-e8a788", Type: "grafana", MenuItemKey: "system"},
		{Url: "/jaeger/search", Type: "jaeger", MenuItemKey: "fullTrace"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&InsertPage{}); err != nil {
			return err
		}
		for _, page := range pages {
			var menuItem MenuItem
			if err := tx.Where("key = ?", page.MenuItemKey).Find(&menuItem).Error; err != nil {
				return err
			}

			// menu item doesn't exist
			if menuItem.ItemID == 0 {
				continue
			}
			page.MenuItemID = menuItem.ItemID
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "menu_item_id"}},
				UpdateAll: true,
			}).Create(&page).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
