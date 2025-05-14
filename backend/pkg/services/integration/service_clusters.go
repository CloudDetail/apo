// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) ListCluster(ctx core.Context) ([]integration.Cluster, error) {
	return s.dbRepo.ListCluster(ctx)
}

func (s *service) DeleteCluster(ctx core.Context, cluster *integration.Cluster) error {
	err := s.dbRepo.DeleteCluster(ctx, cluster)
	if err != nil {
		return err
	}

	return s.dbRepo.DeleteIntegrationConfig(ctx, cluster.ID)
}
