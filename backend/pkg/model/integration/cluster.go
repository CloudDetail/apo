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
	selfCollectMode  = "self-collect"
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
	var modes []string

	switch ci.Trace.Mode {
	case sideCarTraceMode:
		modes = append(modes, "trace-sidecar")
	case collectTraceMode:
		modes = append(modes, "trace-collector")
	case selfCollectMode:
		modes = append(modes, "trace")
	}

	switch ci.Metric.Mode {
	case selfCollectMode:
		modes = append(modes, "metrics")
	}

	logMode := "log"
	if ci.Log.LogSelfCollectConfig != nil && ci.Log.LogSelfCollectConfig.Obj.Mode == "sample" {
		logMode = "log-sample"
	}
	modes = append(modes, logMode)

	jsonObj["modes"] = modes
	return jsonObj, nil
}
