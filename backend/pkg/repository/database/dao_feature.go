// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

// Feature is a collection of APIs, frontend routes and menu items
// that represents the embodiment of access control.
type Feature struct {
	FeatureID   int    `gorm:"column:feature_id;primary_key;auto_increment" json:"featureId"`
	FeatureName string `gorm:"column:feature_name" json:"featureName"`
	ParentID    *int   `gorm:"column:parent_id" json:"-"`
	Custom      bool   `gorm:"column:custom;default:false" json:"-"`

	Children []Feature `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
}

func (t *Feature) TableName() string {
	return "feature"
}

// FeatureAPI maps feature to api.
type FeatureAPI struct {
	FeatureID int    `gorm:"column:feature_id;primary_key"`
	APIPath   string `gorm:"column:api_path;primary_key"`
}

func (t *FeatureAPI) TableName() string {
	return "feature_api"
}

// FeatureRoute maps feature to router.
type FeatureRoute struct {
	FeatureID int    `gorm:"column:feature_id;primary_key"`
	RoutePath string `gorm:"column:route_path;primary_key"`
}

func (t *FeatureRoute) TableName() string {
	return "feature_route"
}

// FeatureMenuItem maps feature to menu item.
type FeatureMenuItem struct {
	FeatureID  int `gorm:"column:feature_id;primary_key"`
	MenuItemID int `gorm:"column:menu_item_id;primary_key"`
}

func (t *FeatureMenuItem) TableName() string {
	return "feature_menu_item"
}

func (repo *daoRepo) GetFeature(featureIDs []int) ([]Feature, error) {
	var features []Feature
	query := repo.db
	if featureIDs != nil {
		query = query.Where("feature_id in ?", featureIDs)
	}

	err := query.Find(&features).Order("custom ASC, id ASC").Error
	return features, err
}

func (repo *daoRepo) GetMappedMenuItem(featureIDs []int) ([]FeatureMenuItem, error) {
	var featureMenuItem []FeatureMenuItem
	err := repo.db.Where("feature_id in ?", featureIDs).Order("menu_item_id").Find(&featureMenuItem).Error
	return featureMenuItem, err
}
