// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
)

func (s *service) CreateCluster(cluster *integration.ClusterIntegration) (*integration.Cluster, error) {
	isExist, err := s.dbRepo.CheckClusterNameExisted(cluster.Name)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, errors.New("cluster name already exists")
	}

	cluster.ID = uuid.NewString()
	err = s.dbRepo.CreateCluster(&cluster.Cluster)
	if err != nil {
		return nil, err
	}

	// HACK 当前强制指定VM和CK配置
	forceSetupMetricLogAPI(cluster)

	err = s.dbRepo.SaveIntegrationConfig(*cluster)
	if err != nil {
		return nil, err
	}

	return &cluster.Cluster, nil
}

func forceSetupMetricLogAPI(cluster *integration.ClusterIntegration) {
	vmCfg := config.Get().Promethues
	cluster.Metric.DSType = "self-collector"
	cluster.Metric.Name = "APO-DEFAULT-VM"
	cluster.Metric.Mode = "pql"
	cluster.Metric.MetricAPI = &integration.JSONField[integration.MetricAPI]{
		Obj: integration.MetricAPI{
			VictoriaMetric: &integration.VictoriaMetricConfig{
				ServerURL: vmCfg.Address,
			},
		},
	}

	chCfg := config.Get().ClickHouse
	cluster.Log.DBType = "self-collector"
	cluster.Log.Name = "APO-DEFAULT-CK"
	cluster.Log.Mode = "sql"
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
}
