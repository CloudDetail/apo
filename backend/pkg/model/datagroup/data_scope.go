package datagroup

const (
	DATASOURCE_TYP_NAMESPACE   = "namespace"
	DATASOURCE_TYP_SERVICE     = "service"
	DATASOURCE_CATEGORY_APM    = "apm"
	DATASOURCE_CATEGORY_NORMAL = "normal"
	DATASOURCE_CATEGORY_LOG    = "log"
	DATASOURCE_CATEGORY_ALERT  = "alert"
)

type DataScope struct {
	// unique id of the dataScope
	ScopeID string `gorm:"column:scope_id;primary_key;" json:"id" `
	// using when search trace/log
	Category string `gorm:"column:category;primary_key;" json:"category"`

	// display name
	Name string `gorm:"column:name;" json:"name"`
	// cluster/namespace/service
	Type string `gorm:"column:type;" json:"type,omitempty"`

	// Special Labels for this Scope
	ClusterID string `gorm:"column:clusterId;" json:"clusterId,omitempty"`
	Namespace string `gorm:"column:namespace;" json:"namespace,omitempty"`
	Service   string `gorm:"column:service;" json:"service,omitempty"`
}

func (DataScope) TableName() string {
	return "data_scope"
}

type DataGroup2Scope struct {
	GroupID int64  `gorm:"group_id"`
	ScopeID string `gorm:"scope_id"`
}

func (DataGroup2Scope) TableName() string {
	return "data_group_2_scope"
}
