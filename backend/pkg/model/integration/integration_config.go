// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"time"
)

type ClusterIntegration struct {
	ClusterID   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	ClusterType string `json:"clusterType"` // k8s,vm

	Trace  TraceIntegration  `json:"trace"`
	Metric MetricIntegration `json:"metric"`
	Log    LogIntegration    `json:"log"`
}

func (ci *ClusterIntegration) RemoveSecret() *ClusterIntegrationVO {
	ci.Trace.TraceAPI.ReplaceSecret()
	ci.Metric.MetricAPI.ReplaceSecret()
	ci.Log.LogAPI.ReplaceSecret()

	return &ClusterIntegrationVO{
		Cluster: Cluster{
			ID:          ci.ClusterID,
			Name:        ci.ClusterName,
			ClusterType: ci.ClusterType,
		},
		Trace:  ci.Trace,
		Metric: ci.Metric,
		Log:    ci.Log,
	}
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
	ClusterID string `json:"clusterId,omitempty" gorm:"primaryKey;column:cluster_id"`

	Mode    string `json:"mode" gorm:"type:varchar(100);column:mode"`
	ApmType string `json:"apmType" gorm:"type:varchar(100);column:apm_type"`

	TraceAPI          JSONField[TraceAPI]          `json:"traceAPI,omitempty" gorm:"type:json;column:trace_api"`
	APOCollector      JSONField[APOCollector]      `json:"apoCollector" gorm:"type:json;column:apo_collector"`
	SelfCollectConfig JSONField[SelfCollectConfig] `json:"selfCollectConfig" gorm:"type:json;column:self_collect_config"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsDeleted bool      `gorm:"column:is_deleted;default:false"`
}

// TraceAPI contains config for different APM providers
// using secret:"true" to hide the secret
type TraceAPI struct {
	Skywalking *SkywalkingConfig `json:"skywalking,omitempty" mapstructure:"skywalking"`
	Jaeger     *JaegerConfig     `json:"jaeger,omitempty" mapstructure:"jaeger"`
	Nbs3       *Nbs3Config       `json:"nbs3,omitempty" mapstructure:"nbs3"`
	Arms       *ArmsConfig       `json:"arms,omitempty" mapstructure:"arms"`
	Huawei     *HuaweiConfig     `json:"huawei,omitempty" mapstructure:"huawei"`
	Elastic    *ElasticConfig    `json:"elastic,omitempty" mapstructure:"elastic"`
	Pinpoint   *PinpointConfig   `json:"pinpoint,omitempty" mapstructure:"pinpoint"`

	// Second
	Timeout int64 `json:"timeout"`
}

type SkywalkingConfig struct {
	Address  string `json:"address" mapstructure:"address"`
	User     string `json:"user" mapstructure:"user" secret:"true"`
	Password string `json:"password" mapstructure:"password" secret:"true"`
}

type JaegerConfig struct {
	Address string `json:"address" mapstructure:"address"`
}

type Nbs3Config struct {
	Address  string `json:"address" mapstructure:"address"`
	User     string `json:"user" mapstructure:"user" secret:"true"`
	Password string `json:"password" mapstructure:"password" secret:"true"`
}

type ArmsConfig struct {
	Address      string `json:"address" mapstructure:"address"`
	AccessKey    string `json:"accessKey" mapstructure:"access_key" secret:"true"`
	AccessSecret string `json:"accessSecret" mapstructure:"access_secret" secret:"true"`
}

type HuaweiConfig struct {
	AccessKey    string `json:"accessKey" mapstructure:"access_key" secret:"true"`
	AccessSecret string `json:"accessSecret" mapstructure:"access_secret" secret:"true"`
}

type ElasticConfig struct {
	Address  string `json:"address" mapstructure:"address"`
	User     string `json:"user" mapstructure:"user" secret:"true"`
	Password string `json:"password" mapstructure:"password" secret:"true"`
}

type PinpointConfig struct {
	Address string `json:"address" mapstructure:"address"`
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
	ClusterID string `json:"clusterId,omitempty" gorm:"primaryKey;column:cluster_id"`

	Mode   string `json:"mode" gorm:"type:varchar(100);column:mode"`
	Name   string `json:"name"`
	DSType string `json:"dsType"`

	MetricAPI *JSONField[MetricAPI] `json:"metricAPI" gorm:"type:json;column:metric_api"`

	UpdatedAt time.Time `json:"-" gorm:"autoUpdateTime"`
	IsDeleted bool      `json:"-" gorm:"column:is_deleted;default:false"`
}

type MetricAPI struct {
	VictoriaMetric *VictoriaMetricConfig `json:"victoriametric,omitempty"`
}

type PrometheusConfig struct {
	ServerURL string `json:"serverURL"`
	Username  string `json:"username" secret:"true"`
	Password  string `json:"password" secret:"true"`
}

type VictoriaMetricConfig PrometheusConfig

type LogIntegration struct {
	ClusterID string `json:"clusterId,omitempty" gorm:"primaryKey;column:cluster_id"`

	Mode string `json:"mode" gorm:"type:varchar(100);column:mode"`

	Name   string `json:"name" gorm:"type:json;column:name"`
	DBType string `json:"dbType" gorm:"type:json;column:db_type"`

	LogAPI *JSONField[LogAPI] `json:"logAPI,omitempty" gorm:"type:json;column:log_api"`

	UpdatedAt time.Time `json:"-" gorm:"autoUpdateTime"`
	IsDeleted bool      `json:"-" gorm:"column:is_deleted;default:false"`
}

type LogAPI struct {
	Clickhouse *ClickhouseConfig `json:"clickhouse"`
}

type ClickhouseConfig struct {
	Address     string `json:"address"`
	UserName    string `json:"userName" secret:"true"`
	Password    string `json:"password" secret:"true"` // Encrypt in B64
	Database    string `json:"database"`
	Replication bool   `json:"replication"`
	Cluster     string `json:"cluster"`
}

type AdapterAPIConfig struct {
	APIs    map[string]any `json:"apis"`
	Timeout int64          `json:"timeout"`
}
