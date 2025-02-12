// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import "github.com/google/uuid"

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

type ClusterIntegrationVO struct {
	Cluster

	Trace  TraceIntegration  `json:"trace"`
	Metric MetricIntegration `json:"metric"`
	Log    LogIntegration    `json:"log"`
}
