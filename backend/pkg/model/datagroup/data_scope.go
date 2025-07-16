// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package datagroup

import (
	"slices"
	"sort"
	"strings"
)

const (
	DATASOURCE_TYP_SYSTEM      = "system"
	DATASOURCE_TYP_CLUSTER     = "cluster"
	DATASOURCE_TYP_NAMESPACE   = "namespace"
	DATASOURCE_TYP_SERVICE     = "service"
	DATASOURCE_TYP_ENDPOINT    = "endpoint"
	DATASOURCE_TYP_INSTANCE    = "instance"
	DATASOURCE_CATEGORY_APM    = "apm"
	DATASOURCE_CATEGORY_NORMAL = "normal"
	DATASOURCE_CATEGORY_LOG    = "log"
	DATASOURCE_CATEGORY_ALERT  = "alert"
)

type DataScope struct {
	// unique id of the dataScope
	ScopeID string `gorm:"column:scope_id;primary_key;" json:"id" `
	// using when search trace/log
	Category string `gorm:"column:category;primary_key;" json:"-"`

	// display name
	Name string `gorm:"column:name;" json:"name"`
	// cluster/namespace/service
	Type string `gorm:"column:type;" json:"type,omitempty"`

	// Special Labels for this Scope
	ScopeLabels

	ClusterName string `gorm:"-" json:"clusterName"`
}

type DataScopeWithFullName struct {
	DataScope
	FullName string
}

type ExtraChild struct {
	ID   string `json:"id" gorm:"-"`
	Name string `json:"name" gorm:"-"`

	Type string `json:"type" gorm:"-"`

	ContainerID string `json:"containerId" gorm:"-"`
	POD         string `json:"pod" gorm:"-"`
	Node        string `json:"node" gorm:"-"`
	Pid         string `json:"pid" gorm:"-"`
	Endpoint    string `json:"endpoint" gorm:"-"`
	Service     string `json:"service" gorm:"-"`
}

type ScopeLabels struct {
	ClusterID string `gorm:"column:cluster_id"   json:"clusterId,omitempty" ch:"cluster_id"`
	Namespace string `gorm:"column:namespace" json:"namespace,omitempty" ch:"namespace"`
	Service   string `gorm:"column:service" json:"service,omitempty" ch:"service"`
}

// This ID is used for identification only and should not be parsed or used for business purposes.
func (l ScopeLabels) ToScopeID() string {
	if l.Namespace == "" {
		return l.ClusterID
	}
	if l.Service == "" {
		return strings.Join([]string{l.ClusterID, l.Namespace}, "#")
	}
	return strings.Join([]string{l.ClusterID, l.Namespace, l.Service}, "#")
}

func (DataScope) TableName() string {
	return "data_scope"
}

func (l DataScope) FullName() string {
	switch l.Type {
	case DATASOURCE_TYP_CLUSTER:
		return l.ClusterID
	case DATASOURCE_TYP_NAMESPACE:
		return strings.Join([]string{l.ClusterID, l.Namespace}, "-")
	case DATASOURCE_TYP_SERVICE:
		return strings.Join([]string{l.ClusterID, l.Namespace, l.Service}, "-")
	default:
		return l.ScopeID
	}
}

type DataScopeTree struct {
	*DataScopeTreeNode

	CategoryIDs map[string][]string
}

type DataScopeTreeNode struct {
	DataScope

	Children []*DataScopeTreeNode `json:"children,omitempty"`

	HasCheckBox bool `json:"hasCheckBox"`
	IsChecked   bool `json:"isChecked"`

	ExtraChildren []*ExtraChild `json:"extraChildren,omitempty"`
}

func (t *DataScopeTreeNode) RecursiveSortScope() {
	if t == nil || len(t.Children) == 0 {
		return
	}

	sort.Slice(t.Children, func(i, j int) bool {
		return t.Children[i].ScopeID < t.Children[j].ScopeID
	})

	for _, child := range t.Children {
		child.RecursiveSortScope()
	}
}

func (t *DataScopeTreeNode) CloneScopeWithPermission(options []string, selected []string) *DataScopeTreeNode {
	return t.cloneWithPermission(ignored, options, selected)
}

func (t *DataScopeTreeNode) GetScopeRef(scopeID string) *DataScopeTreeNode {
	if t.ScopeID == scopeID {
		return t
	}

	for i := 0; i < len(t.Children); i++ {
		child := t.Children[i]
		if n := child.GetScopeRef(scopeID); n != nil {
			return n
		}
	}
	return nil
}

func FillWithClusterName(scopes []DataScope, clusterNameMap map[string]string) []DataScope {
	if clusterNameMap == nil {
		return scopes
	}
	for i := 0; i < len(scopes); i++ {
		if name, find := clusterNameMap[scopes[i].ClusterID]; find {
			scopes[i].ClusterName = name
		} else if len(scopes[i].ClusterID) == 0 {
			scopes[i].ClusterName = "DEFAULT"
		} else {
			scopes[i].ClusterName = scopes[i].ClusterID
		}
	}
	return scopes
}

func (t *DataScopeTreeNode) FillWithClusterName(clusterNameMap map[string]string) {
	if clusterNameMap == nil {
		return
	}

	if name, find := clusterNameMap[t.ClusterID]; find {
		t.ClusterName = name
	} else if len(t.ClusterID) == 0 {
		t.ClusterName = "DEFAULT"
	} else {
		t.ClusterName = t.ClusterID
	}

	for i := 0; i < len(t.Children); i++ {
		t.Children[i].FillWithClusterName(clusterNameMap)
	}
}

func (t *DataScopeTreeNode) GetFullPermissionSvcList(permScopeIDs []string) []string {
	if len(permScopeIDs) == 0 {
		return []string{}
	}
	optionsMap := make(map[string]struct{})

	var dfs func(pPerm scopeStatus, node *DataScopeTreeNode)
	dfs = func(pPerm scopeStatus, node *DataScopeTreeNode) {
		if pPerm == checked || slices.Contains(permScopeIDs, node.ScopeID) {
			optionsMap[node.Service] = struct{}{}
			pPerm = checked
		}

		for i := 0; i < len(node.Children); i++ {
			child := node.Children[i]
			dfs(pPerm, child)
		}
	}

	dfs(notChecked, t)
	var result []string

	for svc := range optionsMap {
		result = append(result, svc)
	}
	return result
}

func (t *DataScopeTreeNode) GetFullPermissionScopeList(options []string) []string {
	if len(options) == 0 {
		return []string{}
	}

	optionsMap := make(map[string]bool)
	for _, id := range options {
		optionsMap[id] = true
	}

	var dfs func(pPerm scopeStatus, node *DataScopeTreeNode)
	dfs = func(pPerm scopeStatus, node *DataScopeTreeNode) {
		if pPerm == checked || slices.Contains(options, node.ScopeID) {
			optionsMap[node.ScopeID] = true
			pPerm = checked
		}

		for i := 0; i < len(node.Children); i++ {
			child := node.Children[i]
			dfs(pPerm, child)
		}
	}

	dfs(notChecked, t)
	var result []string

	for id := range optionsMap {
		result = append(result, id)
	}
	return result
}

func (t *DataScopeTreeNode) AdjustClusterName(clusterNameMap map[string]string) {
	if t.Type == DATASOURCE_TYP_CLUSTER {
		if name, find := clusterNameMap[t.ClusterID]; find {
			t.Name = name
		}
	}

	for i := 0; i < len(t.Children); i++ {
		t.Children[i].AdjustClusterName(clusterNameMap)
	}
}

func (t *DataScopeTreeNode) cloneWithPermission(pPerm scopeStatus, options []string, selected []string) *DataScopeTreeNode {
	selfStatus := checkScopePerm(pPerm, options, selected, t.ScopeID)
	if len(t.Children) == 0 {
		if selfStatus == ignored {
			return nil
		}

		return &DataScopeTreeNode{
			DataScope:   t.DataScope,
			HasCheckBox: true,
			IsChecked:   selfStatus.isChecked(),
		}
	}

	var children []*DataScopeTreeNode

	for i := 0; i < len(t.Children); i++ {
		node := t.Children[i]
		child := node.cloneWithPermission(selfStatus, options, selected)
		if child != nil {
			if selfStatus == ignored {
				selfStatus = notAllowed
			}
			children = append(children, child)
		}
	}

	if selfStatus == ignored {
		return nil
	}

	return &DataScopeTreeNode{
		DataScope:   t.DataScope,
		Children:    children,
		HasCheckBox: selfStatus.hasCheckBox(),
		IsChecked:   selfStatus.isChecked(),
	}
}

func (t *DataScopeTree) CloneWithCategory(selected []string, category string) (*DataScopeTreeNode, map[ScopeLabels]*DataScopeTreeNode) {
	if len(selected) == 0 {
		return nil, nil
	}

	var dfs func(node *DataScopeTreeNode, pStatus scopeStatus, selected []string, category string) *DataScopeTreeNode

	var leafs = make(map[ScopeLabels]*DataScopeTreeNode, 0)
	categoryIDs, find := t.CategoryIDs[category]
	if !find {
		return nil, leafs
	}

	dfs = func(node *DataScopeTreeNode, pStatus scopeStatus, selected []string, category string) *DataScopeTreeNode {
		var selfStatus scopeStatus = ignored
		if slices.Contains(selected, node.ScopeID) {
			selfStatus = checked
		} else if pStatus == checked && slices.Contains(categoryIDs, node.ScopeID) {
			selfStatus = notChecked
		}

		var subChildren []*DataScopeTreeNode
		pStatus = selfStatus
		for i := 0; i < len(node.Children); i++ {
			child := node.Children[i]
			child = dfs(child, pStatus, selected, category)
			if child != nil {
				if selfStatus != checked {
					selfStatus = notChecked
				}
				subChildren = append(subChildren, child)
			}
		}

		if selfStatus == ignored {
			return nil
		}

		newNode := &DataScopeTreeNode{
			DataScope: node.DataScope,
			Children:  subChildren,
			IsChecked: selfStatus.isChecked(),
		}

		if node.Type == DATASOURCE_TYP_SERVICE {
			leafs[node.ScopeLabels] = newNode
		}

		return newNode
	}

	root := dfs(t.DataScopeTreeNode, ignored, selected, category)
	return root, leafs
}

type scopeStatus int

const (
	checked    scopeStatus = iota // is checked
	notChecked                    // has checkbox
	notAllowed                    // no checkbox, but need to show
	ignored                       // no need to show
)

func (s scopeStatus) hasCheckBox() bool {
	switch s {
	case checked, notChecked:
		return true
	default:
		return false
	}
}

func (s scopeStatus) isChecked() bool {
	switch s {
	case checked:
		return true
	default:
		return false
	}
}

func checkScopePerm(pPerm scopeStatus, options []string, selected []string, scopeID string) scopeStatus {
	for _, id := range selected {
		if id == scopeID {
			return checked
		}
	}

	if pPerm == checked || pPerm == notChecked {
		return notChecked
	}

	for _, id := range options {
		if id == scopeID {
			return notChecked
		}
	}

	return ignored
}
