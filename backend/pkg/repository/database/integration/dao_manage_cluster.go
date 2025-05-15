// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepos) CheckClusterNameExisted(ctx core.Context, clusterName string) (bool, error) {
	var count int64
	err := repo.GetContextDB(ctx).Model(&integration.Cluster{}).Where("name = ?", clusterName).Count(&count).Error
	return count > 0, err
}

func (repo *subRepos) CreateCluster(ctx core.Context, cluster *integration.Cluster) error {
	return repo.GetContextDB(ctx).Create(&cluster).Error
}

func (repo *subRepos) UpdateCluster(ctx core.Context, cluster *integration.Cluster) error {
	return repo.GetContextDB(ctx).Model(&integration.Cluster{}).
		Where("id = ?", cluster.ID).
		Updates(cluster).Error
}

func (repo *subRepos) DeleteCluster(ctx core.Context, cluster *integration.Cluster) error {
	err := repo.GetContextDB(ctx).Delete(&alert.AlertSource2Cluster{}, "cluster_id = ?", cluster.ID).Error
	if err != nil {
		return err
	}

	return repo.GetContextDB(ctx).Delete(&integration.Cluster{}, "id = ?", cluster.ID).Error
}

func (repo *subRepos) ListCluster(ctx core.Context) ([]integration.Cluster, error) {
	var clusters []integration.Cluster
	err := repo.GetContextDB(ctx).Find(&clusters).Error
	return clusters, err
}

func (repo *subRepos) GetCluster(ctx core.Context, clusterID string) (integration.Cluster, error) {
	var cluster integration.Cluster
	err := repo.GetContextDB(ctx).First(&cluster, "id = ?", clusterID).Error
	return cluster, err
}
