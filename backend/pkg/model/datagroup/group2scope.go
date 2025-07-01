package datagroup

type DataGroup2Scope struct {
	GroupID int64  `gorm:"group_id"`
	ScopeID string `gorm:"scope_id"`
}

func (DataGroup2Scope) TableName() string {
	return "data_group_2_scope"
}
