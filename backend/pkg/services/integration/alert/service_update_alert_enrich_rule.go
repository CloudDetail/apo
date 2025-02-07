// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"sort"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/enrich"
	"github.com/google/uuid"
	"go.uber.org/multierr"
)

func (s *service) UpdateAlertEnrichRule(req *alert.AlerEnrichRuleConfigRequest) error {
	oldEnricherPtr, find := s.dispatcher.EnricherMap.Load(req.SourceId)
	if !find {
		return alert.ErrAlertSourceNotExist{}
	}

	oldEnricher := oldEnricherPtr.(*enrich.AlertEnricher)
	sourceFrom := &alert.SourceFrom{SourceID: req.SourceId}
	newTagEnricher, err := s.createAlertSource(sourceFrom, req.EnrichRuleConfigs)
	if err != nil {

		return err
	}

	// deleted or modified Rule
	var deletedRules []string
	var conditionsModifiedRules []string
	var modifiedAlertEnrichRules []string
	var schemaTargetModifiedRules []string

	var newConditions []alert.AlertEnrichCondition
	var newSchemaTargets []alert.AlertEnrichSchemaTarget
	var newAlertEnrichRules []alert.AlertEnrichRule

	var existedRule []int
	for _, old := range oldEnricher.Enrichers {
		newEnricher, exist := newTagEnricher.EnricherMap.Load(old.RuleID())
		if !exist {
			deletedRules = append(deletedRules, old.RuleID())
			continue
		}

		oldE := old.(*enrich.TagEnricher)
		newE := newEnricher.(*enrich.TagEnricher)

		existedRule = append(existedRule, newE.Order)

		if oldE.JQParser.FromJQExpression != newE.JQParser.FromJQExpression {
			conditionsModifiedRules = append(conditionsModifiedRules, old.RuleID())

			newConditions = append(newConditions, req.EnrichRuleConfigs[newE.Order].Conditions...)
		}

		if oldE.Order != newE.Order ||
			oldE.FromRegex != newE.FromRegex ||
			oldE.TargetTagId != newE.Order ||
			oldE.CustomTag != newE.CustomTag ||
			oldE.Schema != newE.Schema ||
			oldE.SchemaSource != newE.SchemaSource {
			modifiedAlertEnrichRules = append(modifiedAlertEnrichRules, old.RuleID())
			newAlertEnrichRules = append(newAlertEnrichRules, req.EnrichRuleConfigs[newE.Order].AlertEnrichRule)
		}

		if !schemaTargetSliceEqual(oldE.SchemaTarget, newE.SchemaTarget) {
			schemaTargetModifiedRules = append(schemaTargetModifiedRules, old.RuleID())
		}
	}

	sort.Ints(existedRule)

	var x = 0
	for i := 0; i < len(newTagEnricher.Enrichers); i++ {
		if x < len(existedRule) && existedRule[x] == i {
			x++
			continue
		}

		newE := newTagEnricher.Enrichers[i].(*enrich.TagEnricher)
		newConditions = append(newConditions, req.EnrichRuleConfigs[newE.Order].Conditions...)
		newAlertEnrichRules = append(newAlertEnrichRules, alert.AlertEnrichRule{
			SourceID:     req.EnrichRuleConfigs[newE.Order].SourceID,
			RuleOrder:    newE.Order,
			EnrichRuleID: req.EnrichRuleConfigs[newE.Order].EnrichRuleID,
			RType:        req.EnrichRuleConfigs[newE.Order].RType,
			FromField:    req.EnrichRuleConfigs[newE.Order].FromField,
			FromRegex:    req.EnrichRuleConfigs[newE.Order].FromRegex,
			TargetTagId:  newE.TargetTagId,
			CustomTag:    newE.CustomTag,
			Schema:       newE.Schema,
			SchemaSource: newE.SchemaSource,
		})
		newSchemaTargets = append(newSchemaTargets, req.EnrichRuleConfigs[newE.Order].SchemaTargets...)
	}

	s.dispatcher.AddOrUpdateAlertSourceRule(alert.SourceFrom{SourceID: req.SourceId}, newTagEnricher)

	var storeError error
	if req.SetAsDefault {
		enricher, loaded := s.dispatcher.EnricherMap.Load(req.SourceId)
		if loaded {
			sourceType := enricher.(*enrich.AlertEnricher).SourceType
			s.SetDefaultAlertEnrichRule(sourceType, req.EnrichRuleConfigs)
		}
	}

	err = s.dbRepo.DeleteAlertEnrichRule(deletedRules)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichConditions(deletedRules)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichSchemaTarget(deletedRules)
	storeError = multierr.Append(storeError, err)

	err = s.dbRepo.DeleteAlertEnrichConditions(conditionsModifiedRules)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichConditions(newConditions)
	storeError = multierr.Append(storeError, err)

	err = s.dbRepo.DeleteAlertEnrichSchemaTarget(schemaTargetModifiedRules)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichSchemaTarget(newSchemaTargets)
	storeError = multierr.Append(storeError, err)

	err = s.dbRepo.DeleteAlertEnrichRule(modifiedAlertEnrichRules)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichRule(newAlertEnrichRules)
	storeError = multierr.Append(storeError, err)

	return storeError
}

func schemaTargetSliceEqual(a, b []alert.AlertEnrichSchemaTarget) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s *service) prepareAlertEnrichRule(
	sourceFrom *alert.SourceFrom,
	tagEnrichRules []alert.AlertEnrichRuleVO,
) ([]alert.AlertEnrichRuleVO, []alert.AlertEnrichRule,
	[]alert.AlertEnrichCondition, []alert.AlertEnrichSchemaTarget) {
	var newAlertRules []alert.AlertEnrichRule
	var newConditions []alert.AlertEnrichCondition
	var newSchemaTargets []alert.AlertEnrichSchemaTarget

	var storedRules []alert.AlertEnrichRuleVO

	for _, tagEnrichRule := range tagEnrichRules {
		newRule := tagEnrichRule.AlertEnrichRule
		newRule.SourceID = sourceFrom.SourceID

		ruleId := uuid.New().String()
		newRule.EnrichRuleID = ruleId
		newAlertRules = append(newAlertRules, newRule)

		conditions := make([]alert.AlertEnrichCondition, 0, len(tagEnrichRule.Conditions))
		for i := 0; i < len(tagEnrichRule.Conditions); i++ {
			newCondition := tagEnrichRule.Conditions[i]
			newCondition.SourceID = sourceFrom.SourceID
			newCondition.EnrichRuleID = ruleId
			conditions = append(conditions, newCondition)
			newConditions = append(newConditions, newCondition)
		}

		schemaTargets := make([]alert.AlertEnrichSchemaTarget, 0, len(tagEnrichRule.SchemaTargets))
		for i := 0; i < len(tagEnrichRule.SchemaTargets); i++ {
			newSchemaTarget := tagEnrichRule.SchemaTargets[i]
			newSchemaTarget.SourceID = sourceFrom.SourceID
			newSchemaTarget.EnrichRuleID = ruleId
			schemaTargets = append(schemaTargets, newSchemaTarget)
			newSchemaTargets = append(newSchemaTargets, newSchemaTarget)
		}

		storedRules = append(storedRules, alert.AlertEnrichRuleVO{
			AlertEnrichRule: newRule,
			Conditions:      conditions,
			SchemaTargets:   schemaTargets,
		})
	}

	return storedRules, newAlertRules, newConditions, newSchemaTargets
}
