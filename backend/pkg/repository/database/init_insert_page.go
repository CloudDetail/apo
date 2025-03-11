// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"gorm.io/gorm"
)

func (repo *daoRepo) initInsertPages() error {
	pages := []InsertPage{
		{Url: "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", Type: "grafana"},
		{Url: "grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5", Type: "grafana"},
		{Url: "grafana/dashboards/f/edwu5b9rkv94wb-zh/", Type: "grafana"},
		{Url: "grafana/d/05afd511b2fe54/service-middleware-metrics", Type: "grafana"},
		{Url: "grafana/d/ae3aqpssill34f/e69c8d-e58aa1-e8b083-e794a8-e4b8ad-e997b4-e4bbb6", Type: "grafana"},
		{Url: "jaeger/search", Type: "jaeger"},
		{Url: "grafana/d/d065c262fbbe43/cluster-overview", Type: "grafana"},
		{Url: "grafana/d/bba60ba1600c34/infrastructure-metrics", Type: "grafana"},
		{Url: "grafana/d/3ab420aae391a1/originx-polaris-metrics", Type: "grafana"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		var existingPage, toAdd []InsertPage
		var toDelete []int
		var toUpdate []InsertPage

		if err := tx.Where("custom = ?", false).Find(&existingPage).Error; err != nil {
			return err
		}

		existingMap := make(map[string]int)
		for _, page := range existingPage {
			existingMap[page.Url] = page.PageID
		}

		for _, page := range pages {
			if pageID, exists := existingMap[page.Url]; !exists {
				toAdd = append(toAdd, page)
			} else {
				page.PageID = pageID
				toUpdate = append(toUpdate, page)
				delete(existingMap, page.Url)
			}
		}

		for _, page := range existingMap {
			toDelete = append(toDelete, page)
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

		if len(toUpdate) > 0 {
			for _, page := range toUpdate {
				if err := tx.Model(&InsertPage{}).Where("page_id = ?", page.PageID).Updates(page).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
