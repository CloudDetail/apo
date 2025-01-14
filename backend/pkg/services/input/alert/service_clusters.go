// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"github.com/google/uuid"
)

func (s *service) CreateCluster(cluster *alert.Cluster) error {
	cluster.ID = uuid.NewString()
	return s.dbRepo.CreateCluster(cluster)
}

func (s *service) ListCluster() ([]alert.Cluster, error) {
	return s.dbRepo.ListCluster()
}

func (s *service) UpdateCluster(cluster *alert.Cluster) error {
	return s.dbRepo.UpdateCluster(cluster)
}

func (s *service) DeleteCluster(cluster *alert.Cluster) error {
	return s.dbRepo.DeleteCluster(cluster)
}
