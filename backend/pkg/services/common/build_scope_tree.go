package common

import "github.com/CloudDetail/apo/backend/pkg/model/datagroup"

func convertScopesToScopeTree(scopes []datagroup.DataScope) *datagroup.DataScopeTreeNode {
	root := &datagroup.DataScopeTreeNode{
		DataScope: datagroup.DataScope{
			Type: datagroup.DATASOURCE_TYP_SYSTEM,
			Name: "ALL",
		},
		IsChecked: false,
		Children:  []*datagroup.DataScopeTreeNode{},
	}

	nodesMap := make(map[datagroup.ScopeLabels]*datagroup.DataScopeTreeNode)

	for _, scope := range scopes {
		node := &datagroup.DataScopeTreeNode{
			DataScope: scope,
			Children:  []*datagroup.DataScopeTreeNode{},
			IsChecked: true,
		}
		nodesMap[scope.ScopeLabels] = node
		if scope.Type == datagroup.DATASOURCE_TYP_SYSTEM {
			root = node
		}
	}

	for label, node := range nodesMap {
		switch node.Type {
		case datagroup.DATASOURCE_TYP_CLUSTER:
			root.Children = append(root.Children, node)
		case datagroup.DATASOURCE_TYP_NAMESPACE:
			parentLabel := datagroup.ScopeLabels{ClusterID: label.ClusterID}
			parent := getOrCreateParent(nodesMap, label.ClusterID, datagroup.DATASOURCE_TYP_CLUSTER, parentLabel)
			parent.Children = append(parent.Children, node)
		case datagroup.DATASOURCE_TYP_SERVICE:
			parentLabel := datagroup.ScopeLabels{
				ClusterID: label.ClusterID,
				Namespace: label.Namespace,
			}
			parent := getOrCreateParent(nodesMap, label.Namespace, datagroup.DATASOURCE_TYP_NAMESPACE, parentLabel)
			parent.Children = append(parent.Children, node)
		}
	}
	return root
}

func getOrCreateParent(
	nodesMap map[datagroup.ScopeLabels]*datagroup.DataScopeTreeNode,
	name string, typ string, label datagroup.ScopeLabels,
) *datagroup.DataScopeTreeNode {
	if parent, exists := nodesMap[label]; exists {
		return parent
	}

	parent := &datagroup.DataScopeTreeNode{
		DataScope: datagroup.DataScope{
			ScopeID:     label.ScopeID(),
			Name:        name,
			Type:        typ,
			ScopeLabels: label,
		},
		Children:  []*datagroup.DataScopeTreeNode{},
		IsChecked: false,
	}
	nodesMap[label] = parent

	if typ == datagroup.DATASOURCE_TYP_NAMESPACE {
		grandLabel := datagroup.ScopeLabels{ClusterID: label.ClusterID}
		grand := getOrCreateParent(nodesMap, label.ClusterID, datagroup.DATASOURCE_TYP_CLUSTER, grandLabel)
		grand.Children = append(grand.Children, parent)
	}

	return parent
}
