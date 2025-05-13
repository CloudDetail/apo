// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type Service interface {
	GetDataSource(ctx_core core.Context,) (response.GetDatasourceResponse, error)
	CreateDataGroup(ctx_core core.Context, req *request.CreateDataGroupRequest) error
	DeleteDataGroup(ctx_core core.Context, req *request.DeleteDataGroupRequest) error
	GetDataGroup(ctx_core core.Context, req *request.GetDataGroupRequest) (response.GetDataGroupResponse, error)
	UpdateDataGroup(ctx_core core.Context, req *request.UpdateDataGroupRequest) error
	GetGroupDatasource(ctx_core core.Context, req *request.GetGroupDatasourceRequest, userID int64) (response.GetGroupDatasourceResponse, error)
	DataGroupOperation(ctx_core core.Context, req *request.DataGroupOperationRequest) error
	GetSubjectDataGroup(ctx_core core.Context, req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error)
	// CheckDatasourcePermission Filtering and filling data sources that users are not authorised to view. Expected *string or *[]string.
	CheckDatasourcePermission(ctx_core core.Context, userID, groupID int64, namespaces, services interface{}, fillCategory string) (err error)
	GroupSubsOperation(ctx_core core.Context, req *request.GroupSubsOperationRequest) error
	GetGroupSubs(ctx_core core.Context, req *request.GetGroupSubsRequest) (response.GetGroupSubsResponse, error)
}

type service struct {
	dbRepo		database.Repo
	promRepo	prometheus.Repo
	k8sRepo		kubernetes.Repo
}

func New(dbRepo database.Repo, promRepo prometheus.Repo, k8sRepo kubernetes.Repo) Service {
	return &service{
		dbRepo:		dbRepo,
		promRepo:	promRepo,
		k8sRepo:	k8sRepo,
	}
}
