package integration

import (
	"bytes"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) GetIntegrationInstallDoc(req *integration.GetCInstallRequest) ([]byte, error) {
	cluster, err := s.dbRepo.GetCluster(req.ClusterID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	switch cluster.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "install.md.tmpl", req.ClusterID)
	case integration.ClusterTypeVM:
		err = dockerComposeTmpl.ExecuteTemplate(&buf, "install.md.tmpl", req.ClusterID)
	default:
		return nil, fmt.Errorf("unexpected clusterType: %s", cluster.ClusterType)
	}

	return buf.Bytes(), err
}
