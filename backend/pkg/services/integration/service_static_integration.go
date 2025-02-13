package integration

import (
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/mitchellh/mapstructure"
)

// HACK use static config in configFile now
func (s *service) GetStaticIntegration() map[string]any {
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
		UpdatedAt: time.Time{},
		IsDeleted: false,
	}
	ds.MetricAPI.ReplaceSecret()
	resp["datasource"] = ds

	latestTraceAPI, err := s.dbRepo.GetLatestTraceAPIs(-1)
	if err == nil {
		var traceAPI integration.JSONField[integration.TraceAPI]
		err := mapstructure.Decode(latestTraceAPI.APIs, &traceAPI.Obj)
		if err == nil {
			traceAPI.ReplaceSecret()
			traceAPI.Obj.Timeout = latestTraceAPI.Timeout
			resp["traceAPI"] = traceAPI.Obj
		}
	}
	return resp
}
