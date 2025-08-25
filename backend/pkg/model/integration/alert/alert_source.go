// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/integration"

// ApoVMAlertSourceID is the default alert source id of 'APO-VM-ALERT'
// Comes from `uuid.NewMD5(uuid.UUID{}, []byte("APO-VM-ALERT")).String()`
const ApoVMAlertSourceID = "efc91f08-86c4-3696-aba8-570d4a8dc069"

type AlertSource struct {
	SourceFrom
	Clusters []integration.Cluster `json:"clusters" gorm:"-"`

	Params         integration.JSONField[AlertSourceParams] `json:"params" gorm:"type:varchar(2000);column:params;default:'{}'"`
	EnabledPull    bool                                     `json:"enabledPull" gorm:"type:bool;column:enabled_pull;default:false"`
	LastPullMillTS int64                                    `json:"lastPullMillTS" gorm:"type:bigint;column:last_pull_mill_ts;default:0"`
}

func (AlertSource) TableName() string {
	return "alert_sources"
}

type AlertSourceParams map[string]any

type AlertSource2Cluster struct {
	SourceID  string `gorm:"type:varchar(255);column:source_id"`
	ClusterID string `gorm:"type:varchar(255);column:cluster_id"`
}

type SourceFrom struct {
	SourceID string `form:"sourceId" json:"sourceId" gorm:"primaryKey;type:varchar(255);column:source_id"`
	SourceInfo
}

type SourceInfo struct {
	SourceName string `form:"sourceName" json:"sourceName" gorm:"unique;type:varchar(255);column:source_name"`
	SourceType string `form:"sourceType" json:"sourceType" gorm:"type:varchar(255);column:source_type"`
}
