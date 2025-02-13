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
	ci.Trace.RemoveSecret()
	ci.Metric.RemoveSecret()
	ci.Log.RemoveSecret()

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
	ClusterID string `json:"clusterId" gorm:"primaryKey;column:cluster_id"`

	Mode    string `json:"mode" gorm:"type:varchar(100);column:mode"`
	ApmType string `json:"apmType" gorm:"type:varchar(100);column:apm_type"`

	TraceAPI          JSONField[TraceAPI]          `json:"traceAPI,omitempty" gorm:"type:json;column:trace_api"`
	APOCollector      JSONField[APOCollector]      `json:"apoCollector" gorm:"type:json;column:apo_collector"`
	SelfCollectConfig JSONField[SelfCollectConfig] `json:"selfCollectConfig" gorm:"type:json;column:self_collect_config"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsDeleted bool      `gorm:"column:is_deleted;default:false"`
}

// TraceAPI contains config for different APM providers
// remember to update RemoveSecret when updating the struct
type TraceAPI struct {
	Skywalking *SkywalkingConfig `mapstructure:"skywalking"`
	Jaeger     *JaegerConfig     `mapstructure:"jaeger"`
	Nbs3       *Nbs3Config       `mapstructure:"nbs3"`
	Arms       *ArmsConfig       `mapstructure:"arms"`
	Huawei     *HuaweiConfig     `mapstructure:"huawei"`
	Elastic    *ElasticConfig    `mapstructure:"elastic"`
	Pinpoint   *PinpointConfig   `mapstructure:"pinpoint"`

	Timeout time.Duration `json:"timeout"`
}

func (i *TraceIntegration) RemoveSecret() {
	if i.TraceAPI.Obj.Skywalking != nil {
		i.TraceAPI.Obj.Skywalking.User = ""
		i.TraceAPI.Obj.Skywalking.Password = ""
	}

	if i.TraceAPI.Obj.Arms != nil {
		i.TraceAPI.Obj.Arms.AccessKey = ""
		i.TraceAPI.Obj.Arms.AccessSecret = ""
	}

	if i.TraceAPI.Obj.Nbs3 != nil {
		i.TraceAPI.Obj.Nbs3.User = ""
		i.TraceAPI.Obj.Nbs3.Password = ""
	}

	if i.TraceAPI.Obj.Huawei != nil {
		i.TraceAPI.Obj.Huawei.AccessKey = ""
		i.TraceAPI.Obj.Huawei.AccessSecret = ""
	}

	if i.TraceAPI.Obj.Elastic != nil {
		i.TraceAPI.Obj.Elastic.Address = ""
		i.TraceAPI.Obj.Elastic.User = ""
		i.TraceAPI.Obj.Elastic.Password = ""
	}
}

type SkywalkingConfig struct {
	Address  string `mapstructure:"address"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type JaegerConfig struct {
	Address string `mapstructure:"address"`
}

type Nbs3Config struct {
	Address  string `mapstructure:"address"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type ArmsConfig struct {
	Address      string `mapstructure:"address"`
	AccessKey    string `mapstructure:"access_key"`
	AccessSecret string `mapstructure:"access_secret"`
}

type HuaweiConfig struct {
	AccessKey    string `mapstructure:"access_key"`
	AccessSecret string `mapstructure:"access_secret"`
}

type ElasticConfig struct {
	Address  string `mapstructure:"address"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type PinpointConfig struct {
	Address string `mapstructure:"address"`
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

	Mode   string `json:"mode" gorm:"type:varchar(100);column:mode"`
	Name   string `json:"name"`
	DSType string `json:"dsType"`

	MetricAPI *JSONField[MetricAPI] `json:"metricAPI" gorm:"type:json;column:metric_api"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsDeleted bool      `gorm:"column:is_deleted;default:false"`
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
	IsDeleted bool      `gorm:"column:is_deleted;default:false"`
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

type AdapterAPIConfig struct {
	APIs    map[string]any `json:"apis"`
	Timeout int64          `json:"timeout"`
}
