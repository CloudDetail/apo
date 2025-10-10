// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/template"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

const (
	k8sTmplFilePattern           = "static/integration-tmpl/kubernetes/*.tmpl"
	dockerComposeTmplFilePattern = "static/integration-tmpl/dockercompose/*.tmpl"
)

var (
	k8sTmpl           *template.Template
	dockerComposeTmpl *template.Template
)

func init() {
	var err error

	k8sTmpl, err = template.New("k8sTmpl").Funcs(template.FuncMap{
		"default": defaultValue,
	}).ParseGlob(k8sTmplFilePattern)
	if err != nil {
		log.Printf("[integration] module failed, cannot load k8s integration template files: %v", err)
	}
	dockerComposeTmpl, err = template.New("dockerComposeTmpl").Funcs(template.FuncMap{
		"default": defaultValue,
	}).ParseGlob(dockerComposeTmplFilePattern)
	if err != nil {
		log.Printf("[integration] module failed, cannot load dockerCompose integration template files: %v", err)
	}

	if value, find := os.LookupEnv("APO_CHART_VERSION"); find {
		apoChartVersion = value
	}
	if value, find := os.LookupEnv("APO_DEPLOY_VERSION"); find {
		apoComposeDeployVersion = value
	}

}

func defaultValue(v any, def any) string {
	var defaultValue string
	switch def := def.(type) {
	case string:
		defaultValue = def
	case int:
		defaultValue = strconv.Itoa(def)
	case int64:
		defaultValue = strconv.FormatInt(def, 10)
	}

	if v == nil {
		return defaultValue
	}

	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return defaultValue
	}
}

func (s *service) GetIntegrationInstallConfigFile(ctx core.Context, req *integration.GetCInstallRequest) (*integration.GetCInstallConfigResponse, error) {
	clusterConfig, err := s.dbRepo.GetIntegrationConfig(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}

	if len(clusterConfig.Cluster.APOCollector.CollectorGatewayAddr) == 0 {
		clusterConfig.Cluster.APOCollector.CollectorGatewayAddr = s.serverAddr
	}

	return s.getIntegrationConfigFile(clusterConfig)
}

func (s *service) getIntegrationConfigFile(clusterConfig *integration.ClusterIntegration) (*integration.GetCInstallConfigResponse, error) {
	jsonObj, err := s.convert2DeployValues(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal config failed: %w", err)
	}

	var buf bytes.Buffer
	var fileName string
	switch clusterConfig.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "apo-one-agent-values.yaml.tmpl", jsonObj)
		fileName = "agent-values.yaml"
	case integration.ClusterTypeVM:
		err = dockerComposeTmpl.ExecuteTemplate(&buf, "installCfg.sh.tmpl", jsonObj)
		fileName = "installCfg.sh"
	default:
		return nil, fmt.Errorf("unexpected clusterType: %s", clusterConfig.ClusterType)
	}

	return &integration.GetCInstallConfigResponse{
		FileName: fileName,
		Content:  buf.Bytes(),
	}, err
}

const (
	sideCarTraceMode = "sidecar"
	collectTraceMode = "collect"
	selfCollectMode  = "self-collector"
)

func init() {
	if value, find := os.LookupEnv("APO_CHART_VERSION"); find {
		apoChartVersion = value
	}
}

var (
	apoChartVersion         = "1.11"
	apoComposeDeployVersion = "v1.11.000"
)

func (s *service) convert2DeployValues(ci *integration.ClusterIntegration) (map[string]any, error) {
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
			jsonObj["_java_agent_type"] = "SKYWALKING"
		default:
			jsonObj["_java_agent_type"] = "OPENTELEMETRY"
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

	jsonObj["_deploy_version"] = apoComposeDeployVersion
	jsonObj["_chart_version"] = apoChartVersion
	jsonObj["_cluster_id"] = ci.Cluster.ID
	jsonObj["_is_minimal"] = ci.Cluster.IsMinimal
	jsonObj["_base_url_"] = s.baseURL

	return jsonObj, nil
}
