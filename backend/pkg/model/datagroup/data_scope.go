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
	Category string `gorm:"column:category;primary_key;" json:"-"`

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

type DataScopeTreeNode struct {
	DataScope

	Children []DataScopeTreeNode `json:"children,omitempty"`

	HasCheckBox bool `json:"hasCheckBox,omitempty"`
	IsChecked   bool `json:"isChecked,omitempty"`
}

func (t *DataScopeTreeNode) GetEditableScopeTree(options []string, selected []string) *DataScopeTreeNode {
	return t.cloneWithCheckStatus(ignored, options, selected)
}

func (t *DataScopeTreeNode) cloneWithCheckStatus(pCheckStatus scopeStatus, options []string, selected []string) *DataScopeTreeNode {
	selfStatus := checkScope(pCheckStatus, options, selected, t.ScopeID)
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

	var children []DataScopeTreeNode
	for _, node := range t.Children {
		child := node.cloneWithCheckStatus(selfStatus, options, selected)
		if child != nil {
			if selfStatus == ignored {
				selfStatus = notAllowed
			}
			children = append(children, *child)
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

func checkScope(pStatus scopeStatus, options []string, selected []string, scopeID string) scopeStatus {
	for _, id := range selected {
		if id == scopeID {
			return checked
		}
	}

	if pStatus == checked || pStatus == notChecked {
		return notChecked
	}

	for _, id := range options {
		if id == scopeID {
			return notChecked
		}
	}

	return ignored
}
