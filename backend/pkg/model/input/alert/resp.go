// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type GetTargetTagsResponse struct {
	TargetTags []TargetTag `json:"targetTags"`
}

type ListAlertSourceResponse struct {
	AlertSources []AlertSource `json:"alertSources"`
}

type GetSchemaColumnsResponse struct {
	Columns []string `json:"columns"`
}

type ListClusterResponse struct {
	Clusters []Cluster `json:"clusters"`
}

type ListSchemaResponse struct {
	Schemas []string `json:"schemas"`
}

type ListSchemaWithColumnsResponse struct {
	Schemas map[string][]string `json:"schemas"`
}

type CheckSchemaIsUsedReponse struct {
	IsUsing          bool     `json:"isUsing"`
	AlertSourceNames []string `json:"alertSourceNames"`
}

type GetSchemaDataReponse struct {
	Columns []string           `json:"columns"`
	Rows    map[int64][]string `json:"rows"`
}

type GetAlertEnrichRuleResponse struct {
	SourceId          string              `json:"sourceId"`
	EnrichRuleConfigs []AlertEnrichRuleVO `json:"enrichRuleConfigs"`
}

type DefaultAlertEnrichRuleResponse struct {
	SourceType        string              `json:"sourceType"`
	EnrichRuleConfigs []AlertEnrichRuleVO `json:"enrichRuleConfigs"`
}
