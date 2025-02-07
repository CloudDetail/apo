package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"text/template"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

const (
	k8sTmplFilePattern           = "static/integration-tmpl/kubernetes/*.tmpl"
	dockerComposeTmplFilePattern = "static/integration-tmpl/dockerCompose/*.tmpl"
)

var (
	k8sTmpl           *template.Template
	dockerComposeTmpl *template.Template
)

func init() {
	var err error
	k8sTmpl, err = template.ParseGlob(k8sTmplFilePattern)
	if err != nil {
		log.Printf("[integration] module failed, cannot load k8s integration template files: %v", err)
	}
	dockerComposeTmpl, err = template.ParseGlob(dockerComposeTmplFilePattern)
	if err != nil {
		log.Printf("[integration] module failed, cannot load dockerCompose integration template files: %v", err)
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
	jsonStr, err := json.Marshal(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal config failed: %w", err)
	}
	jsonObj := map[string]any{}
	err = json.Unmarshal(jsonStr, &jsonObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	var buf bytes.Buffer
	var fileName string
	switch clusterConfig.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "apo-one-agent-values.yaml.tmpl", jsonObj)
		fileName = "apo-one-agent-values.yaml"
	case integration.ClusterTypeVM:
		_, err = buf.WriteString("vm cluster not have config now")
		fileName = "empty-config.yaml"
	default:
		return nil, fmt.Errorf("unexpected clusterType: %s", clusterConfig.ClusterType)
	}

	return &integration.GetCInstallConfigResponse{
		FileName: fileName,
		Content:  buf.Bytes(),
	}, err
}
