// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/input/alert"

func (repo *subRepo) CreateCluster(cluster *alert.Cluster) error {
	return repo.db.Create(&cluster).Error
}

func (repo *subRepo) UpdateCluster(cluster *alert.Cluster) error {
	return repo.db.Model(&alert.Cluster{}).
		Where("id = ?", cluster.ID).
		Updates(cluster).Error
}

func (repo *subRepo) DeleteCluster(cluster *alert.Cluster) error {
	err := repo.db.Delete(&alert.AlertSource2Cluster{}, "cluster_id = ?", cluster.ID).Error
	if err != nil {
		return err
	}

	return repo.db.Delete(&alert.Cluster{}, "id = ?", cluster.ID).Error
}

func (repo *subRepo) ListCluster() ([]alert.Cluster, error) {
	var clusters []alert.Cluster
	err := repo.db.Find(&clusters).Error
	return clusters, err
}
