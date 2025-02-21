// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) ListCluster() ([]integration.Cluster, error) {
	return s.dbRepo.ListCluster()
}

func (s *service) DeleteCluster(cluster *integration.Cluster) error {
	err := s.dbRepo.DeleteCluster(cluster)
	if err != nil {
		return err
	}

	return s.dbRepo.DeleteIntegrationConfig(cluster.ID)
}
