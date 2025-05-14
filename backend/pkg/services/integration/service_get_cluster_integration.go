// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetClusterIntegration(ctx_core core.Context, clusterID string) (*integration.ClusterIntegrationVO, error) {
	config, err := s.dbRepo.GetIntegrationConfig(ctx_core, clusterID)
	if err != nil {
		return nil, err
	}

	return &integration.ClusterIntegrationVO{
		ClusterIntegration:	config.RemoveSecret(),
		ChartVersion:		apoChartVersion,
		DeployVersion:		apoComposeDeployVersion,
	}, nil
}
