// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

type Service interface {
	GetDataSource(ctx core.Context) (response.GetDatasourceResponse, error)
	GetDataGroup(ctx core.Context, req *request.GetDataGroupRequest) (response.GetDataGroupResponse, error)
	GetGroupDatasource(ctx core.Context, req *request.GetGroupDatasourceRequest) (response.GetGroupDatasourceResponse, error)
	DataGroupOperation(ctx core.Context, req *request.DataGroupOperationRequest) error
	GetSubjectDataGroup(ctx core.Context, req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error)

	CheckGroupPermission(ctx core.Context, groupID int64) (bool, error)
	CheckScopePermission(ctx core.Context, cluster, namespace, service string) (bool, error)
	CheckServicesPermission(ctx core.Context, services ...string) (bool, error)

	GroupSubsOperation(ctx core.Context, req *request.GroupSubsOperationRequest) error
	GetGroupSubs(ctx core.Context, req *request.GetGroupSubsRequest) (response.GetGroupSubsResponse, error)

	// TestStoreScope(ctx core.Context)

	ListDataGroupV2(ctx core.Context) (*datagroup.DataGroupTreeNode, error)

	ListDataScopeByGroupID(ctx core.Context, req *request.DGScopeListRequest) (*response.ListDataScopesResponse, error)
	GetGroupDetailWithSubGroup(ctx core.Context, groupID int64) (*response.SubGroupDetailResponse, error)

	CreateDataGroupV2(ctx core.Context, req *request.CreateDataGroupRequest) error
	UpdateDataGroupV2(ctx core.Context, req *request.UpdateDataGroupRequest) error
	DeleteDataGroupV2(ctx core.Context, req *request.DeleteDataGroupRequest) error

	GetFilterByGroupID(ctx core.Context, req *request.DGFilterRequest) (*response.ListDataScopeFilterResponse, error)

	CleanExpiredDataScope(ctx core.Context, groupID int64, clean bool) (*response.CleanExpiredDataScopeResponse, error)
}

type service struct {
	dbRepo   database.Repo
	promRepo prometheus.Repo
	chRepo   clickhouse.Repo

	k8sRepo kubernetes.Repo
}

func New(dbRepo database.Repo, promRepo prometheus.Repo, chRepo clickhouse.Repo, k8sRepo kubernetes.Repo) Service {
	common.InitDataGroupStorage(promRepo, chRepo, dbRepo)
	service := &service{
		dbRepo:   dbRepo,
		promRepo: promRepo,
		k8sRepo:  k8sRepo,
		chRepo:   chRepo,
	}
	return service
}
