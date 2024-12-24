package database

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

func (repo *daoRepo) initFeature() error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Feature{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&AuthPermission{}); err != nil {
			return err
		}

		var existingFeatures []Feature
		if err := tx.Where("custom = ?", false).Find(&existingFeatures).Error; err != nil {
			return err
		}

		existingFeatureMap := make(map[string]int)
		for _, feature := range existingFeatures {
			existingFeatureMap[feature.FeatureName] = feature.FeatureID
		}

		newFeatures := []Feature{
			{FeatureName: "服务概览"}, {FeatureName: "日志检索"}, {FeatureName: "故障现场日志"},
			{FeatureName: "全量日志"}, {FeatureName: "链路追踪"}, {FeatureName: "故障现场链路"},
			{FeatureName: "全量链路"}, {FeatureName: "全局资源大盘"}, {FeatureName: "应用基础设施大盘"},
			{FeatureName: "应用指标大盘"}, {FeatureName: "中间件大盘"}, {FeatureName: "告警规则"},
			{FeatureName: "配置中心"}, {FeatureName: "系统管理"}, {FeatureName: "用户管理"},
			{FeatureName: "菜单管理"},
		}

		newFeatureMap := make(map[string]struct{})
		for _, feature := range newFeatures {
			newFeatureMap[feature.FeatureName] = struct{}{}
		}

		var adminID int
		err := tx.Model(&Role{}).Select("role_id").First(&adminID).Error
		if err != nil {
			return err
		}
		for _, feature := range newFeatures {
			// Add new feature
			if _, exists := existingFeatureMap[feature.FeatureName]; !exists {
				if err = tx.Create(&feature).Error; err != nil {
					return err
				}

				authPermission := AuthPermission{
					PermissionID: feature.FeatureID,
					Type:         model.PERMISSION_TYP_FEATURE,
					SubjectType:  model.PERMISSION_SUB_TYP_ROLE,
					SubjectID:    int64(adminID),
				}
				if err = tx.Create(&authPermission).Error; err != nil {
					return err
				}
			}
		}

		// remove feature which not support to exist
		for featureName, featureID := range existingFeatureMap {
			if _, exists := newFeatureMap[featureName]; !exists {
				if err := tx.Where("feature_id = ?", featureID).Delete(&Feature{}).Error; err != nil {
					return err
				}
			}
		}

		// Add parent_id relationships
		parentChildMapping := map[string][]string{
			"日志检索": {"故障现场日志", "全量日志"},
			"链路追踪": {"故障现场链路", "全量链路"},
			"系统管理": {"用户管理", "菜单管理"},
		}
		for parentName, childNames := range parentChildMapping {
			var parent Feature
			if err := tx.Where("feature_name = ?", parentName).First(&parent).Error; err != nil {
				return err
			}
			for _, childName := range childNames {
				if err := tx.Model(&Feature{}).Where("feature_name = ?", childName).Update("parent_id", parent.FeatureID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
