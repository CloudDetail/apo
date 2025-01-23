// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

// ApoVMAlertSourceID is the default alert source id of 'APO-VM-ALERT'
// Comes from `uuid.NewMD5(uuid.UUID{}, []byte("APO-VM-ALERT")).String()`
const ApoVMAlertSourceID = "efc91f08-86c4-3696-aba8-570d4a8dc069"

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
	SourceID string `form:"sourceId" json:"sourceId" gorm:"primaryKey;type:varchar(100);column:source_id"`
	SourceInfo
}

type SourceInfo struct {
	SourceName string `form:"sourceName" json:"sourceName" gorm:"unique;type:varchar(100);column:source_name"`
	SourceType string `form:"sourceType" json:"sourceType" gorm:"type:varchar(100);column:source_type"`
}
