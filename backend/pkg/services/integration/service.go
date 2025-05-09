// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	GetStaticIntegration() map[string]any

	CreateCluster(cluster *integration.ClusterIntegration) (*integration.Cluster, error)
	GetClusterIntegration(clusterID string) (*integration.ClusterIntegrationVO, error)
	UpdateClusterIntegration(cluster *integration.ClusterIntegration) error

	ListCluster() ([]integration.Cluster, error)
	DeleteCluster(cluster *integration.Cluster) error

	GetIntegrationInstallConfigFile(req *integration.GetCInstallRequest) (*integration.GetCInstallConfigResponse, error)
	// Deprecated
	GetIntegrationInstallDoc(req *integration.GetCInstallRequest) ([]byte, error)

	TriggerAdapterUpdate(req *integration.TriggerAdapterUpdateRequest)
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
