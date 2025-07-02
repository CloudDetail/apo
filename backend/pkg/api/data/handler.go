// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"go.uber.org/zap"
)

type Handler interface {
	// GetDatasource Gets all datasource.
	// @Tags API.data
	// @Router /api/data/datasource [get]
	GetDatasource() core.HandlerFunc

	// CreateDataGroup Create a data group.
	// @Tags API.data
	// @Router /api/data/group/create [post]
	// CreateDataGroup() core.HandlerFunc

	// DeleteDataGroup Delete the data group.
	// @Tags API.data
	// @Router /api/data/group/delete [post]
	DeleteDataGroup() core.HandlerFunc

	// GetDataGroup Get data group.
	// @Tags API.data
	// @Router /api/data/group [post]
	GetDataGroup() core.HandlerFunc

	// GetGroupDatasource Get group's datasource.
	// @Tags API.data
	// @Router /api/data/group/data [get]
	GetGroupDatasource() core.HandlerFunc

	// DataGroupOperation Assign data groups to users or teams, or remove them from data groups.
	// @Tags API.data
	// @Router /api/data/group/operation [post]
	DataGroupOperation() core.HandlerFunc

	// GetSubjectDataGroup Get subject's assigned data group.
	// @Tags API.data
	// @Router /api/data/sub/group [get]
	GetSubjectDataGroup() core.HandlerFunc

	// GetUserDataGroup Get user's assigned data group.
	// @Tags API.data
	// @Router /api/data/user/group [get]
	GetUserDataGroup() core.HandlerFunc

	// GroupSubsOperation Manage group's assigned subject.
	// @Tags API.data
	// @Router /api/data/subs/operation [post]
	GroupSubsOperation() core.HandlerFunc

	// GetGroupSubs Get group's assigned subjects.
	// @Tags API.data
	// @Router /api/data/subs [get]
	GetGroupSubs() core.HandlerFunc

	// ################## V2 API #######################

	// @Router /api/v2/data/group
	GetDataGroupV2() core.HandlerFunc

	GetDGDetailV2() core.HandlerFunc

	// // @Router /api/v2/data/group/datasource/list
	GetDGScopeList() core.HandlerFunc

	CreateDataGroupV2() core.HandlerFunc

	UpdateDataGroupV2() core.HandlerFunc

	// DeleteDGroup() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	dataService data.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, promRepo prometheus.Repo, chRepo clickhouse.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:      logger,
		dataService: data.New(dbRepo, promRepo, chRepo, k8sRepo),
	}
}
