// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/google/uuid"
	"go.uber.org/multierr"
)

var (
	uuidZero          = uuid.UUID{}
	defaultSourceName = "APO_DEFAULT_ENRICH_RULE"
	commonSourceType  = "json"
)

func (s *service) ClearDefaultAlertEnrichRule(ctx core.Context, sourceType string) (bool, error) {
	_, find := s.defaultEnrichRules.LoadAndDelete(sourceType)

	sourceUUID := uuid.NewMD5(uuidZero, []byte(sourceType)).String()

	var storeError error
	err := s.dbRepo.DeleteAlertEnrichRuleBySourceId(ctx, sourceUUID)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichConditionsBySourceId(ctx, sourceUUID)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichSchemaTargetBySourceId(ctx, sourceUUID)
	storeError = multierr.Append(storeError, err)
	return find, storeError
}

func (s *service) GetDefaultAlertEnrichRule(ctx core.Context, sourceType string) (string, []alert.AlertEnrichRuleVO) {
	rules, find := s.defaultEnrichRules.Load(sourceType)
	if find {
		return sourceType, rules.([]alert.AlertEnrichRuleVO)
	}

	rules, find = s.defaultEnrichRules.Load(commonSourceType)
	if find {
		return commonSourceType, rules.([]alert.AlertEnrichRuleVO)
	}

	return "", []alert.AlertEnrichRuleVO{}
}

func (s *service) SetDefaultAlertEnrichRule(ctx core.Context, sourceType string, tagEnrichRules []alert.AlertEnrichRuleVO) error {
	existed, err := s.ClearDefaultAlertEnrichRule(ctx, sourceType)
	if err != nil {
		return err
	}

	sourceUUID := uuid.NewMD5(uuidZero, []byte(sourceType)).String()
	sourceFrom := &alert.SourceFrom{
		SourceID: sourceUUID,
		SourceInfo: alert.SourceInfo{
			SourceName: fmt.Sprintf("%s_%s", defaultSourceName, strings.ToUpper(sourceType)),
			SourceType: sourceType,
		},
	}
	rules, newR, newC, newS := s.prepareAlertEnrichRule(sourceFrom, tagEnrichRules)
	s.defaultEnrichRules.Store(sourceType, rules)

	var storeError error
	if !existed {
		err := s.dbRepo.CreateAlertSource(ctx, &alert.AlertSource{SourceFrom: *sourceFrom})
		storeError = multierr.Append(storeError, err)
	}

	err = s.dbRepo.AddAlertEnrichRule(ctx, newR)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichConditions(ctx, newC)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichSchemaTarget(ctx, newS)
	storeError = multierr.Append(storeError, err)
	return storeError
}
