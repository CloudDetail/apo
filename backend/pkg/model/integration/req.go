// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

type GetCInstallRequest struct {
	ClusterID string `json:"clusterId" form:"clusterId"`
}

type TriggerAdapterUpdateRequest struct {
	LastUpdateTS int64 `form:"lastUpdateTS" json:"lastUpdateTS"`
}
