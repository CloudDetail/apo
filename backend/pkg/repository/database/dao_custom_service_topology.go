// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/core"
)

type CustomServiceTopology struct {
	ID         int    `gorm:"column:id;primary_key;auto_increment" json:"id"`
	ClusterId  string `gorm:"column:cluster_id;type:varchar(100)" json:"clusterId"`
	LeftNode   string `gorm:"column:left_node;type:varchar(200)" json:"leftNode"`
	LeftType   string `gorm:"column:left_type;type:varchar(20)" json:"leftType"`
	RightNode  string `gorm:"column:right_node;type:varchar(200)" json:"rightNode"`
	RightType  string `gorm:"column:right_type;type:varchar(20)" json:"rightType"`
	StartTime  int64  `gorm:"column:start_time" json:"startTime"`
	ExpireTime int64  `gorm:"column:expire_time" json:"expireTime"`
}

func (CustomServiceTopology) TableName() string {
	return "custom_service_topology"
}

func (repo *daoRepo) CreateCustomServiceTopology(ctx core.Context, topology *CustomServiceTopology) error {
	var count int64
	repo.GetContextDB(ctx).Model(&CustomServiceTopology{}).Where("cluster_id = ? AND left_node = ? AND right_node = ?", topology.ClusterId, topology.LeftNode, topology.RightNode).Count(&count)
	if count > 0 {
		return errors.New("已经存在自定义拓扑")
	}
	return repo.GetContextDB(ctx).Create(topology).Error
}

func (repo *daoRepo) ListCustomServiceTopology(ctx core.Context) ([]CustomServiceTopology, error) {
	var topologies []CustomServiceTopology
	err := repo.GetContextDB(ctx).
		Model(&CustomServiceTopology{}).
		Order("cluster_id ASC, left_node ASC").
		Scan(&topologies).Error
	return topologies, err
}

func (repo *daoRepo) DeleteCustomServiceTopology(ctx core.Context, id int) error {
	return repo.GetContextDB(ctx).Model(&CustomServiceTopology{}).Where("id = ?", id).Delete(nil).Error
}
