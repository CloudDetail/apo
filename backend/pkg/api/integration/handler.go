// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/integration"
)

type Handler interface {
	// ListCluster ListCluster
	// @Tags API.integration
	// @Router /api/integration/cluster/list [get]
	ListCluster() core.HandlerFunc

	// CreateCluster CreateCluster
	// @Tags API.integration
	// @Router /api/integration/cluster/create [post]
	CreateCluster() core.HandlerFunc

	// UpdateCluster UpdateCluster
	// @Tags API.integration
	// @Router /api/integration/cluster/update [post]
	UpdateCluster() core.HandlerFunc

	// DeleteCluster DeleteCluster
	// @Tags API.integration
	// @Router /api/integration/cluster/delete [post]
	DeleteCluster() core.HandlerFunc

	// GetIntegrationInstallDoc
	// @Tags API.integration
	// @Router /api/integration/cluster/install/cmd [get]
	GetIntegrationInstallDoc() core.HandlerFunc

	// GetIntegrationInstallConfigFile
	// @Tags API.integration
	// @Router /api/integration/cluster/install/config [get]
	GetIntegrationInstallConfigFile() core.HandlerFunc
}

var _ Handler = &handler{}

type handler struct {
	integrationService integration.Service
}

func New(database database.Repo) Handler {
	return &handler{
		integrationService: integration.New(database),
	}
}
