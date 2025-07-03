package datagroup

type DataGroup2Scope struct {
	GroupID int64  `gorm:"group_id;primary_key"`
	ScopeID string `gorm:"scope_id;primary_key"`
}

func (DataGroup2Scope) TableName() string {
	return "data_group_2_scope"
}
