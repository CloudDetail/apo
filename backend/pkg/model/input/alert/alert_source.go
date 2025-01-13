// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type AlertSource struct {
	SourceFrom

	Clusters []Cluster `json:"clusters" gorm:"-"`
}

type Cluster struct {
	ID   string `json:"id" gorm:"primaryKey;type:varchar(100);column:id"`
	Name string `json:"name" gorm:"unique;type:varchar(100);column:name"`
}

type AlertSource2Cluster struct {
	SourceID  string `gorm:"type:varchar(100);column:source_id"`
	ClusterID string `gorm:"type:varchar(100);column:cluster_id"`
}

type SourceFrom struct {
	SourceID string `json:"sourceId" gorm:"primaryKey;type:varchar(100);column:source_id"`
	SourceInfo
}

type SourceInfo struct {
	SourceName string `json:"sourceName" gorm:"unique;type:varchar(100);column:source_name"`
	SourceType string `json:"sourceType" gorm:"type:varchar(100);column:source_type"`
}
