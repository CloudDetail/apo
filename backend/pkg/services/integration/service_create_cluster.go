// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
)

func (s *service) CreateCluster(cluster *integration.ClusterIntegrationVO) error {
	isExist, err := s.dbRepo.CheckClusterNameExisted(cluster.Name)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("cluster name already exists")
	}

	cluster.ID = uuid.NewString()
	err = s.dbRepo.CreateCluster(&cluster.Cluster)
	if err != nil {
		return err
	}

	// HACK 当前强制指定VM和CK配置
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

	return s.dbRepo.SaveIntegrationConfig(integration.ClusterIntegration{
		ClusterID:   cluster.ID,
		ClusterName: cluster.Name,
		ClusterType: cluster.ClusterType,
		Trace:       cluster.Trace,
		Metric:      cluster.Metric,
		Log:         cluster.Log,
	})
}
