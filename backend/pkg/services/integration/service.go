// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	GetDatasourceAndDatabase() map[string]any

	CreateCluster(cluster *integration.ClusterIntegrationVO) error
	GetClusterIntegration(clusterID string) (*integration.ClusterIntegrationVO, error)
	UpdateClusterIntegration(cluster *integration.ClusterIntegrationVO) error

	ListCluster() ([]integration.Cluster, error)
	DeleteCluster(cluster *integration.Cluster) error

	GetIntegrationInstallConfigFile(req *integration.GetCInstallRequest) (*integration.GetCInstallConfigResponse, error)
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

// HACK use static config in configFile now
func (s *service) GetDatasourceAndDatabase() map[string]any {
	chCfg := config.Get().ClickHouse

	db := integration.LogIntegration{
		Name:   "APO-DEFAULT-CH",
		DBType: "clickhouse",
		Mode:   "sql",
		LogAPI: &integration.JSONField[integration.LogAPI]{
			Obj: integration.LogAPI{
				CHConfig: &integration.ClickhouseConfig{
					Address:     chCfg.Address,
					Database:    chCfg.Database,
					Replication: chCfg.Replica,
					Cluster:     chCfg.Cluster,
				},
			},
		},
	}

	vmCfg := config.Get().Promethues
	ds := integration.MetricIntegration{
		Name:   "APO-DEFAULT-VM",
		DSType: "victoriametric",
		Mode:   "pql",
		MetricAPI: &integration.JSONField[integration.MetricAPI]{
			Obj: integration.MetricAPI{
				VMConfig: &integration.VictoriaMetricConfig{
					ServerURL: vmCfg.Address,
				},
			},
		},
		UpdatedAt: time.Time{},
		IsDeleted: false,
	}

	return map[string]any{
		"datasource": ds,
		"database":   db,
	}
}
