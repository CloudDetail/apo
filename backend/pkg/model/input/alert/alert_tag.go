// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/google/uuid"
)

type TagMapping struct {
	SourceID    string
	TargetTagID uint

	CustomField string `gorm:"type:varchar(100);column:custom_field"`

	// TODO 增加条件筛选
	FromType  string `gorm:"type:varchar(100);column:from_type"`  // only support jsonpath now
	FromField string `gorm:"type:varchar(255);column:from_field"` // json格式, 用于支持多来源的情况
	RegexExpr string `gorm:"type:varchar(255);column:regex_expr"` // 正则表达式

	MappingOrder uint `gorm:"type:uint(10);default:10;column:mapping_order"` // 匹配优先级, 默认为10
}

type TargetTag struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TagName  string `gorm:"type:varchar(100);column:tag_name" json:"tagName"`
	Describe string `gorm:"type:varchar(255);column:describe" json:"describe"`
	Field    string `gorm:"type:varchar(100);column:field" json:"targetTag"`
}

type TagMappingVO struct {
	SourceID   uuid.UUID `gorm:"column:source_id"`
	SourceName string    `gorm:"type:varchar(100);column:source_name"`
	SourceType string    `gorm:"type:varchar(100);column:source_type"`

	TargetTagID uint   `gorm:"column:target_tag_id"`
	TagName     string `gorm:"type:varchar(100);column:tag_name"`
	Describe    string `gorm:"type:varchar(255);column:describe"`
	Field       string `gorm:"type:varchar(100);column:field"`
	CustomField string `gorm:"type:varchar(100);column:custom_field"`
	RegexExpr   string `gorm:"type:varchar(255);column:regex_expr"` // 正则表达式

	FromType     string `gorm:"type:varchar(100);column:from_type"`            // only support jsonpath now
	FromField    string `gorm:"type:varchar(255);column:from_field"`           // json格式, 用于支持多来源的情况
	MappingOrder uint   `gorm:"type:uint(10);default:10;column:mapping_order"` // 匹配优先级, 默认为10
}

type DefaultTagMapping struct {
	SourceType   string `gorm:"type:varchar(100);column:source_type"`
	TargetTagID  uint   `gorm:"column:target_tag_id"`
	FromType     string `gorm:"type:varchar(100);column:from_type"`            // only support jsonpath now
	FromField    string `gorm:"type:varchar(255);column:from_field"`           // json格式, 用于支持多来源的情况
	RegexExpr    string `gorm:"type:varchar(255);column:regex_expr"`           // 正则表达式取值
	MappingOrder uint   `gorm:"type:uint(10);default:10;column:mapping_order"` // 匹配优先级, 默认为10
	Describe     string `gorm:"type:varchar(1000);column:describe"`            // 对于默认匹配规则的使用场景的描述
}
