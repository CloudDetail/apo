package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"gorm.io/gorm/clause"
)

var DaoDataScopeImpl DaoDataScope = nil

type DaoDataScope interface {
	SaveScopes(ctx core.Context, scopes []datagroup.DataScope) error
	LoadScopes(ctx core.Context) (*datagroup.DataScopeTreeNode, error)
	DeleteScopes(ctx core.Context, scopesID ...string) error

	GetScopesOptionByGroupID(ctx core.Context, groupID int64) (options []string, err error)
	GetScopesSelectedByGroupID(ctx core.Context, groupID int64) (selected []string, err error)

	UpdateGroup2Scope(ctx core.Context, groupID int64, scopeIDs []string) error
	DeleteGroup2Scope(ctx core.Context, groupID int64) error

	GetScopesByGroupID(ctx core.Context, groupID int64, category string) ([]datagroup.DataScope, error)
}

func (repo *daoRepo) UpdateGroup2Scope(ctx core.Context, groupID int64, scopeIDs []string) error {
	err := repo.GetContextDB(ctx).Where("group_id = ?", groupID).Delete(&datagroup.DataGroup2Scope{}).Error
	if err != nil {
		return err
	}

	var inputs []datagroup.DataGroup2Scope
	for _, scopeID := range scopeIDs {
		inputs = append(inputs, datagroup.DataGroup2Scope{
			GroupID: groupID,
			ScopeID: scopeID,
		})
	}

	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "scope_id"}},
		DoNothing: true,
	}).Create(&inputs).Error
}

func (repo *daoRepo) DeleteGroup2Scope(ctx core.Context, groupID int64) error {
	return repo.GetContextDB(ctx).Where("group_id = ?", groupID).Delete(&datagroup.DataGroup2Scope{}).Error
}

func (repo *daoRepo) GetScopesOptionByGroupID(ctx core.Context, groupID int64) (options []string, err error) {
	err = repo.GetContextDB(ctx).Table("data_group as dg").
		Joins("RIGHT JOIN data_group_2_scope dgs ON dg.parent_group_id = dgs.group_id").
		Where("dg.group_id = ?", groupID).
		Distinct("dgs.scope_id").
		Pluck("dgs.scope_id", &options).Error
	return options, err
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
	var root *datagroup.DataScopeTreeNode

	nodesMap := make(map[datagroup.ScopeLabels]*datagroup.DataScopeTreeNode)
	for i := 0; i < len(res); i++ {
		treeNode := datagroup.DataScopeTreeNode{
			DataScope: res[i],
			Children:  []*datagroup.DataScopeTreeNode{},
		}
		if res[i].Type == datagroup.DATASOURCE_TYP_SYSTEM {
			root = &treeNode
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

func (repo *daoRepo) DeleteScopes(ctx core.Context, scopesID ...string) error {
	// TODO
	panic("not implemented")
	return repo.GetContextDB(ctx).Model(&datagroup.DataScope{}).Where("scope_id in ?", scopesID).Delete(nil).Error
}

func (repo *daoRepo) GetScopesByGroupID(ctx core.Context, groupID int64, category string) ([]datagroup.DataScope, error) {
	var res []datagroup.DataScope
	qb := repo.GetContextDB(ctx).Table("data_group_2_scope as dgs").
		Where("group_id = ?", groupID).
		Select("DISTINCT dgs.scope_id, name, type, cluster_id, namespace, service")

	if len(category) > 0 {
		qb.Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id AND ds.category = ?", category)
	} else {
		qb.Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id")
	}

	err := qb.Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}
