// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package datagroup

import "github.com/CloudDetail/apo/backend/pkg/model"

const (
	DATA_GROUP_SUB_TYP_USER   = "user"
	DATA_GROUP_SUB_TYP_TEAM   = "team"
	DATA_GROUP_SOURCE_DEFAULT = "default"

	DATA_GROUP_PERMISSION_TYPE_KNOWN  = "known"
	DATA_GROUP_PERMISSION_TYPE_VIEW   = "view"
	DATA_GROUP_PERMISSION_TYPE_EDIT   = "edit"
	DATA_GROUP_PERMISSION_TYPE_IGNORE = "ignore"
)

type DataGroup struct {
	GroupID     int64  `gorm:"column:group_id;primary_key;auto_increment" json:"groupId"`
	GroupName   string `gorm:"column:group_name;type:varchar(110)" json:"groupName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"` // The description of data group.

	AuthType string `json:"authType,omitempty"`
	Source   string `gorm:"-" json:"source,omitempty"`

	ParentGroupID int64 `gorm:"column:parent_group_id" json:"parentGroupId"`
}

type DataGroupWithScopes struct {
	DataGroup

	Scopes         []DataScope `json:"datasources"`
	PermissionType string      `json:"permissionType"`
}

func (DataGroup) TableName() string {
	return "data_group"
}

type DataGroupTreeNode struct {
	DataGroup

	SubGroups      []*DataGroupTreeNode `json:"subGroups,omitempty"`
	PermissionType string               `json:"permissionType"`
}

func (t *DataGroupTreeNode) CloneWithPermission(permGroupIDs []int64) *DataGroupTreeNode {
	return t.cloneWithPermission(DATA_GROUP_PERMISSION_TYPE_KNOWN, permGroupIDs)
}

func (t *DataGroupTreeNode) CheckGroupPermission(groupID int64, permGroupIDs []int64) bool {
	if containsInInt(permGroupIDs, groupID) {
		return true
	}

	// return
	// -1 no permission
	// 0 not found
	// 1 have permission
	var dfs func(node *DataGroupTreeNode, pPerm string) int
	dfs = func(node *DataGroupTreeNode, pPerm string) int {
		if node.GroupID == groupID {
			if pPerm == DATA_GROUP_PERMISSION_TYPE_VIEW {
				return 1
			}
			return -1
		}

		if containsInInt(permGroupIDs, node.GroupID) {
			pPerm = DATA_GROUP_PERMISSION_TYPE_VIEW
		}

		for i := 0; i < len(node.SubGroups); i++ {
			sub := node.SubGroups[i]
			res := dfs(sub, pPerm)
			if res == 1 || res == -1 {
				return res
			}
		}
		return 0
	}

	return dfs(t, DATA_GROUP_PERMISSION_TYPE_KNOWN) == 1
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
			SubGroups:      []*DataGroupTreeNode{},
		}
	}

	subGroups := make([]*DataGroupTreeNode, 0, len(t.SubGroups))
	for i := 0; i < len(t.SubGroups); i++ {
		sub := t.SubGroups[i]
		if subNode := sub.cloneWithPermission(selfPerm, groupsIDs); subNode != nil {
			if selfPerm == DATA_GROUP_PERMISSION_TYPE_IGNORE {
				selfPerm = DATA_GROUP_PERMISSION_TYPE_KNOWN
			}
			subGroups = append(subGroups, subNode)
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

func (t *DataGroupTreeNode) CloneGroupNodeWithSubGroup(groupID int64, groupsIDs []int64) *DataGroupTreeNode {
	return t.cloneGroupNodeWithSubGroup(groupID, DATA_GROUP_PERMISSION_TYPE_KNOWN, groupsIDs)
}

func (t *DataGroupTreeNode) cloneGroupNodeWithSubGroup(groupID int64, pPerm string, groupsIDs []int64) *DataGroupTreeNode {
	selfPerm := checkPermission(pPerm, groupsIDs, t.GroupID)
	if t.GroupID == groupID {
		if selfPerm == DATA_GROUP_PERMISSION_TYPE_IGNORE {
			return nil
		}
		subGroups := make([]*DataGroupTreeNode, 0, len(t.SubGroups))
		for i := 0; i < len(t.SubGroups); i++ {
			sub := subGroups[i]
			perm := checkPermission(selfPerm, groupsIDs, sub.GroupID)
			if perm == DATA_GROUP_PERMISSION_TYPE_IGNORE {
				continue
			}
			subGroups = append(subGroups, &DataGroupTreeNode{
				DataGroup:      sub.DataGroup,
				SubGroups:      nil, // Drop sub groups
				PermissionType: perm,
			})
		}

		return &DataGroupTreeNode{
			DataGroup:      t.DataGroup,
			PermissionType: selfPerm,
			SubGroups:      subGroups,
		}
	}

	for i := 0; i < len(t.SubGroups); i++ {
		sub := t.SubGroups[i]
		if subNode := sub.cloneGroupNodeWithSubGroup(groupID, selfPerm, groupsIDs); subNode != nil {
			return subNode
		}

	}
	return nil
}

func (t *DataGroupTreeNode) GetGroupNodeRef(groupID int64) *DataGroupTreeNode {
	if t.GroupID == groupID {
		return t
	}

	for i := 0; i < len(t.SubGroups); i++ {
		sub := t.SubGroups[i]
		if subNode := sub.GetGroupNodeRef(groupID); subNode != nil {
			return subNode
		}
	}
	return nil
}

func (t *DataGroupTreeNode) GetFullSubGroupIDs(groupID int64) []int64 {
	var subGroupIDs []int64
	var dfs func(isSubGroup bool, node *DataGroupTreeNode)
	dfs = func(pNode bool, node *DataGroupTreeNode) {
		if pNode || node.GroupID == groupID {
			subGroupIDs = append(subGroupIDs, node.GroupID)
			pNode = true
		}
		for i := 0; i < len(node.SubGroups); i++ {
			sub := node.SubGroups[i]
			dfs(pNode, sub)
		}
	}
	dfs(false, t)
	return subGroupIDs
}

func (t *DataGroupTreeNode) GetFullPermissionGroupWithSource(groupIDs []int64, fromUser, fromTeam []int64) []DataGroup {
	var permGroups []DataGroup

	var dfs func(pPerm string, pSource string, node *DataGroupTreeNode)
	dfs = func(pPerm string, pSource string, node *DataGroupTreeNode) {
		if pPerm == DATA_GROUP_PERMISSION_TYPE_VIEW ||
			containsInInt(groupIDs, node.GroupID) {
			group := node.DataGroup
			if containsInInt(fromUser, node.GroupID) {
				group.Source = model.DATA_GROUP_SUB_TYP_USER
				pSource = model.DATA_GROUP_SUB_TYP_USER
			} else if containsInInt(fromTeam, node.GroupID) {
				group.Source = model.DATA_GROUP_SUB_TYP_TEAM
				pSource = model.DATA_GROUP_SUB_TYP_TEAM
			} else {
				group.Source = pSource
			}
			permGroups = append(permGroups, group)
			pPerm = DATA_GROUP_PERMISSION_TYPE_VIEW
		}
		for i := 0; i < len(node.SubGroups); i++ {
			child := node.SubGroups[i]
			dfs(pPerm, pSource, child)
		}
	}

	dfs(DATA_GROUP_PERMISSION_TYPE_KNOWN, DATA_GROUP_SOURCE_DEFAULT, t)
	return permGroups
}

func (t *DataGroupTreeNode) GetFullPermissionGroup(groupIDs []int64) []DataGroup {
	var permGroups []DataGroup

	var dfs func(pPerm string, node *DataGroupTreeNode)
	dfs = func(pPerm string, node *DataGroupTreeNode) {
		if pPerm == DATA_GROUP_PERMISSION_TYPE_VIEW ||
			containsInInt(groupIDs, node.GroupID) {
			permGroups = append(permGroups, node.DataGroup)
			pPerm = DATA_GROUP_PERMISSION_TYPE_VIEW
		}

		for i := 0; i < len(node.SubGroups); i++ {
			child := node.SubGroups[i]
			dfs(pPerm, child)
		}
	}

	dfs(DATA_GROUP_PERMISSION_TYPE_KNOWN, t)
	return permGroups
}

func checkPermission(pPerm string, groupsIDs []int64, groupID int64) string {
	if pPerm == DATA_GROUP_PERMISSION_TYPE_EDIT || pPerm == DATA_GROUP_PERMISSION_TYPE_VIEW {
		return DATA_GROUP_PERMISSION_TYPE_EDIT
	}
	for _, id := range groupsIDs {
		if id == groupID {
			return DATA_GROUP_PERMISSION_TYPE_VIEW
		}
	}
	return DATA_GROUP_PERMISSION_TYPE_IGNORE
}

func containsInInt(options []int64, input int64) bool {
	for _, v := range options {
		if v == input {
			return true
		}
	}
	return false
}
