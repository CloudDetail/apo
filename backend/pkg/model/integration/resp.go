// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

type ListClusterResponse struct {
	Clusters []Cluster `json:"clusters"`
}

type GetCInstallDocResponse struct {
	InstallMD []byte `json:"installMd"`
}

type GetCInstallConfigResponse struct {
	FileName string `json:"fileName"`
	Content  []byte `json:"content"`
}
