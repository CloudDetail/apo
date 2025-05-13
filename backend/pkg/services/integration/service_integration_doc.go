// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// Deprecated
func (s *service) GetIntegrationInstallDoc(ctx_core core.Context, req *integration.GetCInstallRequest) ([]byte, error) {
	cluster, err := s.dbRepo.GetCluster(req.ClusterID)
	if err != nil {
		return nil, err
	}

	clusterConfig, err := s.dbRepo.GetIntegrationConfig(req.ClusterID)
	if err != nil {
		return nil, err
	}

	jsonObj, err := convert2DeployValues(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal config failed: %w", err)
	}

	var buf bytes.Buffer
	switch cluster.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "install.md.tmpl", jsonObj)
	case integration.ClusterTypeVM:
		err = dockerComposeTmpl.ExecuteTemplate(&buf, "install.md.tmpl", jsonObj)
	default:
		return nil, fmt.Errorf("unexpected clusterType: %s", cluster.ClusterType)
	}

	return buf.Bytes(), err
}
