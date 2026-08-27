// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"
	"testing"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	model "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/stretchr/testify/require"
)

type clusterIdentityRepo struct {
	database.Repo
	created                *model.Cluster
	saveIntegrationConfigs int
}

func (r *clusterIdentityRepo) CheckClusterNameExisted(core.Context, string) (bool, error) {
	return false, nil
}

func (r *clusterIdentityRepo) CreateCluster(_ core.Context, cluster *model.Cluster) error {
	copy := *cluster
	r.created = &copy
	return nil
}

func (r *clusterIdentityRepo) SaveIntegrationConfig(core.Context, model.ClusterIntegration) error {
	r.saveIntegrationConfigs++
	return errors.New("cluster identity creation must not save integration config")
}

func TestCreateClusterIdentityOnlyPersistsIdentity(t *testing.T) {
	repo := &clusterIdentityRepo{}
	service := &service{dbRepo: repo}

	cluster, err := service.CreateClusterIdentity(nil, &model.CreateClusterIdentityRequest{
		ID:          " cluster-a ",
		Name:        " Production ",
		ClusterType: "kubernetes",
	})

	require.NoError(t, err)
	require.Equal(t, &model.Cluster{ID: "cluster-a", Name: "Production", ClusterType: model.ClusterTypeK8s}, cluster)
	require.Equal(t, cluster, repo.created)
	require.Zero(t, repo.saveIntegrationConfigs)
}
