// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MenuItem is a menu item on the left or top menu bar.
//
// ! Handle the `Key` field carefully, it is a reserved word in MySQL, and must be delimited with the identifier '"'

type MenuItem struct {
	ItemID       int    `gorm:"column:item_id;primary_key" json:"itemId"`
	Key          string `gorm:"column:key;type:varchar(64);uniqueIndex" json:"key"`
	Label        string `gorm:"-" json:"label"` // AKA item name.
	Icon         string `gorm:"column:icon;type:varchar(150)" json:"icon"`
	ParentID     *int   `gorm:"column:parent_id" json:"-"`
	Abbreviation string `gorm:"-" json:"abbreviation,omitempty"`
	RouterID     int    `gorm:"column:router_id" json:"-"`
	Order        int    `gorm:"column:sort_order;index:sort_order_idx" json:"-"` // The order of a menu item.

	Children []MenuItem `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
	Router   *Router    `gorm:"-" json:"router,omitempty"` // Frontend router.
}

func (t *MenuItem) TableName() string {
	return "menu_item"
}

func (repo *daoRepo) GetMenuItems(ctx core.Context) ([]MenuItem, error) {
	var menuItems []MenuItem

	err := repo.GetContextDB(ctx).Find(&menuItems).Error
	return menuItems, err
}

func (repo *daoRepo) UpsertCustomDashboardMenu(ctx core.Context, id string, name string, visibility string) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		return repo.upsertCustomDashboardMenuTx(ctx, tx, id, name, visibility)
	})
}

func (repo *daoRepo) syncCustomDashboardMenus(ctx core.Context) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		var dashboards []CustomDashboard
		if err := tx.Find(&dashboards).Error; err != nil {
			return err
		}
		for _, dashboard := range dashboards {
			if err := repo.upsertCustomDashboardMenuTx(ctx, tx, dashboard.ID, dashboard.Name, dashboard.Visibility); err != nil {
				return err
			}
			if dashboard.OwnerID <= 0 {
				continue
			}
			featureID, err := repo.getFeatureIDTx(tx, customDashboardFeatureName(dashboard.ID))
			if err != nil {
				return err
			}
			if err := upsertUserFeaturePermissionTx(tx, dashboard.OwnerID, featureID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *daoRepo) upsertCustomDashboardMenuTx(ctx core.Context, tx *gorm.DB, id string, name string, visibility string) error {
	key := customDashboardMenuKey(id)
	route := customDashboardRoute(id)
	label := strings.TrimSpace(name)
	if label == "" {
		label = "Custom Dashboard"
	}

	var parent MenuItem
	if err := tx.Where(`"key" = ?`, "customDashboard").First(&parent).Error; err != nil {
		return err
	}

	router := Router{RouterTo: route, Custom: true}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "router_to"}},
		UpdateAll: true,
	}).Create(&router).Error; err != nil {
		return err
	}
	if err := tx.Where("router_to = ?", route).First(&router).Error; err != nil {
		return err
	}

	order := parent.Order + 1
	item := MenuItem{Key: key, Icon: "customDashboard", ParentID: &parent.ItemID, RouterID: router.RouterID, Order: order}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"parent_id":  item.ParentID,
			"router_id":  item.RouterID,
			"sort_order": item.Order,
		}),
	}).Create(&item).Error; err != nil {
		return err
	}
	if err := tx.Where(`"key" = ?`, key).First(&item).Error; err != nil {
		return err
	}

	parentFeatureID, err := repo.getFeatureIDTx(tx, "我的大盘")
	if err != nil {
		return err
	}
	featureID, err := repo.upsertCustomDashboardFeatureTx(tx, id, label, parentFeatureID)
	if err != nil {
		return err
	}
	if err := upsertFeatureMappingTx(tx, featureID, item.ItemID, model.MAPPED_TYP_MENU); err != nil {
		return err
	}
	if err := upsertFeatureMappingTx(tx, featureID, router.RouterID, model.MAPPED_TYP_ROUTER); err != nil {
		return err
	}
	if ctx != nil && ctx.UserID() > 0 {
		if err := upsertUserFeaturePermissionTx(tx, ctx.UserID(), featureID); err != nil {
			return err
		}
	}
	return repo.upsertCustomDashboardMenuTranslationsTx(tx, item.ItemID, label)
}

func (repo *daoRepo) DeleteCustomDashboardMenu(ctx core.Context, id string) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		return repo.deleteCustomDashboardMenuTx(tx, id)
	})
}

func (repo *daoRepo) deleteCustomDashboardMenuTx(tx *gorm.DB, id string) error {
	key := customDashboardMenuKey(id)
	route := customDashboardRoute(id)
	var item MenuItem
	if err := tx.Where(`"key" = ?`, key).First(&item).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if item.ItemID > 0 {
		if err := tx.Where("mapped_id = ? AND mapped_type = ?", item.ItemID, model.MAPPED_TYP_MENU).Delete(&FeatureMapping{}).Error; err != nil {
			return err
		}
		if err := tx.Where("entity_id = ? AND entity_type = ?", item.ItemID, model.TRANSLATION_TYP_MENU).Delete(&I18nTranslation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("item_id = ?", item.ItemID).Delete(&MenuItem{}).Error; err != nil {
			return err
		}
	}
	var router Router
	if err := tx.Where("router_to = ?", route).First(&router).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if router.RouterID > 0 {
		if err := tx.Where("mapped_id = ? AND mapped_type = ?", router.RouterID, model.MAPPED_TYP_ROUTER).Delete(&FeatureMapping{}).Error; err != nil {
			return err
		}
		if err := tx.Where("router_id = ? AND custom = ?", router.RouterID, true).Delete(&Router{}).Error; err != nil {
			return err
		}
	}
	featureName := customDashboardFeatureName(id)
	var feature profile.Feature
	if err := tx.Where("feature_name = ? AND custom = ?", featureName, true).First(&feature).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if feature.FeatureID > 0 {
		if err := tx.Where("feature_id = ?", feature.FeatureID).Delete(&FeatureMapping{}).Error; err != nil {
			return err
		}
		if err := tx.Where("permission_id = ? AND type = ?", feature.FeatureID, model.PERMISSION_TYP_FEATURE).Delete(&AuthPermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("entity_id = ? AND entity_type = ?", feature.FeatureID, model.TRANSLATION_TYP_FEATURE).Delete(&I18nTranslation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("feature_id = ? AND custom = ?", feature.FeatureID, true).Delete(&profile.Feature{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func customDashboardMenuKey(id string) string {
	key := strings.ReplaceAll(id, "-", "")
	if len(key) > 17 {
		key = key[:17]
	}
	return "cd-" + key
}

func customDashboardFeatureName(id string) string {
	return customDashboardMenuKey(id)
}

func customDashboardRoute(id string) string {
	return "/custom-dashboard/" + id
}

func (repo *daoRepo) getFeatureIDTx(tx *gorm.DB, name string) (int, error) {
	var id int
	err := tx.Table("feature").Select("feature_id").Where("feature_name = ?", name).Row().Scan(&id)
	return id, err
}

func upsertFeatureMappingTx(tx *gorm.DB, featureID int, mappedID int, mappedType string) error {
	var count int64
	if err := tx.Model(&FeatureMapping{}).
		Where("feature_id = ? AND mapped_id = ? AND mapped_type = ?", featureID, mappedID, mappedType).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return tx.Create(&FeatureMapping{FeatureID: featureID, MappedID: mappedID, MappedType: mappedType}).Error
}

func (repo *daoRepo) upsertCustomDashboardFeatureTx(tx *gorm.DB, id string, label string, parentFeatureID int) (int, error) {
	featureName := customDashboardFeatureName(id)
	var feature profile.Feature
	if err := tx.Where("feature_name = ? AND custom = ?", featureName, true).First(&feature).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
		feature = profile.Feature{FeatureName: featureName, ParentID: &parentFeatureID, Custom: true}
		if err := tx.Create(&feature).Error; err != nil {
			return 0, err
		}
	} else if feature.ParentID == nil || *feature.ParentID != parentFeatureID {
		if err := tx.Model(&profile.Feature{}).Where("feature_id = ?", feature.FeatureID).Update("parent_id", parentFeatureID).Error; err != nil {
			return 0, err
		}
	}
	if err := upsertCustomDashboardFeatureTranslationsTx(tx, feature.FeatureID, label); err != nil {
		return 0, err
	}
	return feature.FeatureID, nil
}

func upsertUserFeaturePermissionTx(tx *gorm.DB, userID int64, featureID int) error {
	var count int64
	if err := tx.Model(&AuthPermission{}).
		Where("subject_id = ? AND subject_type = ? AND type = ? AND permission_id = ?", userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE, featureID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return tx.Create(&AuthPermission{
		SubjectID:    userID,
		SubjectType:  model.PERMISSION_SUB_TYP_USER,
		Type:         model.PERMISSION_TYP_FEATURE,
		PermissionID: featureID,
	}).Error
}

func upsertCustomDashboardFeatureTranslationsTx(tx *gorm.DB, featureID int, label string) error {
	if err := tx.Where("entity_id = ? AND entity_type = ?", featureID, model.TRANSLATION_TYP_FEATURE).Delete(&I18nTranslation{}).Error; err != nil {
		return err
	}
	translations := []I18nTranslation{
		{Language: "zh", Translation: label, FieldName: "feature_name", EntityID: featureID, EntityType: model.TRANSLATION_TYP_FEATURE},
		{Language: "en", Translation: label, FieldName: "feature_name", EntityID: featureID, EntityType: model.TRANSLATION_TYP_FEATURE},
	}
	for _, translation := range translations {
		if err := tx.Create(&translation).Error; err != nil {
			return err
		}
	}
	return nil
}

func (repo *daoRepo) upsertCustomDashboardMenuTranslationsTx(tx *gorm.DB, itemID int, label string) error {
	if err := tx.Where("entity_id = ? AND entity_type = ?", itemID, model.TRANSLATION_TYP_MENU).Delete(&I18nTranslation{}).Error; err != nil {
		return err
	}
	translations := []I18nTranslation{
		{Language: "zh", Translation: label, FieldName: "label", EntityID: itemID, EntityType: model.TRANSLATION_TYP_MENU},
		{Language: "zh", Translation: label, FieldName: "abbreviation", EntityID: itemID, EntityType: model.TRANSLATION_TYP_MENU},
		{Language: "en", Translation: label, FieldName: "label", EntityID: itemID, EntityType: model.TRANSLATION_TYP_MENU},
		{Language: "en", Translation: label, FieldName: "abbreviation", EntityID: itemID, EntityType: model.TRANSLATION_TYP_MENU},
	}
	for _, translation := range translations {
		if err := tx.Create(&translation).Error; err != nil {
			return err
		}
	}
	return nil
}
