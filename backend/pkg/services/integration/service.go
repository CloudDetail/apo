// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"regexp"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	GetStaticIntegration(ctx core.Context) map[string]any

	CreateCluster(ctx core.Context, cluster *integration.ClusterIntegration) (*integration.Cluster, error)
	GetClusterIntegration(ctx core.Context, clusterID string) (*integration.ClusterIntegrationVO, error)
	UpdateClusterIntegration(ctx core.Context, cluster *integration.ClusterIntegration) error

	GetBaseURL() string
	GetServerAddr() string

	ListCluster(ctx core.Context) ([]integration.Cluster, error)
	DeleteCluster(ctx core.Context, cluster *integration.Cluster) error

	GetIntegrationInstallConfigFile(ctx core.Context, req *integration.GetCInstallRequest) (*integration.GetCInstallConfigResponse, error)
	// Deprecated
	GetIntegrationInstallDoc(ctx core.Context, req *integration.GetCInstallRequest) ([]byte, error)

	TriggerAdapterUpdate(ctx core.Context, req *integration.TriggerAdapterUpdateRequest)
}

var _ Service = &service{}

type service struct {
	dbRepo database.Repo

	serverAddr string
	baseURL    string
}

var baseSchema = regexp.MustCompile(`^(?:https?:\/\/)?([^:/]+)`)

func New(database database.Repo) Service {

	cfg := config.Get().Server

	service := &service{
		dbRepo: database,
	}

	if len(cfg.BaseURL) == 0 {
		service.baseURL = "http://apo-backend-svc.apo:8080"
		service.serverAddr = "apo-backend-svc.apo"
	} else {
		match := baseSchema.FindStringSubmatch(cfg.BaseURL)
		if len(match) > 1 {
			service.serverAddr = match[1]
		} else {
			service.serverAddr = cfg.BaseURL
		}
		service.baseURL = cfg.BaseURL
	}

	return service
}
