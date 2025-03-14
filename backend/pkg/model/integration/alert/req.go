// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type AlertEnrichRuleConfigRequest struct {
	SourceId          string              `json:"sourceId"`
	EnrichRuleConfigs []AlertEnrichRuleVO `json:"enrichRuleConfigs"`

	SetAsDefault bool `json:"setAsDefault,omitempty"`
}

type AlertSchemaRequest struct {
	Schema string `json:"schema" form:"schema"`
}

type CreateSchemaRequest struct {
	Schema  string   `json:"schema"`
	Columns []string `json:"columns"`

	FullRows [][]string `json:"fullRows"`
	// Rows     []map[string]string `json:"rows"`
}

type UpdateSchemaDataRequest struct {
	Schema  string   `json:"schema"`
	Columns []string `json:"columns"`

	ClearAll bool `json:"clearAll"`

	NewRows    [][]string       `json:"newRows"`
	UpdateRows map[int][]string `json:"updateRows"`
	// UpdateCells map[int]map[string]string `json:"updateCells"`
}

type DefaultAlertEnrichRuleRequest struct {
	SourceType string `json:"sourceType" form:"sourceType"`
}

type SetDefaultAlertEnrichRuleRequest struct {
	SourceType        string              `json:"sourceType"`
	EnrichRuleConfigs []AlertEnrichRuleVO `json:"enrichRuleConfigs"`
}
