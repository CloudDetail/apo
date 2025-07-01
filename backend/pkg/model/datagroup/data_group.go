package datagroup

var RootDataGroup = DataGroup{
	ParentGroupID: -1,
	GroupID:       0,
	GroupName:     "ALL",
	Description:   "Contains all data",
}

const (
	DATA_GROUP_SUB_TYP_USER   = "user"
	DATA_GROUP_SUB_TYP_TEAM   = "team"
	DATA_GROUP_SOURCE_DEFAULT = "default"
)

type DataGroup struct {
	GroupID     int64  `gorm:"column:group_id;primary_key;auto_increment" json:"groupId"`
	GroupName   string `gorm:"column:group_name;type:varchar(20)" json:"groupName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"` // The description of data group.

	AuthType string `json:"authType,omitempty"`
	Source   string `gorm:"-" json:"source,omitempty"`

	ParentGroupID int64 `gorm:"column:parent_group_id" json:"parentGroupId"`
}

func (DataGroup) TableName() string {
	return "data_group"
}
