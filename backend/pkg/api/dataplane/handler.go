// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/dataplane"
	"go.uber.org/zap"
)

type Handler interface {
	// QueryServices Get all services.
	// @Tags API.dataplane
	// @Router /api/dataplane/services [get]
	QueryServices() core.HandlerFunc
	// QueryServiceRedCharts Get service's redcharts.
	// @Tags API.dataplane
	// @Router /api/dataplane/redcharts [get]
	QueryServiceRedCharts() core.HandlerFunc
	// QueryServiceEndpoints Get service's endpoints.
	// @Tags API.dataplane
	// @Router /api/dataplane/endpoints [get]
	QueryServiceEndpoints() core.HandlerFunc
	// QueryServiceInstances Get service's instances.
	// @Tags API.dataplane
	// @Router /api/dataplane/instances [get]
	QueryServiceInstances() core.HandlerFunc
	// QueryServiceName Get service name by instance.
	// @Tags API.dataplane
	// @Router /api/dataplane/servicename [post]
	QueryServiceName() core.HandlerFunc
	// QueryTopology Get service's topology.
	// @Tags API.dataplane
	// @Router /api/dataplane/topology [get]
	QueryTopology() core.HandlerFunc
	// CreateCustomTopology Create custom topology.
	// @Tags API.dataplane
	// @Router /api/dataplane/customtopology/create [post]
	CreateCustomTopology() core.HandlerFunc
	// ListCustomTopology List custom topology.
	// @Tags API.dataplane
	// @Router /api/dataplane/customtopology/list [get]
	ListCustomTopology() core.HandlerFunc
	// DeleteCustomTopology Delete custom topology.
	// @Tags API.dataplane
	// @Router /api/dataplane/customtopology/delete [post]
	DeleteCustomTopology() core.HandlerFunc
	// CheckServiceNameRule Check servicename rule.
	// @Tags API.dataplane
	// @Router /api/dataplane/servicename/checkRule [post]
	CheckServiceNameRule() core.HandlerFunc
	// SetServiceNameRule Create or update servicename rule.
	// @Tags API.dataplane
	// @Router /api/dataplane/servicename/upsertRule [post]
	SetServiceNameRule() core.HandlerFunc
	// ListServiceNameRule List servicename rule.
	// @Tags API.dataplane
	// @Router /api/dataplane/servicename/listRule [get]
	ListServiceNameRule() core.HandlerFunc
}

type handler struct {
	logger           *zap.Logger
	dataplaneService dataplane.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, promRepo prometheus.Repo, dbRepo database.Repo) Handler {
	return &handler{
		logger:           logger,
		dataplaneService: dataplane.New(chRepo, promRepo, dbRepo),
	}
}
