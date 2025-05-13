// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) CreateCluster(ctx_core core.Context, cluster *input.Cluster) error {
	cluster.ID = uuid.NewString()
	return s.dbRepo.CreateCluster(cluster)
}

func (s *service) ListCluster(ctx_core core.Context,) ([]input.Cluster, error) {
	return s.dbRepo.ListCluster()
}

func (s *service) UpdateCluster(ctx_core core.Context, cluster *input.Cluster) error {
	return s.dbRepo.UpdateCluster(cluster)
}

func (s *service) DeleteCluster(ctx_core core.Context, cluster *input.Cluster) error {
	return s.dbRepo.DeleteCluster(cluster)
}
