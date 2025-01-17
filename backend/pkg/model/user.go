// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

// RoleFilter These fields can not use at the same time.
type RoleFilter struct {
	Names []string
	Name  string
	IDs   []int
	ID    int
}

// DataGroupFilter These fields can not use at the same time.
type DataGroupFilter struct {
	Names          []string
	Name           string
	IDs            []int64
	ID             int64
	DatasourceList []Datasource

	CurrentPage *int
	PageSize    *int
}

type Datasource struct {
	Datasource string   `json:"datasource"`       // namespaceName or serviceName
	Type       string   `json:"type,omitempty"`   // namespace or service
	Category   string   `json:"category"`         // normal or apm
	Nested     []string `json:"nested,omitempty"` // Nested datasource (namespace service belongs to or service under namespace)
}
