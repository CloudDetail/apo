// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type Service interface {
	GetStaticIntegration(ctx_core core.Context,) map[string]any

	CreateCluster(ctx_core core.Context, cluster *integration.ClusterIntegration) (*integration.Cluster, error)
	GetClusterIntegration(ctx_core core.Context, clusterID string) (*integration.ClusterIntegrationVO, error)
	UpdateClusterIntegration(ctx_core core.Context, cluster *integration.ClusterIntegration) error

	ListCluster(ctx_core core.Context,) ([]integration.Cluster, error)
	DeleteCluster(ctx_core core.Context, cluster *integration.Cluster) error

	GetIntegrationInstallConfigFile(ctx_core core.Context, req *integration.GetCInstallRequest) (*integration.GetCInstallConfigResponse, error)
	// Deprecated
	GetIntegrationInstallDoc(ctx_core core.Context, req *integration.GetCInstallRequest) ([]byte, error)

	TriggerAdapterUpdate(ctx_core core.Context, req *integration.TriggerAdapterUpdateRequest)
}

var _ Service = &service{}

type service struct {
	dbRepo database.Repo
}

func New(database database.Repo) Service {
	return &service{
		dbRepo: database,
	}
}
