package profile

type Feature struct {
	FeatureID   int    `gorm:"column:feature_id;primary_key;auto_increment" json:"featureId"`
	FeatureName string `gorm:"column:feature_name;type:varchar(20)" json:"featureName"`
	ParentID    *int   `gorm:"column:parent_id" json:"-"`
	Custom      bool   `gorm:"column:custom;default:false" json:"-"`

	Children []Feature `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
	Source   string    `gorm:"-" json:"source,omitempty"`
}

func (t *Feature) TableName() string {
	return "feature"
}
