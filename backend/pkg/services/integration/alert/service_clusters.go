// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
)

func (s *service) CreateCluster(ctx core.Context, cluster *input.Cluster) error {
	cluster.ID = uuid.NewString()
	return s.dbRepo.CreateCluster(cluster)
}

func (s *service) ListCluster(ctx core.Context) ([]input.Cluster, error) {
	return s.dbRepo.ListCluster()
}

func (s *service) UpdateCluster(ctx core.Context, cluster *input.Cluster) error {
	return s.dbRepo.UpdateCluster(cluster)
}

func (s *service) DeleteCluster(ctx core.Context, cluster *input.Cluster) error {
	return s.dbRepo.DeleteCluster(cluster)
}
