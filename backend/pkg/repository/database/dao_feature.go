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

func (repo *daoRepo) GetFeature(featureIDs []int) ([]Feature, error) {
	var features []Feature
	query := repo.db
	if featureIDs != nil {
		query = query.Where("feature_id in ?", featureIDs)
	}

	err := query.Find(&features).Order("custom ASC, id ASC").Error
	return features, err
}

func (repo *daoRepo) GetFeatureMapping(featureIDs []int, mappedType string) ([]FeatureMapping, error) {
	var featureMenuItem []FeatureMapping
	err := repo.db.Where("feature_id in ? AND mapped_type = ?", featureIDs, mappedType).Order("mapped_id").Find(&featureMenuItem).Error
	return featureMenuItem, err
}
