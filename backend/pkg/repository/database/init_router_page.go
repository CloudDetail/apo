// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	"gorm.io/gorm"
)

func isValidRouter(routerTo string) bool {
	for _, router := range validRouters {
		if router.RouterTo == routerTo {
			return true
		}
	}

	return false
}

func isValidPageURL(url string) bool {
	for _, page := range validPages {
		if page.Url == url {
			return true
		}
	}

	return false
}

// initRouterPage init router-insertPage mapping records.
func (repo *daoRepo) initRouterPage() error {
	type page struct {
		url      string
		language string
	}

	// For build-in dashboards, make sure each language has at least one page even if the language doesn't match
	routerPageMap := map[string][]page{
		"/system-dashboard": {
			{url: "grafana/d/k8s_views_global/e99b86-e7bea4-e680bb-e8a788", language: "zh"},
			{url: "grafana/d/d065c262fbbe43/cluster-overview", language: "en"},
		},
		"/basic-dashboard": {
			{url: "grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5", language: "zh"},
			{url: "grafana/d/bba60ba1600c34/infrastructure-metrics", language: "en"},
		},
		"/application-dashboard": {
			{url: "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", language: "zh"},
			{url: "grafana/d/3ab420aae391a1/originx-polaris-metrics", language: "en"},
		},
		"/middleware-dashboard": {
			{url: "grafana/dashboards/f/edwu5b9rkv94wb-zh/", language: "zh"},
			{url: "grafana/dashboards/f/edwu5b9rkv94wb-en/", language: "en"},
		},
		"/trace/full": {
			{url: "jaeger/search", language: "zh"},
			{url: "jaeger/search", language: "en"},
		},
	}
	return repo.Admin().Transaction(func(tx *gorm.DB) error {
		var routerIDs, pageIDs []int
		if err := tx.Model(&Router{}).Select("router_id").Find(&routerIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&InsertPage{}).Select("page_id").Find(&pageIDs).Error; err != nil {
			return err
		}

		// delete mapping whose router or page has been already deleted
		err := tx.Model(&RouterInsertPage{}).Where("router_id NOT IN ? OR page_id NOT IN ?", routerIDs, pageIDs).Delete(nil).Error
		if err != nil {
			return err
		}
		for router, pages := range routerPageMap {
			if !isValidRouter(router) {
				continue
			}

			var routerID int
			err := tx.Model(&Router{}).Select("router_id").Where("router_to = ?", router).First(&routerID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// router doesn't exist, skip
				continue
			} else if err != nil {
				return err
			}

			if routerID <= 0 {
				continue
			}

			validPages := make([]page, 0, len(pages))
			for _, page := range pages {
				if isValidPageURL(page.url) {
					validPages = append(validPages, page)
				}
			}

			for _, page := range validPages {
				var pageID int
				err = tx.Model(&InsertPage{}).Select("page_id").Where("url = ?", page.url).First(&pageID).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// page doesn't exist, skip
					continue
				} else if err != nil {
					return err
				}

				if pageID <= 0 {
					continue
				}

				routerPage := RouterInsertPage{
					PageID:   pageID,
					RouterID: routerID,
					Language: page.language,
				}

				var count int64
				err := tx.Model(&RouterInsertPage{}).Where(`router_id = ? AND page_id = ? AND "language" = ?`, routerID, pageID, page.language).Count(&count).Error
				if err != nil {
					return err
				}

				if count > 0 {
					continue
				}

				err = tx.Create(&routerPage).Error
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
