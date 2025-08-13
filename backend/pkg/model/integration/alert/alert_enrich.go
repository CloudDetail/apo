// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type AlertEnrichRule struct {
	EnrichRuleID string `json:"enrichRuleId" gorm:"primaryKey;type:varchar(255);column:enrich_rule_id"`

	SourceID  string `json:"sourceId" gorm:"type:varchar(255);column:source_id"`
	RType     string `json:"rType" gorm:"type:varchar(255);column:r_type"`
	RuleOrder int    `json:"-" gorm:"type:int;column:rule_order"`

	FromField string `json:"fromField" gorm:"type:varchar(255);column:from_field"`
	FromRegex string `json:"fromRegex" gorm:"type:varchar(255);column:from_regex"`

	// ---------------- tagMapping -----------------
	TargetTagId int    `json:"targetTagId" gorm:"type:int;column:target_tag_id"`
	CustomTag   string `json:"customTag" gorm:"type:varchar(255);column:custom_tag"`

	// ---------------- schemaMapping --------------
	Schema       string `json:"schema,omitempty" gorm:"type:varchar(255);column:schema"`
	SchemaSource string `json:"schemaSource,omitempty" gorm:"type:varchar(255);column:schema_source"`
}

func (AlertEnrichRule) TableName() string {
	return "alert_enrich_rules"
}

type AlertEnrichCondition struct {
	EnrichRuleID string `json:"-" gorm:"type:varchar(255);column:enrich_rule_id;index"`
	SourceID     string `json:"-" gorm:"type:varchar(255);column:source_id;index"`

	FromField string `json:"fromField" gorm:"type:varchar(255);column:from_field"`
	Operation string `json:"operation" gorm:"type:varchar(255);column:operation"` // support match,not match,gt,lt,ge,le,eq
	Expr      string `json:"expr" gorm:"type:varchar(255);column:expr"`
}

type AlertEnrichSchemaTarget struct {
	SourceID     string `json:"-" gorm:"type:varchar(255);column:source_id;index"`
	EnrichRuleID string `json:"-" gorm:"type:varchar(255);column:enrich_rule_id;index"`

	SchemaField string `json:"schemaField" gorm:"type:varchar(255);column:schema_field"`
	TargetTagID int    `json:"targetTagId" gorm:"type:int;column:target_tag_id"`
	CustomTag   string `json:"customTag" gorm:"type:varchar(255);column:custom_tag"`
}

type AlertEnrichRuleVO struct {
	AlertEnrichRule

	// --------------- conditions ----------------
	Conditions []AlertEnrichCondition `json:"conditions" gorm:"type:varchar(255);column:conditions"`

	// --------------- schemaMapping -------------
	SchemaTargets []AlertEnrichSchemaTarget `json:"schemaTargets" gorm:"type:varchar(255);column:schema_targets"`
}
