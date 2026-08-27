// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (s *service) UpdateClusterDisplayName(
	ctx core.Context,
	clusterID string,
	displayName string,
) (*integration.Cluster, error) {
	identity := integration.Cluster{ID: clusterID, Name: displayName}
	if err := normalizeClusterIdentity(&identity); err != nil {
		return nil, err
	}

	cluster, err := s.dbRepo.GetCluster(ctx, identity.ID)
	if err != nil {
		return nil, err
	}
	if cluster.Name == identity.Name {
		return &cluster, nil
	}

	exists, err := s.dbRepo.CheckClusterNameExisted(ctx, identity.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("cluster display name already exists")
	}

	cluster.Name = identity.Name
	if err := s.dbRepo.UpdateCluster(ctx, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}
