// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepos) CheckClusterNameExisted(clusterName string) (bool, error) {
	var count int64
	err := repo.db.Model(&integration.Cluster{}).Where("name = ?", clusterName).Count(&count).Error
	return count > 0, err
}

func (repo *subRepos) CreateCluster(cluster *integration.Cluster) error {
	return repo.db.Create(&cluster).Error
}

func (repo *subRepos) UpdateCluster(cluster *integration.Cluster) error {
	return repo.db.Model(&integration.Cluster{}).
		Where("id = ?", cluster.ID).
		Updates(cluster).Error
}

func (repo *subRepos) DeleteCluster(cluster *integration.Cluster) error {
	err := repo.db.Delete(&alert.AlertSource2Cluster{}, "cluster_id = ?", cluster.ID).Error
	if err != nil {
		return err
	}

	return repo.db.Delete(&integration.Cluster{}, "id = ?", cluster.ID).Error
}

func (repo *subRepos) ListCluster() ([]integration.Cluster, error) {
	var clusters []integration.Cluster
	err := repo.db.Find(&clusters).Error
	return clusters, err
}

func (repo *subRepos) GetCluster(clusterID string) (integration.Cluster, error) {
	var cluster integration.Cluster
	err := repo.db.First(&cluster, "id = ?", clusterID).Error
	return cluster, err
}
