// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

type ListClusterResponse struct {
	Clusters []ClusterListItem `json:"clusters"`
}

type ClusterListItem struct {
	Cluster
	IntegrationConfigured bool `json:"integrationConfigured"`
}

type GetCInstallDocResponse struct {
	InstallMD string `json:"installMd"`
}

type GetCInstallConfigResponse struct {
	FileName string `json:"fileName"`
	Content  []byte `json:"content"`
}
