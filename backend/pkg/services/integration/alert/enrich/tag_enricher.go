// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package enrich

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/itchyny/gojq"
)

type TagEnricher struct {
	ID    string
	Order int

	RType string

	*JQParser

	DBRepo      database.Repo
	FromRegex   string
	fromPattern *regexp.Regexp

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
		FromRegex:    enrichRule.FromRegex,
		TargetTagId:  enrichRule.TargetTagId,
		CustomTag:    enrichRule.CustomTag,
		Schema:       enrichRule.Schema,
		SchemaSource: enrichRule.SchemaSource,
		SchemaTarget: enrichRule.SchemaTargets,
	}

	if enrichRule.FromRegex != "" {
		fromPattern, err := regexp.Compile(enrichRule.FromRegex)
		if err == nil {
			tagEnricher.fromPattern = fromPattern
		}
	}

	targetTags, err := dbRepo.ListAlertTargetTags(core.EmptyCtx())
	if err != nil {
		return nil, err
	}
	if enrichRule.RType == "tagMapping" {
		if enrichRule.TargetTagId == 0 {
			tagEnricher.targetTag = enrichRule.CustomTag
		} else {
			tagEnricher.targetTag = getTargetTag(targetTags, enrichRule.TargetTagId)
		}
	} else if enrichRule.RType == "schemaMapping" {
		tagEnricher.TargetTags = targetTags
	}

	return tagEnricher, nil
}

func getTargetTag(tags []alert.TargetTag, id int) string {
	for _, tag := range tags {
		if tag.ID == uint(id) {
			return tag.Field
		}
	}
	return "undefined"
}

func (e *TagEnricher) Enrich(alertEvent *alert.AlertEvent) {
	iter := e.JQParser.Run(map[string]any(alertEvent.Tags))
	v, ok := iter.Next()
	if !ok || v == nil {
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

	if e.fromPattern != nil {
		vs := e.fromPattern.FindStringSubmatch(value)
		if len(vs) > 1 {
			value = vs[1] // Extract first capture group
		} else if len(vs) == 1 {
			value = vs[0] // Extract entire match (no capture group)
		} else {
			value = "" // No match found
		}
	} // else: retain original value if e.fromPattern is nil

	switch e.RType {
	case "tagMapping":
		alertEvent.EnrichTags[e.targetTag] = value
	case "schemaMapping":
		targets, err := e.DBRepo.SearchSchemaTarget(core.EmptyCtx(), e.Schema, e.SchemaSource, value, e.SchemaTarget)
		if err != nil {
			return
		}

		for idx, schemaTarget := range e.SchemaTarget {
			if schemaTarget.TargetTagID == 0 {
				alertEvent.EnrichTags[schemaTarget.CustomTag] = targets[idx]
			} else if len(e.TargetTags) > schemaTarget.TargetTagID {
				field := getTargetTag(e.TargetTags, schemaTarget.TargetTagID)
				alertEvent.EnrichTags[field] = targets[idx]
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
	FromJQExpression string // JQ expression composed of condition and fromField
	*gojq.Query
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
		jqExpression = fmt.Sprintf(`if %s then %s else null end`, strings.Join(conditions, " and "), enrichRule.FromField)
	} else {
		jqExpression = fmt.Sprintf(` %s `, enrichRule.FromField)
	}

	jqParser, err := gojq.Parse(jqExpression)
	if err != nil {
		return nil, err
	}

	return &JQParser{
		FromJQExpression: jqExpression,
		Query:            jqParser,
	}, nil
}
