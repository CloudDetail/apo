// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import "time"

type ClusterIntegration struct {
	ClusterID   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	ClusterType string `json:"clusterType"` // k8s,vm

	Trace  TraceIntegration  `json:"trace"`
	Metric MetricIntegration `json:"metric"`
	Log    LogIntegration    `json:"log"`
}

func (ci *ClusterIntegration) RemoveSecret() {
	ci.Trace.RemoveSecret()
	ci.Metric.RemoveSecret()
	ci.Log.RemoveSecret()
}

const (
	TraceModeSidecar   = "trace-sidecar"
	TraceModeAll       = "trace"
	TraceModeCollector = "collector"

	MetricModeAll       = "metrics"
	MetricModeCollector = "collector"

	LogModeAll    = "log"
	LogModeSample = "log-sample"
)

type TraceIntegration struct {
	ClusterID string `json:"clusterId" gorm:"primaryKey;column:cluster_id"`

	Mode    string `json:"mode" gorm:"type:varchar(100);column:mode"`
	ApmType string `json:"apmType" gorm:"type:varchar(100);column:apm_type"`

	TraceAPI          JSONField[TraceAPI]          `json:"traceAPI,omitempty" gorm:"type:json;column:trace_api"`
	APOCollector      JSONField[APOCollector]      `json:"apoCollector" gorm:"type:json;column:apo_collector"`
	SelfCollectConfig JSONField[SelfCollectConfig] `json:"selfCollectConfig" gorm:"type:json;column:self_collect_config"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type TraceAPI struct {
	Jaeger     *JaegerAPI     `json:"jaeger,omitempty"`
	Skywalking *SkywalkingAPI `json:"skywalking,omitempty"`
	ARMS       *ARMSAPI       `json:"arms,omitempty"`
	NBS3       *NBS3API       `json:"nbs3,omitempty"`
}

func (i *TraceIntegration) RemoveSecret() {
	if i.TraceAPI.Obj.Skywalking != nil {
		i.TraceAPI.Obj.Skywalking.User = ""
		i.TraceAPI.Obj.Skywalking.Password = ""
	}

	if i.TraceAPI.Obj.ARMS != nil {
		i.TraceAPI.Obj.ARMS.AccessKey = ""
		i.TraceAPI.Obj.ARMS.AccessSecret = ""
	}

	if i.TraceAPI.Obj.NBS3 != nil {
		i.TraceAPI.Obj.NBS3.User = ""
		i.TraceAPI.Obj.NBS3.Password = ""
	}
}

type JaegerAPI struct {
	Address string `json:"address"`
}

type SkywalkingAPI struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ARMSAPI struct {
	Address      string `json:"address"`
	AccessKey    string `json:"accessKey"`
	AccessSecret string `json:"accessSecret"`
}

type NBS3API struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type SelfCollectConfig struct {
	InstrumentAll        bool     `json:"instrumentAll"`
	InstrumentNS         []string `json:"instrumentNS,omitempty"`
	InstrumentDisabledNS []string `json:"instrumentDisabledNS,omitempty"`
}

type APOCollector struct {
	CollectorAddr        string `json:"collectorAddr,omitempty"`
	CollectorGatewayAddr string `json:"collectorGatewayAddr"`
}

type MetricIntegration struct {
	ClusterID string `json:"clusterId" gorm:"primaryKey;column:cluster_id"`

	Mode string `json:"mode" gorm:"type:varchar(100);column:mode"`

	Name   string `json:"name"`
	DSType string `json:"dsType"`

	MetricAPI *JSONField[MetricAPI] `json:"metricAPI" gorm:"type:json;column:metric_api"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type MetricAPI struct {
	VMConfig *VictoriaMetricConfig `json:"vmConfig,omitempty"`
}

func (i *MetricIntegration) RemoveSecret() {
	if i.MetricAPI.Obj.VMConfig != nil {
		i.MetricAPI.Obj.VMConfig.Username = ""
		i.MetricAPI.Obj.VMConfig.Password = ""
	}
}

type PrometheusConfig struct {
	ServerURL string `json:"serverURL"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type VictoriaMetricConfig PrometheusConfig

type LogIntegration struct {
	ClusterID string `json:"clusterId" gorm:"primaryKey;column:cluster_id"`

	Mode string `json:"mode" gorm:"type:varchar(100);column:mode"`

	Name   string `json:"name" gorm:"type:json;column:name"`
	DBType string `json:"dbType" gorm:"type:json;column:db_type"`

	LogAPI *JSONField[LogAPI] `json:"logAPI,omitempty" gorm:"type:json;column:log_api"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (i *LogIntegration) RemoveSecret() {
	if i.LogAPI.Obj.CHConfig != nil {
		i.LogAPI.Obj.CHConfig.UserName = ""
		i.LogAPI.Obj.CHConfig.Password = ""
	}
}

type LogAPI struct {
	CHConfig *ClickhouseConfig `json:"chConfig"`
}

type ClickhouseConfig struct {
	Address     string `json:"address"`
	UserName    string `json:"userName"`
	Password    string `json:"password"` // Encrypt in B64
	Database    string `json:"database"`
	Replication bool   `json:"replication"`
	Cluster     string `json:"cluster"`
}
