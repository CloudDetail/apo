package database

import (
	"gorm.io/gorm"
)

func (repo *daoRepo) initInsertPages() error {
	pages := []InsertPage{
		{Url: "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", Type: "grafana"},
		{Url: "grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5"},
		{Url: "grafana/dashboards/f/edwu5b9rkv94wb/", Type: "grafana"},
		{Url: "grafana/d/k8s_views_global/e99b86-e7bea4-e680bb-e8a788", Type: "grafana"},
		{Url: "/jaeger/search", Type: "jaeger"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&InsertPage{}); err != nil {
			return err
		}

		var existingPage, toAdd []InsertPage
		var toDelete []int

		if err := tx.Where("custom = ?", false).Find(&existingPage).Error; err != nil {
			return err
		}

		existingMap := make(map[string]InsertPage)
		for _, page := range existingPage {
			existingMap[page.Url] = page
		}

		for _, page := range pages {
			if _, exists := existingMap[page.Url]; !exists {
				toAdd = append(toAdd, page)
			} else {
				delete(existingMap, page.Url)
			}
		}

		for _, page := range existingMap {
			toDelete = append(toDelete, page.PageID)
		}

		if len(toAdd) > 0 {
			if err := tx.Create(&toAdd).Error; err != nil {
				return err
			}
		}

		if len(toDelete) > 0 {
			err := tx.Model(&InsertPage{}).Where("page_id in ?", toDelete).Delete(nil).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}
