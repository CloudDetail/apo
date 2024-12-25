package database

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

func (repo *daoRepo) initFeatureMenuItems() error {
	featureMenuMapping := []struct {
		FeatureName string
		MenuKey     string
	}{
		{"服务概览", "service"},
		{"故障现场日志", "faultSite"},
		{"全量日志", "full"},
		{"故障现场链路", "faultSiteTrace"},
		{"全量链路", "fullTrace"},
		{"全局资源大盘", "system"},
		{"应用基础设施大盘", "basic"},
		{"应用指标大盘", "application"},
		{"中间件大盘", "middleware"},
		{"告警规则", "alerts"},
		{"配置中心", "config"},
		{"用户管理", "userManage"},
		{"菜单管理", "menuManage"},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&FeatureMapping{}); err != nil {
			return err
		}

		var featureIDs []int
		if err := tx.Model(&Feature{}).Select("feature_id").Find(&featureIDs).Error; err != nil {
			return err
		}

		// delete mapping whose feature has been already deleted
		if err := tx.Model(&FeatureMapping{}).Where("feature_id not in ?", featureIDs).Delete(nil).Error; err != nil {
			return err
		}

		for _, mapping := range featureMenuMapping {
			var feature Feature
			if err := tx.Where("feature_name = ?", mapping.FeatureName).First(&feature).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}

			var menuItem MenuItem
			if err := tx.Where("key = ?", mapping.MenuKey).First(&menuItem).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}

			var count int64
			if err := tx.Model(&FeatureMapping{}).
				Where("feature_id = ? AND mapped_id = ? AND mapped_type = ?", feature.FeatureID, menuItem.ItemID, model.MAPPED_TYP_MENU).
				Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				featureMenuItem := FeatureMapping{
					FeatureID:  feature.FeatureID,
					MappedID:   menuItem.ItemID,
					MappedType: model.MAPPED_TYP_MENU,
				}
				if err := tx.Create(&featureMenuItem).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// TODO add mapping of feature and router or api
