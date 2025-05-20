// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import

// Feature is a collection of APIs, frontend routes and menu items
// that represents the embodiment of access control.
(
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
)

// FeatureMapping maps feature to menu item, router and api.
type FeatureMapping struct {
	ID         int    `gorm:"column:id;primary_key"`
	FeatureID  int    `gorm:"column:feature_id;index:feature_mapped_idx"`
	MappedID   int    `gorm:"column:mapped_id;index:feature_mapped_idx"`
	MappedType string `gorm:"column:mapped_type"` // menu router api
}

func (t *FeatureMapping) TableName() string {
	return "feature_mapping"
}

func (repo *daoRepo) GetFeature(ctx core.Context, featureIDs []int) ([]profile.Feature, error) {
	var features []profile.Feature
	query := repo.GetContextDB(ctx)
	if featureIDs != nil {
		query = query.Where("feature_id in ?", featureIDs)
	}

	err := query.Find(&features).Order("custom ASC, id ASC").Error
	return features, err
}

func (repo *daoRepo) GetFeatureMappingByFeature(ctx core.Context, featureIDs []int, mappedType string) ([]FeatureMapping, error) {
	var featureMenuItem []FeatureMapping
	err := repo.GetContextDB(ctx).Where("feature_id in ? AND mapped_type = ?", featureIDs, mappedType).Order("mapped_id").Find(&featureMenuItem).Error
	return featureMenuItem, err
}

func (repo *daoRepo) GetFeatureMappingByMapped(ctx core.Context, mappedID int, mappedType string) (FeatureMapping, error) {
	var fm FeatureMapping
	err := repo.GetContextDB(ctx).Where("mapped_id = ? AND mapped_type = ?", mappedID, mappedType).Find(&fm).Error
	return fm, err
}

func (repo *daoRepo) GetFeatureByName(ctx core.Context, name string) (int, error) {
	var id int
	err := repo.GetContextDB(ctx).Model(&profile.Feature{}).Select("feature_id").Where("feature_name = ?", name).Find(&id).Error
	return id, err
}
