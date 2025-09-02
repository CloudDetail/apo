// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"encoding/base64"
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) GetIntegrationInstallDoc(ctx core.Context, req *integration.GetCInstallRequest) (string, error) {
	cluster, err := s.dbRepo.GetCluster(ctx, req.ClusterID)
	if err != nil {
		return "", err
	}

	clusterConfig, err := s.dbRepo.GetIntegrationConfig(ctx, req.ClusterID)
	if err != nil {
		return "", err
	}

	if len(clusterConfig.Cluster.APOCollector.CollectorGatewayAddr) == 0 {
		clusterConfig.Cluster.APOCollector.CollectorGatewayAddr = s.serverAddr
	}

	if cluster.IsMinimal {
		return s.getMinimalIntegrationInstallCmd(clusterConfig)
	}

	jsonObj, err := s.convert2DeployValues(clusterConfig)
	if err != nil {
		return "", fmt.Errorf("marshal config failed: %w", err)
	}

	var buf bytes.Buffer
	switch cluster.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "install.md.tmpl", jsonObj)
	case integration.ClusterTypeVM:
		err = dockerComposeTmpl.ExecuteTemplate(&buf, "install.md.tmpl", jsonObj)
	default:
		return "", fmt.Errorf("unexpected clusterType: %s", cluster.ClusterType)
	}

	md := buf.Bytes()
	mdB64 := base64.StdEncoding.EncodeToString(md)
	return mdB64, err
}

func (s *service) getMinimalIntegrationInstallCmd(clusterConfig *integration.ClusterIntegration) (string, error) {
	jsonObj, err := s.convert2DeployValues(clusterConfig)
	if err != nil {
		return "", fmt.Errorf("marshal config failed: %w", err)
	}

	var buf bytes.Buffer
	switch clusterConfig.ClusterType {
	case integration.ClusterTypeK8s:
		err = k8sTmpl.ExecuteTemplate(&buf, "install-cmd-minimal.tmpl", jsonObj)
	case integration.ClusterTypeVM:
		err = dockerComposeTmpl.ExecuteTemplate(&buf, "install-cmd-minimal.tmpl", jsonObj)
	default:
		return "", fmt.Errorf("unexpected clusterType: %s", clusterConfig.ClusterType)
	}

	return buf.String(), err
}
