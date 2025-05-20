// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) GetClusterIntegration(ctx core.Context, clusterID string) (*integration.ClusterIntegrationVO, error) {
	config, err := s.dbRepo.GetIntegrationConfig(clusterID)
	if err != nil {
		return nil, err
	}

	return &integration.ClusterIntegrationVO{
		ClusterIntegration: config.RemoveSecret(),
		ChartVersion:       apoChartVersion,
		DeployVersion:      apoComposeDeployVersion,
	}, nil
}
