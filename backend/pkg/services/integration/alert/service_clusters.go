// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
)

func (s *service) CreateCluster(cluster *input.Cluster) error {
	cluster.ID = uuid.NewString()
	return s.dbRepo.CreateCluster(cluster)
}

func (s *service) ListCluster() ([]input.Cluster, error) {
	return s.dbRepo.ListCluster()
}

func (s *service) UpdateCluster(cluster *input.Cluster) error {
	return s.dbRepo.UpdateCluster(cluster)
}

func (s *service) DeleteCluster(cluster *input.Cluster) error {
	return s.dbRepo.DeleteCluster(cluster)
}
