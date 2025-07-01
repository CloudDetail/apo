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

	DATA_GROUP_PERMISSION_TYPE_VIEW   = "view"
	DATA_GROUP_PERMISSION_TYPE_EDIT   = "edit"
	DATA_GROUP_PERMISSION_TYPE_IGNORE = "ignore"
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
	return "data_group_v2"
}

type DataGroupTreeNode struct {
	DataGroup

	SubGroups []DataGroupTreeNode `json:"subGroups,omitempty"`

	PermissionType string `json:"permissionType"`
}

func (t *DataGroupTreeNode) GetEditableGroups(editableGroupIDs []int64) *DataGroupTreeNode {
	return t.cloneWithPermission(DATA_GROUP_PERMISSION_TYPE_VIEW, editableGroupIDs)
}

func (t *DataGroupTreeNode) cloneWithPermission(pPerm string, groupsIDs []int64) *DataGroupTreeNode {
	selfPerm := checkPermission(pPerm, groupsIDs, t.GroupID)
	if len(t.SubGroups) == 0 {
		if selfPerm == DATA_GROUP_PERMISSION_TYPE_IGNORE {
			return nil
		}
		return &DataGroupTreeNode{
			DataGroup:      t.DataGroup,
			PermissionType: selfPerm,
			SubGroups:      []DataGroupTreeNode{},
		}
	}

	subGroups := make([]DataGroupTreeNode, 0, len(t.SubGroups))
	for _, sub := range t.SubGroups {
		if subNode := sub.cloneWithPermission(selfPerm, groupsIDs); subNode != nil {
			if selfPerm != DATA_GROUP_PERMISSION_TYPE_EDIT {
				selfPerm = DATA_GROUP_PERMISSION_TYPE_VIEW
			}
			subGroups = append(subGroups, *subNode)
		}
	}

	if selfPerm == DATA_GROUP_PERMISSION_TYPE_IGNORE {
		return nil
	}

	return &DataGroupTreeNode{
		DataGroup:      t.DataGroup,
		PermissionType: selfPerm,
		SubGroups:      subGroups,
	}
}

func checkPermission(pPerm string, groupsIDs []int64, groupID int64) string {
	if pPerm == DATA_GROUP_PERMISSION_TYPE_EDIT {
		return DATA_GROUP_PERMISSION_TYPE_EDIT
	}
	for _, id := range groupsIDs {
		if id == groupID {
			return DATA_GROUP_PERMISSION_TYPE_EDIT
		}
	}
	return DATA_GROUP_PERMISSION_TYPE_IGNORE
}
