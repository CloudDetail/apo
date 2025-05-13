// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) ListCluster(ctx_core core.Context,) ([]integration.Cluster, error) {
	return s.dbRepo.ListCluster()
}

func (s *service) DeleteCluster(ctx_core core.Context, cluster *integration.Cluster) error {
	err := s.dbRepo.DeleteCluster(cluster)
	if err != nil {
		return err
	}

	return s.dbRepo.DeleteIntegrationConfig(cluster.ID)
}
