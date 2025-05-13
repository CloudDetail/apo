// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (repo *subRepos) CheckClusterNameExisted(ctx_core core.Context, clusterName string) (bool, error) {
	var count int64
	err := repo.db.Model(&integration.Cluster{}).Where("name = ?", clusterName).Count(&count).Error
	return count > 0, err
}

func (repo *subRepos) CreateCluster(ctx_core core.Context, cluster *integration.Cluster) error {
	return repo.db.Create(&cluster).Error
}

func (repo *subRepos) UpdateCluster(ctx_core core.Context, cluster *integration.Cluster) error {
	return repo.db.Model(&integration.Cluster{}).
		Where("id = ?", cluster.ID).
		Updates(cluster).Error
}

func (repo *subRepos) DeleteCluster(ctx_core core.Context, cluster *integration.Cluster) error {
	err := repo.db.Delete(&alert.AlertSource2Cluster{}, "cluster_id = ?", cluster.ID).Error
	if err != nil {
		return err
	}

	return repo.db.Delete(&integration.Cluster{}, "id = ?", cluster.ID).Error
}

func (repo *subRepos) ListCluster(ctx_core core.Context,) ([]integration.Cluster, error) {
	var clusters []integration.Cluster
	err := repo.db.Find(&clusters).Error
	return clusters, err
}

func (repo *subRepos) GetCluster(ctx_core core.Context, clusterID string) (integration.Cluster, error) {
	var cluster integration.Cluster
	err := repo.db.First(&cluster, "id = ?", clusterID).Error
	return cluster, err
}
