// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"text/template"

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

func (s *service) GetIntegrationInstallConfigFile(req *integration.GetCInstallRequest) (*integration.GetCInstallConfigResponse, error) {
	clusterConfig, err := s.dbRepo.GetIntegrationConfig(req.ClusterID)
	if err != nil {
		return nil, err
	}

	return getIntegrationConfigFile(clusterConfig)
}

func getIntegrationConfigFile(clusterConfig *integration.ClusterIntegration) (*integration.GetCInstallConfigResponse, error) {
	jsonObj, err := clusterConfig.ConvertToHelmValues()
	if err != nil {
		return nil, fmt.Errorf("marshal config failed: %w", err)
	}

	var buf bytes.Buffer
	var fileName string
	switch clusterConfig.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "apo-one-agent-values.yaml.tmpl", jsonObj)
		fileName = "apo-one-agent-values.yaml"
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
