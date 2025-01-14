// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	alertin "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/input/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/input/alert/enrich"
	"github.com/google/uuid"
	"go.uber.org/multierr"
)

func (s *service) CreateAlertSource(source *alertin.AlertSource) (*alertin.AlertSource, error) {
	_, find := s.dispatcher.SourceName2EnricherMap.Load(source.SourceName)
	if find {
		return nil, alertin.ErrAlertSourceAlreadyExist{
			Name: source.SourceName,
		}
	}

	source.SourceID = uuid.NewString()
	err := s.dbRepo.CreateAlertSource(source)
	if err != nil {
		return nil, err
	}

	_, err = s.initDefaultAlertSource(&source.SourceFrom)

	return source, err
}

// create default enrich for specific alertSource
// always return a vaild enricher
// - If already exists, return existing enricher
// - If the default rule is wrong, return empty enricher
func (s *service) initDefaultAlertSource(source *alertin.SourceFrom) (*enrich.AlertEnricher, error) {
	s.AddAlertSourceLock.Lock()
	defer s.AddAlertSourceLock.Unlock()

	// Double Check before create alertSource
	if len(source.SourceID) == 0 {
		if enricher, find := s.dispatcher.SourceName2EnricherMap.Load(source.SourceName); find {
			return enricher.(*enrich.AlertEnricher), alertin.ErrAlertSourceAlreadyExist{
				Name: source.SourceName,
			}
		}
		source.SourceID = uuid.NewString()
	} else if enricher, find := s.dispatcher.EnricherMap.Load(source.SourceID); find {
		return enricher.(*enrich.AlertEnricher), alertin.ErrAlertSourceAlreadyExist{}
	}

	_, defaultRules := s.GetDefaultAlertEnrichRule(source.SourceType)
	storedRules, newR, newC, newS := s.prepareAlertEnrichRule(source, defaultRules)

	enricher, err := s.createAlertSource(source, storedRules)
	if err != nil {
		// return empty enricher
		enricher = &enrich.AlertEnricher{
			SourceFrom: source,
			Enrichers:  []enrich.Enricher{},
		}
		err = alertin.ErrIllegalAlertRule{Err: err}
		s.dispatcher.AddAlertSource(source, enricher)
		return enricher, err
	}

	var storeError error
	err = s.dbRepo.AddAlertEnrichRule(newR)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichConditions(newC)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichSchemaTarget(newS)
	storeError = multierr.Append(storeError, err)

	s.dispatcher.AddAlertSource(source, enricher)

	return enricher, storeError
}

func (s *service) createAlertSource(
	source *alertin.SourceFrom,
	enrichRules []alertin.AlertEnrichRuleVO,
) (*enrich.AlertEnricher, error) {
	if len(source.SourceID) == 0 {
		source.SourceID = uuid.New().String()
	}

	enricher := &enrich.AlertEnricher{
		SourceFrom: source,
		Enrichers:  make([]enrich.Enricher, 0, len(enrichRules)),
	}

	for i := 0; i < len(enrichRules); i++ {
		enrichRules[i].SourceID = source.SourceID
		ruleID := enrichRules[i].EnrichRuleID
		if len(ruleID) == 0 {
			// generate ruleID for default rule or new Rule
			enrichRules[i].EnrichRuleID = uuid.NewString()
			ruleID = enrichRules[i].EnrichRuleID
		}

		// check if schemaTarget is legal
		if enrichRules[i].RType == "schemaMapping" {
			if !alert.AllowSchema.MatchString(enrichRules[i].Schema) {
				return nil, alertin.ErrNotAllowSchema{Table: enrichRules[i].Schema}
			}
			if !alert.AllowSchema.MatchString(enrichRules[i].SchemaSource) {
				return nil, alertin.ErrNotAllowSchema{Column: enrichRules[i].SchemaSource}
			}
		}

		for s := 0; s < len(enrichRules[i].Conditions); s++ {
			enrichRules[i].Conditions[s].EnrichRuleID = ruleID
			enrichRules[i].Conditions[s].SourceID = source.SourceID
		}

		for s := 0; s < len(enrichRules[i].SchemaTargets); s++ {
			if !alert.AllowSchema.MatchString(enrichRules[i].SchemaTargets[s].SchemaField) {
				return nil, alertin.ErrNotAllowSchema{Column: enrichRules[i].SchemaTargets[s].SchemaField}
			}

			enrichRules[i].SchemaTargets[s].EnrichRuleID = ruleID
			enrichRules[i].SchemaTargets[s].SourceID = source.SourceID
		}

		tagEnricher, err := enrich.NewTagEnricher(enrichRules[i], s.dbRepo, i)
		if err != nil {
			return nil, err
		}
		enricher.Enrichers = append(enricher.Enrichers, tagEnricher)
	}

	return enricher, nil
}

// load existed enricher from db when process initializing
func (s *service) initExistedAlertSource(source *alertin.SourceFrom, enrichRules []alertin.AlertEnrichRuleVO) (*enrich.AlertEnricher, error) {
	enricher := &enrich.AlertEnricher{
		SourceFrom: source,
		Enrichers:  make([]enrich.Enricher, 0, len(enrichRules)),
	}
	for idx, rule := range enrichRules {
		tagEnricher, err := enrich.NewTagEnricher(rule, s.dbRepo, idx)
		if err != nil {
			return nil, err
		}
		enricher.Enrichers = append(enricher.Enrichers, tagEnricher)
	}
	return enricher, nil
}
