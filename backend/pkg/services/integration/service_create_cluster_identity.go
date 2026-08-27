// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	model "github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) CreateClusterIdentity(ctx core.Context, req *model.CreateClusterIdentityRequest) (*model.Cluster, error) {
	cluster := &model.Cluster{
		ID:          req.ID,
		Name:        req.Name,
		ClusterType: req.ClusterType,
	}
	if err := normalizeClusterIdentity(cluster); err != nil {
		return nil, err
	}

	isExist, err := s.dbRepo.CheckClusterNameExisted(ctx, cluster.Name)
	if err != nil {
		return nil, core.Error(code.CreateClusterFailed, fmt.Sprintf("failed to check display name %q: %v", cluster.Name, err))
	}
	if isExist {
		return nil, errors.New("cluster name already exists")
	}

	if err := s.dbRepo.CreateCluster(ctx, cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}
