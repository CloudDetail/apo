// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var (
	uuidZero		= uuid.UUID{}
	defaultSourceName	= "APO_DEFAULT_ENRICH_RULE"
	commonSourceType	= "json"
)

func (s *service) ClearDefaultAlertEnrichRule(ctx_core core.Context, sourceType string) (bool, error) {
	_, find := s.defaultEnrichRules.LoadAndDelete(sourceType)

	sourceUUID := uuid.NewMD5(uuidZero, []byte(sourceType)).String()

	var storeError error
	err := s.dbRepo.DeleteAlertEnrichRuleBySourceId(ctx_core, sourceUUID)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichConditionsBySourceId(ctx_core, sourceUUID)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichSchemaTargetBySourceId(ctx_core, sourceUUID)
	storeError = multierr.Append(storeError, err)
	return find, storeError
}

func (s *service) GetDefaultAlertEnrichRule(ctx_core core.Context, sourceType string) (string, []alert.AlertEnrichRuleVO) {
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

func (s *service) SetDefaultAlertEnrichRule(ctx_core core.Context, sourceType string, tagEnrichRules []alert.AlertEnrichRuleVO) error {
	existed, err := s.ClearDefaultAlertEnrichRule(ctx_core, sourceType)
	if err != nil {
		return err
	}

	sourceUUID := uuid.NewMD5(uuidZero, []byte(sourceType)).String()
	sourceFrom := &alert.SourceFrom{
		SourceID:	sourceUUID,
		SourceInfo: alert.SourceInfo{
			SourceName:	fmt.Sprintf("%s_%s", defaultSourceName, strings.ToUpper(sourceType)),
			SourceType:	sourceType,
		},
	}
	rules, newR, newC, newS := s.prepareAlertEnrichRule(sourceFrom, tagEnrichRules)
	s.defaultEnrichRules.Store(sourceType, rules)

	var storeError error
	if !existed {
		err := s.dbRepo.CreateAlertSource(ctx_core, &alert.AlertSource{SourceFrom: *sourceFrom})
		storeError = multierr.Append(storeError, err)
	}

	err = s.dbRepo.AddAlertEnrichRule(ctx_core, newR)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichConditions(ctx_core, newC)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.AddAlertEnrichSchemaTarget(ctx_core, newS)
	storeError = multierr.Append(storeError, err)
	return storeError
}
