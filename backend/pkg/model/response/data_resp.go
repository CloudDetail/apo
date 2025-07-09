// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type GetDatasourceResponse struct {
	NamespaceList []model.Datasource `json:"namespaceList"`
	ServiceList   []model.Datasource `json:"serviceList"`
}

type GetDataGroupResponse struct {
	DataGroupList    []datagroup.DataGroup `json:"dataGroupList"`
	model.Pagination `json:",inline"`
}

type GetGroupDatasourceResponse struct {
	GroupDatasource

	ClusterScopedDatasource []ClusterScopedDatasource `json:"clusterScoped"`
}

type ClusterScopedDatasource struct {
	ClusterID   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`

	GroupDatasource
}

type GroupDatasource struct {
	NamespaceMap map[string][]string `json:"namespaceMap"` // namespace: services
	ServiceMap   map[string][]string `json:"serviceMap"`   // service: endpoints
}

type GetSubjectDataGroupResponse []datagroup.DataGroup

type GetGroupSubsResponse []database.AuthDataGroup

type ListDataScopesResponse struct {
	Scopes      *datagroup.DataScopeTreeNode `json:"view"`
	DataSources []string                     `json:"datasources"`
}

type ListDataScopeFilterResponse struct {
	Scopes *datagroup.DataScopeTreeNode `json:"view"`
}

type CleanExpiredDataScopeResponse struct {
	ToBeDeleted []datagroup.DataScopeWithFullName `json:"toBeDeleted"`
	Protected   []datagroup.DataScopeWithFullName `json:"protected"`
}

type SubGroupDetailResponse struct {
	Datasources []datagroup.DataScope           `json:"datasources"`
	SubGroups   []datagroup.DataGroupWithScopes `json:"subGroups"`
}
