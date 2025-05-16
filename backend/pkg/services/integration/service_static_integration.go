// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/mitchellh/mapstructure"
)

// HACK use static config in configFile now
func (s *service) GetStaticIntegration(ctx core.Context) map[string]any {
	resp := make(map[string]any)

	chCfg := config.Get().ClickHouse
	log := integration.LogIntegration{
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
					UserName:    chCfg.Username,
					Password:    chCfg.Password,
				},
			},
		},
	}
	log.LogAPI.ReplaceSecret()
	resp["database"] = log

	vmCfg := config.Get().Promethues
	ds := integration.MetricIntegration{
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
		UpdatedAt: 0,
		IsDeleted: false,
	}
	ds.MetricAPI.ReplaceSecret()
	resp["datasource"] = ds

	if latestTraceAPI, err := s.dbRepo.GetLatestTraceAPIs(-1); err == nil {
		if latestTraceAPI == nil {
			resp["traceAPI"] = integration.TraceAPI{}
		} else {
			var traceAPI integration.JSONField[integration.TraceAPI]
			err := mapstructure.Decode(latestTraceAPI.APIs, &traceAPI.Obj)
			if err == nil {
				traceAPI.ReplaceSecret()
				traceAPI.Obj.Timeout = int64(latestTraceAPI.Timeout)
				resp["traceAPI"] = traceAPI.Obj
			}
		}
	}

	resp["chartVersion"] = apoChartVersion
	resp["deployVersion"] = apoComposeDeployVersion

	return resp
}
