package database

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// initRouterPage init router-insertPage mapping records.
func (repo *daoRepo) initRouterPage() error {
	routerPageMap := map[string]string{
		"/system-dashboard":      "grafana/d/k8s_views_global/e99b86-e7bea4-e680bb-e8a788",
		"/basic-dashboard":       "grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5",
		"/application-dashboard": "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level",
		"/middleware-dashboard":  "grafana/dashboards/f/edwu5b9rkv94wb/",
		"/trace/full":            "/jaeger/search",
	}
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&RouterInsertPage{}); err != nil {
			return err
		}

		var routerIDs, pageIDs []int
		if err := tx.Model(&Router{}).Select("router_id").Find(&routerIDs).Error; err != nil {
			return err
		}

		// delete mapping whose router or page has been already deleted
		if err := tx.Model(&InsertPage{}).Select("page_id").Find(&pageIDs).Error; err != nil {
			return err
		}

		err := tx.Model(&RouterInsertPage{}).Where("router_id NOT IN ? OR page_id NOT IN ?", routerIDs, pageIDs).Delete(nil).Error
		if err != nil {
			return err
		}
		for router, page := range routerPageMap {
			var routerID, pageID int
			err := tx.Model(&Router{}).Select("router_id").Where("router_to = ?", router).First(&routerID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// router doesn't exist, skip
				continue
			} else if err != nil {
				return err
			}

			err = tx.Model(&InsertPage{}).Select("page_id").Where("url = ?", page).First(&pageID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// page doesn't exist, skip
				continue
			} else if err != nil {
				return err
			}

			routerPage := RouterInsertPage{
				PageID:   pageID,
				RouterID: routerID,
			}

			err = tx.Clauses(clause.OnConflict{
				// one router should have exactly one page.
				Columns:   []clause.Column{{Name: "router_id"}},
				UpdateAll: true,
			}).Create(&routerPage).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}
