// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

var PlatformClusterID = uuid.NewMD5(uuid.Nil, []byte("APO-Platform"))

const (
	ClusterTypeK8s = "k8s"
	ClusterTypeVM  = "vm"
)

type Cluster struct {
	ID           string       `form:"id" json:"id" gorm:"primaryKey;type:varchar(100);column:id"`
	Name         string       `form:"name" json:"name" gorm:"unique;type:varchar(100);column:name"`
	ClusterType  string       `form:"clusterType" json:"clusterType" gorm:"type:varchar(100);column:cluster_type"`
	APOCollector APOCollector `json:"apoCollector,omitempty" gorm:"type:json;column:apo_collector"`
}

type ClusterIntegration struct {
	Cluster

	Trace  TraceIntegration  `json:"trace"`
	Metric MetricIntegration `json:"metric"`
	Log    LogIntegration    `json:"log"`
}

const (
	sideCarTraceMode = "sidecar"
	collectTraceMode = "collect"
	selfCollectMode  = "self-collector"
)

func (ci *ClusterIntegration) ConvertToHelmValues() (map[string]any, error) {
	jsonStr, err := json.Marshal(ci)
	if err != nil {
		return nil, fmt.Errorf("marshal config failed: %w", err)
	}
	jsonObj := map[string]any{}
	err = json.Unmarshal(jsonStr, &jsonObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	var modes = make(map[string]string)
	switch ci.Trace.Mode {
	case sideCarTraceMode:
		modes["trace"] = "trace-sidecar"
	case collectTraceMode:
		modes["trace"] = "trace-collector"
	case selfCollectMode:
		modes["trace"] = "trace"
		switch ci.Trace.ApmType {
		case "skywalking":
			modes["_java_agent_type"] = "SKYWALKING"
		default:
			modes["_java_agent_type"] = "OPENTELEMETRY"
		}
	}

	if ci.Metric.DSType == selfCollectMode {
		modes["metric"] = "metrics"
	}

	if ci.Log.DBType == selfCollectMode {
		logMode := "log"
		if ci.Log.LogSelfCollectConfig != nil && ci.Log.LogSelfCollectConfig.Obj.Mode == "sample" {
			logMode = "log-sample"
		}
		modes["log"] = logMode
	}

	jsonObj["_modes"] = modes

	if ci.ClusterType == ClusterTypeK8s {
		jsonObj["_deploy_version"] = "v1.2.000"
		jsonObj["_app_version"] = "v1.2.0"
	} else {
		jsonObj["_deploy_version"] = "v1.3.000"
		jsonObj["_app_version"] = "v1.3.0"
	}

	return jsonObj, nil
}
