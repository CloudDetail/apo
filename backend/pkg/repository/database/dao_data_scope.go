package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"gorm.io/gorm/clause"
)

var DaoDataScopeImpl DaoDataScope = nil

type DaoDataScope interface {
	SaveScopes(ctx core.Context, scopes []datagroup.DataScope) error
	DeleteScopes(ctx core.Context, scopesID ...string) error

	GetScopesOptionByGroupID(ctx core.Context, groupID int64) (options []string, err error)
	GetScopesSelectedByGroupID(ctx core.Context, groupID int64) (selected []string, err error)

	GetScopesByGroupID(ctx core.Context, groupID int64, category string) ([]datagroup.DataScope, error)

	SaveGroup2Scope(ctx core.Context, groupID int64, scopesID []string) error
}

func (repo *daoRepo) GetScopesSelectedByGroupID(ctx core.Context, groupID int64) (options []string, err error) {
	err = repo.GetContextDB(ctx).Model(&datagroup.DataGroup2Scope{}).
		Where("group_id = ?", groupID).
		Pluck("scope_id", &options).Error
	return options, err
}

func (repo *daoRepo) SaveScopes(ctx core.Context, scopes []datagroup.DataScope) error {
	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "scope_id"}, {Name: "category"}},
		DoNothing: true,
	}).Create(&scopes).Error
}

func (repo *daoRepo) LoadScopes(ctx core.Context) (*datagroup.DataScopeTreeNode, error) {
	var res []datagroup.DataScope
	err := repo.GetContextDB(ctx).Model(&datagroup.DataScope{}).
		Select("DISTINCT scope_id, name, type, cluster_id, namespace, service").
		Find(&res).Error
	if err != nil {
		return nil, err
	}

	// TODO init RootNode
	var root = &datagroup.DataScopeTreeNode{
		DataScope: datagroup.DataScope{
			ScopeID: "all",
			Name:    "ALL",
			Type:    datagroup.DATASOURCE_TYP_ALL,
		},
		Children: []*datagroup.DataScopeTreeNode{},
	}

	nodesMap := make(map[datagroup.ScopeLabels]*datagroup.DataScopeTreeNode)
	for i := 0; i < len(res); i++ {
		treeNode := datagroup.DataScopeTreeNode{
			DataScope: res[i],
			Children:  []*datagroup.DataScopeTreeNode{},
		}
		nodesMap[res[i].ScopeLabels] = &treeNode
	}

	for label, node := range nodesMap {
		switch node.Type {
		case datagroup.DATASOURCE_TYP_CLUSTER:
			root.Children = append(root.Children, node)
		case datagroup.DATASOURCE_TYP_NAMESPACE:
			pLabel := datagroup.ScopeLabels{
				ClusterID: label.ClusterID,
			}
			if pNode, find := nodesMap[pLabel]; find {
				pNode.Children = append(pNode.Children, node)
			}
		case datagroup.DATASOURCE_TYP_SERVICE:
			pLabel := datagroup.ScopeLabels{
				ClusterID: label.ClusterID,
				Namespace: label.Namespace,
			}
			if pNode, find := nodesMap[pLabel]; find {
				pNode.Children = append(pNode.Children, node)
			}
		}
	}
	return root, nil
}
