// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) UpdateClusterIntegration(ctx core.Context, cluster *integration.ClusterIntegration) error {
	// TODO 当前强制指定VM和CK配置
	vmCfg := config.Get().Promethues
	cluster.Metric.MetricAPI = &integration.JSONField[integration.MetricAPI]{
		Obj: integration.MetricAPI{
			VictoriaMetric: &integration.VictoriaMetricConfig{
				ServerURL: vmCfg.Address,
			},
		},
	}

	chCfg := config.Get().ClickHouse
	cluster.Log.LogAPI = &integration.JSONField[integration.LogAPI]{
		Obj: integration.LogAPI{
			Clickhouse: &integration.ClickhouseConfig{
				Address:     chCfg.Address,
				Database:    chCfg.Database,
				Replication: chCfg.Replica,
				Cluster:     chCfg.Cluster,
				UserName:    chCfg.Username,
				Password:    chCfg.Password,
			},
		},
	}

	cluster.APOCollector.RemoveHttpPrefix()
	err := s.dbRepo.UpdateCluster(&cluster.Cluster)
	if err != nil {
		return err
	}

	return s.dbRepo.SaveIntegrationConfig(*cluster)
}
