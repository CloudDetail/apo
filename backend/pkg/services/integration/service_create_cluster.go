// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
)

func (s *service) CreateCluster(cluster *integration.ClusterIntegrationVO) error {
	isExist, err := s.dbRepo.CheckClusterNameExisted(cluster.Name)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("cluster name already exists")
	}

	cluster.ID = uuid.NewString()
	err = s.dbRepo.CreateCluster(&cluster.Cluster)
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
