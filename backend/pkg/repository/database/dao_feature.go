package database

type Feature struct {
	FeatureID   int    `gorm:"column:feature_id;primary_key;auto_increment" json:"featureId"`
	FeatureName string `gorm:"column:feature_name" json:"featureName"`
	ParentID    *int   `gorm:"column:parent_id" json:"-"`

	Children []Feature `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
}

func (t *Feature) TableName() string {
	return "feature"
}

type FeatureAPI struct {
	FeatureID int    `gorm:"column:feature_id;primary_key"`
	APIPath   string `gorm:"column:api_path;primary_key"`
}

func (t *FeatureAPI) TableName() string {
	return "feature_api"
}

type FeatureRoute struct {
	FeatureID int    `gorm:"column:feature_id;primary_key"`
	RoutePath string `gorm:"column:route_path;primary_key"`
}

func (t *FeatureRoute) TableName() string {
	return "feature_route"
}

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

	err := query.Find(&features).Error
	return features, err
}

func (repo *daoRepo) GetMappedMenuItem(featureIDs []int) ([]FeatureMenuItem, error) {
	var featureMenuItem []FeatureMenuItem
	err := repo.db.Where("feature_id in ?", featureIDs).Find(&featureMenuItem).Error
	return featureMenuItem, err
}
