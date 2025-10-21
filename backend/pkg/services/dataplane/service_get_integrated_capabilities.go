// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"os"
	"slices"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

var TelemetryDescriptions = map[integration.TelemetryType]map[string]string{
	integration.TelemetryMetrics: {
		"zh": "缺少 Metrics 数据会导致无法对资源使用进行分析",
		"en": "Lack of metrics data makes it impossible to analyze resource usage",
	},
	integration.TelemetryLogs: {
		"zh": "缺少 Log 数据会导致无法查看错误与审计日志",
		"en": "Lack of log data makes it impossible to view error and audit logs",
	},
	integration.TelemetryTraces: {
		"zh": "缺少 Trace 数据会导致链路拓扑分析异常",
		"en": "Lack of trace data causes issues with trace topology analysis",
	},
	integration.TelemetryRedMetric: {
		"zh": "缺少 RED 指标会影响对服务的请求数、错误率和延迟的分析",
		"en": "Lack of RED metrics will affect request rate, error rate and latency analysis",
	},
	integration.TelemetryTopology: {
		"zh": "缺少拓扑数据会影响对全局服务关系图的分析",
		"en": "Lack of topology data makes it impossible to analyze global service relationship graphs",
	},
	integration.TelemetryServiceInstance: {
		"zh": "缺少服务实例信息会影响实例维度分析",
		"en": "Lack of service instance data makes it impossible to analyze instance-level metrics",
	},
}

func getTelemetryCapabilityDescription(t integration.TelemetryType, lang string) string {
	if desc, ok := TelemetryDescriptions[t]; ok {
		if desc, ok := desc[lang]; ok {
			return desc
		}
	}
	return TelemetryDescriptions[t]["en"]
}

func (s *service) GetIntegratedCapabilities(ctx core.Context) (*response.GetIntegratedCapabilitiesResponse, error) {
	dps, err := s.dbRepo.ListDataPlane(ctx)
	if err != nil {
		return nil, err
	}

	var res response.GetIntegratedCapabilitiesResponse
	res.TotalSupported = make([]integration.TelemetryCapability, 0)

	capMap := make(map[integration.TelemetryType]*integration.TelemetryCapability)
	for _, t := range integration.AllTelemetryTypes {
		capMap[t] = &integration.TelemetryCapability{
			Type:        t,
			Enabled:     false,
			Sources:     []integration.DataplaneWithName{},
			Description: getTelemetryCapabilityDescription(t, ctx.LANG()),
		}
	}

	for _, dp := range dps {
		for _, t := range integration.AllTelemetryTypes {
			cap := capMap[t]
			if slices.Contains(dp.DataPlaneWithType.DataPlane.Capability.Obj, string(t)) {
				dpType := dp.DataPlaneWithType.DataPlane.Typ
				dpWithName := integration.DataplaneWithName{
					Type: dpType,
					Name: dp.DataPlaneWithType.DataPlane.Name,
				}
				if !slices.Contains(cap.Sources, dpWithName) && (dpType != "apo" || os.Getenv(showAPODataPlaneTypeENV) == "true") {
					cap.Enabled = true
					cap.Sources = append(cap.Sources, dpWithName)
				}
			}
		}
	}

	for _, cap := range capMap {
		res.TotalSupported = append(res.TotalSupported, *cap)
	}

	return &res, nil
}
