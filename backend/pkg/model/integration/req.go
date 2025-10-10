// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

type GetCInstallRequest struct {
	ClusterID string `json:"clusterId" form:"clusterId"`
}

type TriggerAdapterUpdateRequest struct {
	LastUpdateTS int64 `form:"lastUpdateTS" json:"lastUpdateTS"`
}

type GetIntegrationVarRequest struct {
	Variable string `uri:"variable" binding:"required,max=64" `
}
