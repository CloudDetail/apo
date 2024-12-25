package database

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

func (repo *daoRepo) initPermissions() error {
	roleFeatures := map[string][]string{
		"admin": {
			"服务概览", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "配置中心",
			"系统管理", "用户管理", "菜单管理",
		},
		"manager": {
			"服务概览", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "配置中心",
		},
		"viewer": {
			"服务概览", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "配置中心",
		},
		"anonymous": {
			"服务概览", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "配置中心",
		},
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&AuthPermission{}); err != nil {
			return err
		}

		var featureIDs []int
		if err := tx.Model(&Feature{}).Select("feature_id").Find(&featureIDs).Error; err != nil {
			return err
		}

		// revoke feature permission whose feature has already been deleted
		err := tx.Model(&AuthPermission{}).
			Where("permission_id NOT IN ? AND type = ?", featureIDs, model.PERMISSION_TYP_FEATURE).
			Delete(nil).Error
		if err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&AuthPermission{}).Count(&count).Error; err != nil {
			return err
		}

		// Initialised only on first boot
		if count > 0 {
			return nil
		}
		for roleName, featureNames := range roleFeatures {
			var role Role
			if err := tx.Where("role_name = ?", roleName).First(&role).Error; err != nil {
				return err
			}

			var features []Feature
			if err := tx.Where("feature_name IN ?", featureNames).Find(&features).Error; err != nil {
				return err
			}

			for _, feature := range features {
				permission := AuthPermission{
					SubjectID:    int64(role.RoleID),
					SubjectType:  "role",
					Type:         "feature",
					PermissionID: feature.FeatureID,
				}
				if err := tx.Create(&permission).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
