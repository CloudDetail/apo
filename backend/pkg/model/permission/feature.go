package permission

// Feature is a collection of APIs, frontend routes and menu items
// that represents the embodiment of access control.
type Feature struct {
	FeatureID   int    `gorm:"column:feature_id;primary_key;auto_increment" json:"featureId"`
	FeatureName string `gorm:"column:feature_name;type:varchar(20)" json:"featureName"`
	ParentID    *int   `gorm:"column:parent_id" json:"-"`
	Custom      bool   `gorm:"column:custom;default:false" json:"-"`

	Children []Feature `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
	Source   string    `gorm:"-" json:"source,omitempty"`
}
