// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type Service interface {
	GetDataSource() (response.GetDatasourceResponse, error)
	CreateDataGroup(req *request.CreateDataGroupRequest) error
	DeleteDataGroup(req *request.DeleteDataGroupRequest) error
	GetDataGroup(req *request.GetDataGroupRequest) (response.GetDataGroupResponse, error)
	UpdateDataGroup(req *request.UpdateDataGroupRequest) error
	GetGroupDatasource(req *request.GetGroupDatasourceRequest, userID int64) (response.GetGroupDatasourceResponse, error)
	DataGroupOperation(req *request.DataGroupOperationRequest) error
	GetSubjectDataGroup(req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error)
	// CheckDatasourcePermission Filtering and filling data sources that users are not authorised to view. Expected *string or *[]string.
	CheckDatasourcePermission(userID, groupID int64, namespaces, services interface{}, fillCategory string) (err error)
	GroupSubsOperation(req *request.GroupSubsOperationRequest) error
	GetGroupSubs(req *request.GetGroupSubsRequest) (response.GetGroupSubsResponse, error)
}

type service struct {
	dbRepo   database.Repo
	promRepo prometheus.Repo
	k8sRepo  kubernetes.Repo
}

func New(dbRepo database.Repo, promRepo prometheus.Repo, k8sRepo kubernetes.Repo) Service {
	return &service{
		dbRepo:   dbRepo,
		promRepo: promRepo,
		k8sRepo:  k8sRepo,
	}
}
