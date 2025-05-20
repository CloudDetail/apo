// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"gorm.io/gorm"
)

func isValidRoleName(roleName string) bool {
	for _, role := range validRoles {
		if roleName == role.RoleName {
			return true
		}
	}

	return false
}

func (repo *daoRepo) initPermissions(ctx core.Context) error {
	roleFeatures := map[string][]string{
		model.ROLE_ADMIN: {
			"服务概览", "工作流", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "告警通知", "配置中心", "数据接入", "告警接入", "告警事件", "告警事件详情",
			"系统管理", "用户管理", "菜单管理", "系统配置", "数据组管理", "团队管理", "角色管理",
		},
		// model.ROLE_MANAGER: {
		// 	"服务概览", "工作流", "日志检索", "故障现场日志", "全量日志", "链路追踪",
		// 	"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
		// 	"应用指标大盘", "中间件大盘", "告警规则", "告警通知", "数据接入", "告警接入", "告警事件","告警事件详情",
		// 	"配置中心",
		// },
		model.ROLE_VIEWER: {
			"服务概览", "工作流", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "告警通知", "配置中心", "告警事件", "告警事件详情",
		},
		model.ROLE_ANONYMOS: {
			"服务概览", "工作流", "日志检索", "故障现场日志", "全量日志", "链路追踪",
			"故障现场链路", "全量链路", "全局资源大盘", "应用基础设施大盘",
			"应用指标大盘", "中间件大盘", "告警规则", "告警通知", "配置中心", "告警事件", "告警事件详情",
		},
	}

	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		var featureIDs []int
		if err := tx.Model(&profile.Feature{}).Select("feature_id").Find(&featureIDs).Error; err != nil {
			return err
		}

		// revoke feature permission whose feature has already been deleted
		err := tx.Model(&AuthPermission{}).
			Where(`permission_id NOT IN ? AND "type" = ?`, featureIDs, model.PERMISSION_TYP_FEATURE).
			Delete(nil).Error
		if err != nil {
			return err
		}

		for roleName, featureNames := range roleFeatures {
			if !isValidRoleName(roleName) {
				continue
			}

			var role profile.Role
			if err = tx.Where("role_name = ?", roleName).First(&role).Error; err != nil {
				return err
			}

			if role.RoleID <= 0 {
				continue
			}

			var count int64
			err = tx.Model(&AuthPermission{}).
				Where(`subject_id = ? AND subject_type = ? AND "type" = ?`, role.RoleID, model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE).
				Count(&count).
				Error
			if err != nil {
				return err
			}

			// Initialised only if the role has not been assigned permissions
			if count > 0 {
				continue
			}

			validFeatureNames := make([]string, 0, len(featureNames))
			for _, feature := range featureNames {
				if isValidFeature(feature) {
					validFeatureNames = append(validFeatureNames, feature)
				}
			}

			var features []profile.Feature
			if err = tx.Where("feature_name IN ?", validFeatureNames).Find(&features).Error; err != nil {
				return err
			}

			for _, feature := range features {
				if feature.FeatureID <= 0 {
					continue
				}

				permission := AuthPermission{
					SubjectID:    int64(role.RoleID),
					SubjectType:  "role",
					Type:         "feature",
					PermissionID: feature.FeatureID,
				}
				if err = tx.Create(&permission).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
