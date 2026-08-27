// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) CreateCluster(ctx core.Context, cluster *integration.ClusterIntegration) (*integration.Cluster, error) {
	if err := normalizeClusterIdentity(&cluster.Cluster); err != nil {
		return nil, err
	}

	isExist, err := s.dbRepo.CheckClusterNameExisted(ctx, cluster.Name)
	if err != nil {
		return nil, core.Error(code.CreateClusterFailed, fmt.Sprintf("failed to check display name %q: %v", cluster.Name, err))
	}
	if isExist {
		return nil, errors.New("cluster name already exists")
	}

	cluster.APOCollector.RemoveHttpPrefix()
	err = s.dbRepo.CreateCluster(ctx, &cluster.Cluster)
	if err != nil {
		return nil, err
	}

	// HACK 当前强制指定VM和CK配置
	forceSetupMetricLogAPI(cluster)

	err = s.dbRepo.SaveIntegrationConfig(ctx, *cluster)
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
