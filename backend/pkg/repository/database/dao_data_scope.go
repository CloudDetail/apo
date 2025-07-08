package database

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DaoDataScopeImpl DaoDataScope = nil

type DaoDataScope interface {
	SaveScopes(ctx core.Context, scopes []datagroup.DataScope) error
	LoadScopes(ctx core.Context) (*datagroup.DataScopeTree, error)
	DeleteScopes(ctx core.Context, scopesIDs []string) error

	GetScopesByScopeIDs(ctx core.Context, scopeIDs []string) ([]datagroup.DataScope, error)
	GetScopeIDsOptionByGroupID(ctx core.Context, groupID int64) (options []string, err error)
	GetScopeIDsSelectedByGroupID(ctx core.Context, groupID int64) (selected []string, err error)
	GetScopeIDsSelectedByPermGroupIDs(ctx core.Context, permGroupIDs []int64) ([]string, error)

	UpdateGroup2Scope(ctx core.Context, groupID int64, scopeIDs []string) error
	DeleteGroup2Scope(ctx core.Context, groupID int64) error
	DeleteGroup2ScopeByGroupIDScopeIDs(ctx core.Context, groupID int64, scopeIDs []string) error

	GetScopeIDsByGroupIDAndCat(ctx core.Context, groupID int64, category string) ([]string, error)
	GetScopesByGroupIDAndCat(ctx core.Context, groupID int64, category string) ([]datagroup.DataScope, error)

	CheckScopePermission(ctx core.Context, permGroupIDs []int64, cluster, namespace, service string) (bool, error)

	CheckScopesPermission(ctx core.Context, permGroupIDs []int64, scopeIDs []string) (perm []string, err error)
}

func (repo *daoRepo) GetScopesByScopeIDs(ctx core.Context, scopeIDs []string) ([]datagroup.DataScope, error) {
	var res []datagroup.DataScope
	return res, repo.GetContextDB(ctx).Where("scope_id in ?", scopeIDs).Find(&res).Error
}

func (repo *daoRepo) CheckScopesPermission(ctx core.Context, permGroupIDs []int64, scopeIDs []string) (perm []string, err error) {
	err = repo.GetContextDB(ctx).Table("data_group_2_scope as dgs").
		Where("dgs.group_id in ?", permGroupIDs).
		Where("dgs.scope_id in ?", scopeIDs).
		Pluck("dgs.scope_id", &perm).Error

	return
}

func (repo *daoRepo) CheckScopePermission(ctx core.Context, permGroupIDs []int64, cluster, namespace, service string) (bool, error) {
	var res datagroup.DataScope

	db := repo.GetContextDB(ctx)
	err := db.Table("data_group_2_scope as dgs").
		Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id").
		Where("dgs.group_id in ?", permGroupIDs).
		Where(
			db.Where("ds.cluster_id = ? and ds.namespace = ? and ds.service = ?", cluster, namespace, service).
				Or("ds.cluster_id = ? and ds.namespace = ? and ds.type = 'namespace'", cluster, namespace).
				Or("ds.cluster_id = ? and ds.type = 'cluster'", cluster),
		).First(&res).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
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

	if inputs == nil {
		return nil
	}

	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "scope_id"}},
		DoNothing: true,
	}).Create(&inputs).Error
}

func (repo *daoRepo) DeleteGroup2Scope(ctx core.Context, groupID int64) error {
	return repo.GetContextDB(ctx).Where("group_id = ?", groupID).Delete(&datagroup.DataGroup2Scope{}).Error
}

func (repo *daoRepo) DeleteGroup2ScopeByGroupIDScopeIDs(ctx core.Context, groupID int64, scopeIDs []string) error {
	return repo.GetContextDB(ctx).
		Where("group_id = ?", groupID).
		Where("scope_id in ?", scopeIDs).
		Delete(&datagroup.DataGroup2Scope{}).Error
}

func (repo *daoRepo) GetScopeIDsOptionByGroupID(ctx core.Context, groupID int64) (options []string, err error) {
	err = repo.GetContextDB(ctx).Table("data_group as dg").
		Joins("RIGHT JOIN data_group_2_scope dgs ON dg.parent_group_id = dgs.group_id").
		Where("dg.group_id = ?", groupID).
		Distinct("dgs.scope_id").
		Pluck("dgs.scope_id", &options).Error
	return options, err
}

func (repo *daoRepo) GetScopeIDsSelectedByGroupID(ctx core.Context, groupID int64) (options []string, err error) {
	err = repo.GetContextDB(ctx).Model(&datagroup.DataGroup2Scope{}).
		Where("group_id = ?", groupID).
		Pluck("scope_id", &options).Error
	return options, err
}

func (repo *daoRepo) GetScopeIDsSelectedByPermGroupIDs(ctx core.Context, permGroupIDs []int64) (selected []string, err error) {
	err = repo.GetContextDB(ctx).Model(&datagroup.DataGroup2Scope{}).
		Where("group_id in ?", permGroupIDs).
		Pluck("scope_id", &selected).Error
	return selected, err
}

func (repo *daoRepo) SaveScopes(ctx core.Context, scopes []datagroup.DataScope) error {
	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "scope_id"}, {Name: "category"}},
		DoNothing: true,
	}).Create(&scopes).Error
}

func (repo *daoRepo) LoadScopes(ctx core.Context) (*datagroup.DataScopeTree, error) {
	var res []datagroup.DataScope
	err := repo.GetContextDB(ctx).Model(&datagroup.DataScope{}).
		Find(&res).Error
	if err != nil {
		return nil, err
	}

	// TODO init RootNode
	var root *datagroup.DataScopeTreeNode
	var categoryMaps = make(map[string][]string)

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
		categoryMaps[res[i].Category] = append(categoryMaps[res[i].Category], res[i].ScopeID)
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
	return &datagroup.DataScopeTree{DataScopeTreeNode: root, CategoryIDs: categoryMaps}, nil
}

func (repo *daoRepo) DeleteScopes(ctx core.Context, scopesIDs []string) error {
	return repo.GetContextDB(ctx).Where("scope_id in ?", scopesIDs).Delete(&datagroup.DataScope{}).Error
}

func (repo *daoRepo) GetScopesByGroupIDAndCat(ctx core.Context, groupID int64, category string) ([]datagroup.DataScope, error) {
	var res []datagroup.DataScope
	qb := repo.GetContextDB(ctx).Table("data_group_2_scope as dgs").
		Where("group_id = ?", groupID).
		Select("DISTINCT dgs.scope_id, name, type, cluster_id, namespace, service")

	if len(category) > 0 {
		qb.Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id AND (ds.category = ? or ds.category = 'system')", category)
	} else {
		qb.Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id")
	}

	err := qb.Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *daoRepo) GetScopeIDsByGroupIDAndCat(ctx core.Context, groupID int64, category string) ([]string, error) {
	if groupID == 0 {
		// ALL Group
		var res []string
		qb := repo.GetContextDB(ctx).Table("data_scope").
			Distinct("scope_id").
			Pluck("scope_id", &res)

		if len(category) > 0 {
			qb.Where("category = ? or category = 'system'", category)
		}

		err := qb.Find(&res).Error
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	var res []string
	qb := repo.GetContextDB(ctx).Table("data_group_2_scope as dgs").
		Where("group_id = ? ", groupID).
		Distinct("dgs.scope_id").
		Pluck("dgs.scope_id", &res)

	if len(category) > 0 {
		qb.Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id AND (ds.category = ? or ds.category = 'system')", category)
	} else {
		qb.Joins("INNER JOIN data_scope ds ON ds.scope_id = dgs.scope_id")
	}

	err := qb.Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}
