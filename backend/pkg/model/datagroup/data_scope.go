package datagroup

import "strings"

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

func (t *DataScopeTreeNode) CloneScopeWithPermission(options []string, selected []string) *DataScopeTreeNode {
	return t.cloneWithPermission(ignored, options, selected)
}

func (t *DataScopeTreeNode) GetScopeRef(scopeID string) *DataScopeTreeNode {
	if t.ScopeID == scopeID {
		return t
	}

	for _, child := range t.Children {
		if n := child.GetScopeRef(scopeID); n != nil {
			return n
		}
	}
	return nil
}

func (t *DataScopeTreeNode) GetFullPermissionScopeList(options []string) []string {
	optionsMap := make(map[string]bool)
	for _, id := range options {
		optionsMap[id] = true
	}

	var dfs func(pPerm scopeStatus, node *DataScopeTreeNode)
	dfs = func(pPerm scopeStatus, node *DataScopeTreeNode) {
		if pPerm == checked || containsInStr(options, node.ScopeID) {
			optionsMap[node.ScopeID] = true
			for _, child := range node.Children {
				dfs(checked, child)
			}
			return
		}
		for _, child := range node.Children {
			dfs(notChecked, child)
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
	for _, node := range t.Children {
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
	var dfs func(node *DataScopeTreeNode, pStatus scopeStatus, selected []string, category string) *DataScopeTreeNode

	var leafs = make(map[ScopeLabels]*DataScopeTreeNode, 0)
	categoryIDs, find := t.CategoryIDs[category]
	if !find {
		return nil, leafs
	}

	dfs = func(node *DataScopeTreeNode, pStatus scopeStatus, selected []string, category string) *DataScopeTreeNode {
		var selfStatus scopeStatus = ignored
		if containsInStr(selected, node.ScopeID) {
			selfStatus = checked
		} else if pStatus == checked && containsInStr(categoryIDs, node.ScopeID) {
			selfStatus = notChecked
		}

		var subChildren []*DataScopeTreeNode
		pStatus = selfStatus
		for _, child := range node.Children {
			child := dfs(child, pStatus, selected, category)
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

func containsInStr(options []string, input string) bool {
	for _, v := range options {
		if v == input {
			return true
		}
	}
	return false
}
