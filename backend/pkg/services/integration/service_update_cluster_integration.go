// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import "github.com/CloudDetail/apo/backend/pkg/model/integration"

func (s *service) UpdateClusterIntegration(cluster *integration.ClusterIntegrationVO) error {
	err := s.dbRepo.UpdateCluster(&cluster.Cluster)
	if err != nil {
		return err
	}

	return s.dbRepo.SaveIntegrationConfig(integration.ClusterIntegration{
		ClusterID:   cluster.ID,
		ClusterType: cluster.ClusterType,
		Trace:       cluster.Trace,
		Metric:      cluster.Metric,
		Log:         cluster.Log,
	})
}
