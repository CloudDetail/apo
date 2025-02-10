// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/mitchellh/mapstructure"
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
	resp := make(map[string]any)

	chCfg := config.Get().ClickHouse
	resp["database"] = integration.LogIntegration{
		Name:   "APO-DEFAULT-CH",
		DBType: "clickhouse",
		Mode:   "sql",
		LogAPI: &integration.JSONField[integration.LogAPI]{
			Obj: integration.LogAPI{
				Clickhouse: &integration.ClickhouseConfig{
					Address:     chCfg.Address,
					Database:    chCfg.Database,
					Replication: chCfg.Replica,
					Cluster:     chCfg.Cluster,
				},
			},
		},
	}

	vmCfg := config.Get().Promethues
	resp["datasource"] = integration.MetricIntegration{
		Name:   "APO-DEFAULT-VM",
		DSType: "victoriametric",
		Mode:   "pql",
		MetricAPI: &integration.JSONField[integration.MetricAPI]{
			Obj: integration.MetricAPI{
				VictoriaMetric: &integration.VictoriaMetricConfig{
					ServerURL: vmCfg.Address,
				},
			},
		},
		UpdatedAt: time.Time{},
		IsDeleted: false,
	}

	latestTraceAPI, err := s.dbRepo.GetLatestTraceAPIs(-1)
	if err == nil {
		var traceAPI integration.JSONField[integration.TraceAPI]
		err := mapstructure.Decode(latestTraceAPI.APIs, &traceAPI.Obj)
		if err == nil {
			traceAPI.ReplaceSecret()
			traceAPI.Obj.Timeout = latestTraceAPI.Timeout
			resp["traceAPI"] = traceAPI
		}
	}
	return resp
}
