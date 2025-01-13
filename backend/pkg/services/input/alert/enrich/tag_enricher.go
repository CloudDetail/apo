// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package enrich

import (
	"fmt"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/itchyny/gojq"
)

type TagEnricher struct {
	ID    string
	Order int

	RType string

	*JQParser

	DBRepo database.Repo

	// ---------------- tagMapping ----------------
	TargetTagId int
	CustomTag   string

	targetTag string // TODO query from cache

	// ---------------- schemaMapping ----------------
	Schema       string
	SchemaSource string
	SchemaTarget []alert.AlertEnrichSchemaTarget

	TargetTags []alert.TargetTag
}

func NewTagEnricher(
	enrichRule alert.AlertEnrichRuleVO,
	dbRepo database.Repo,
	Order int,
) (*TagEnricher, error) {
	jqEnricher, err := newJQEnricher(enrichRule)
	if err != nil {
		return nil, err
	}

	tagEnricher := &TagEnricher{
		JQParser:     jqEnricher,
		ID:           enrichRule.EnrichRuleID,
		Order:        Order,
		RType:        enrichRule.RType,
		TargetTagId:  enrichRule.TargetTagId,
		CustomTag:    enrichRule.CustomTag,
		Schema:       enrichRule.Schema,
		SchemaSource: enrichRule.SchemaSource,
		SchemaTarget: enrichRule.SchemaTargets,
	}

	targetTags, err := dbRepo.ListAlertTargetTags()
	if err != nil {
		return nil, err
	}
	if enrichRule.RType == "tagMapping" {
		if enrichRule.TargetTagId == 0 {
			tagEnricher.targetTag = enrichRule.CustomTag
		} else if len(targetTags) > enrichRule.TargetTagId {
			tagEnricher.targetTag = targetTags[enrichRule.TargetTagId].Field
		}
	} else if enrichRule.RType == "schemaMapping" {
		tagEnricher.TargetTags = targetTags
	}

	return tagEnricher, nil
}

func (e *TagEnricher) Enrich(alertEvent *alert.AlertEvent) {
	iter := e.JQParser.JQParser.Run(alertEvent.RawTags)
	v, ok := iter.Next()
	if !ok {
		return
	}

	var value string
	if vStr, ok := v.(string); ok {
		value = vStr
	} else if vStrs, ok := v.([]string); ok && len(vStrs) > 0 {
		value = vStrs[0]
	} else {
		return
	}

	switch e.RType {
	case "tagMapping":
		alertEvent.Tags[e.targetTag] = value
	case "schemaMapping":
		targets, err := e.DBRepo.SearchSchemaTarget(e.Schema, e.SchemaSource, value, e.SchemaTarget)
		if err != nil {
			return
		}

		for idx, schemaTarget := range e.SchemaTarget {
			if schemaTarget.TargetTagID == 0 {
				alertEvent.Tags[schemaTarget.CustomTag] = targets[idx]
			} else if len(e.TargetTags) > schemaTarget.TargetTagID {
				alertEvent.Tags[e.TargetTags[schemaTarget.TargetTagID].Field] = targets[idx]
			}
		}
	}
}

func (e *TagEnricher) RuleID() string {
	return e.ID
}

func (e *TagEnricher) RuleOrder() int {
	return e.Order
}

type JQParser struct {
	FromJQExpression string      // 条件和提取配置组成的JQ表达式
	JQParser         *gojq.Query // 预解析的JQ表达式
}

func newJQEnricher(enrichRule alert.AlertEnrichRuleVO) (*JQParser, error) {
	var conditions []string
	for _, condition := range enrichRule.Conditions {
		var jqCondition string
		switch condition.Operation {
		case "match":
			jqCondition = fmt.Sprintf(`( %s | test("%s"))`, condition.FromField, condition.Expr)
		case "notMatch":
			jqCondition = fmt.Sprintf(`( %s | test("%s") | not)`, condition.FromField, condition.Expr)
		case "le":
			jqCondition = fmt.Sprintf(`( %s | . <= %s)`, condition.FromField, condition.Expr)
		case "ge":
			jqCondition = fmt.Sprintf(`( %s | . >= %s)`, condition.FromField, condition.Expr)
		case "lt":
			jqCondition = fmt.Sprintf(`( %s | . < %s)`, condition.FromField, condition.Expr)
		case "gt":
			jqCondition = fmt.Sprintf(`( %s | . > %s)`, condition.FromField, condition.Expr)
		}
		conditions = append(conditions, jqCondition)
	}

	var jqExpression string
	if len(conditions) > 0 {
		jqExpression = fmt.Sprintf(`if %s then %s else "" end`, strings.Join(conditions, " and "), enrichRule.FromField)
	} else {
		jqExpression = fmt.Sprintf(` %s `, enrichRule.FromField)
	}

	jqParser, err := gojq.Parse(jqExpression)
	if err != nil {
		return nil, err
	}

	return &JQParser{
		FromJQExpression: jqExpression,
		JQParser:         jqParser,
	}, nil
}
