// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type AlertEnrichRule struct {
	EnrichRuleID string `json:"enrichRuleId" gorm:"primaryKey;type:varchar(100);column:enrich_rule_id"` // 关联规则ID

	SourceID  string `json:"sourceId" gorm:"type:varchar(100);column:source_id"` // 告警源ID
	RType     string `json:"rType" gorm:"type:varchar(100);column:r_type"`       // 规则类型
	RuleOrder int    `json:"-" gorm:"type:int(10);column:rule_order"`            // 规则顺序

	FromField string `json:"fromField" gorm:"type:varchar(100);column:from_field"` // 来源字段 (支持拼接)
	FromRegex string `json:"fromRegex" gorm:"type:varchar(100);column:from_regex"` // 从来源字段中正则截取

	// ---------------- tagMapping -----------------
	TargetTagId int    `json:"targetTagId" gorm:"type:int(10);column:target_tag_id"` // 目标TagID
	CustomTag   string `json:"customTag" gorm:"type:varchar(100);column:custom_tag"` // 自定义Tag字段

	// ---------------- schemaMapping --------------
	Schema       string `json:"schema,omitempty" gorm:"type:varchar(100);column:schema"`              // 匹配的映射结构
	SchemaSource string `json:"schemaSource,omitempty" gorm:"type:varchar(100);column:schema_source"` // 匹配的映射结构的源字段
}

type AlertEnrichCondition struct {
	EnrichRuleID string `json:"-" gorm:"type:varchar(100);column:enrich_rule_id;index"` // 关联规则ID
	SourceID     string `json:"-" gorm:"type:varchar(100);column:source_id;index"`      // 告警源ID

	FromField string `json:"fromField" gorm:"type:varchar(100);column:from_field"` // 来源字段
	Operation string `json:"operation" gorm:"type:varchar(100);column:operation"`  // 比较方式 match,not match,gt,lt,ge,le,eq
	Expr      string `json:"expr" gorm:"type:varchar(100);column:expr"`            // 比较表达式
}

type AlertEnrichSchemaTarget struct {
	SourceID     string `json:"-" gorm:"type:varchar(100);column:source_id;index"`      // 告警源ID
	EnrichRuleID string `json:"-" gorm:"type:varchar(100);column:enrich_rule_id;index"` // 关联规则ID

	SchemaField string `json:"schemaField" gorm:"type:varchar(100);column:schema_field"` // 来自于映射表的指定字段
	TargetTagID int    `json:"targetTagId" gorm:"type:int(10);column:target_tag_id"`     // 目标TagID
	CustomTag   string `json:"customTag" gorm:"type:varchar(100);column:custom_tag"`     // 自定义Tag字段
}

type AlertEnrichRuleVO struct {
	AlertEnrichRule

	// --------------- conditions ----------------
	Conditions []AlertEnrichCondition `json:"conditions" gorm:"type:varchar(100);column:conditions"`

	// --------------- schemaMapping -------------
	SchemaTargets []AlertEnrichSchemaTarget `json:"schemaTargets" gorm:"type:varchar(100);column:schema_targets"` // 目标映射字段
}
