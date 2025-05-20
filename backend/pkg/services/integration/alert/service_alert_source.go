// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"errors"
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/enrich"
	"go.uber.org/multierr"
	"gorm.io/gorm"
)

func (s *service) GetAlertSource(ctx core.Context, source *alert.SourceFrom) (*alert.AlertSource, error) {
	// TODO support search by sourceName
	alertSource, err := s.dbRepo.GetAlertSource(ctx, source.SourceID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, alert.ErrAlertSourceNotExist{}
	}
	return alertSource, err
}

func (s *service) UpdateAlertSource(ctx core.Context, source *alert.AlertSource) (*alert.AlertSource, error) {
	if len(source.SourceID) <= 0 {
		return nil, fmt.Errorf("must use sourceId to specify the data source")
	}

	if e, find := s.dispatcher.SourceName2EnricherMap.Load(source.SourceName); find {
		if enricher, ok := e.(*enrich.AlertEnricher); ok && enricher.SourceID != source.SourceID {
			return nil, alert.ErrAlertSourceAlreadyExist{
				Name: enricher.SourceName,
			}
		}
	} else if _, find := s.dispatcher.EnricherMap.Load(source.SourceID); !find {
		return nil, alert.ErrAlertSourceNotExist{}
	}

	err := s.dbRepo.UpdateAlertSource(ctx, source)
	return source, err
}

func (s *service) ListAlertSource(ctx core.Context) ([]alert.AlertSource, error) {
	return s.dbRepo.ListAlertSource(ctx)
}

func (s *service) DeleteAlertSource(ctx core.Context, source alert.SourceFrom) (*alert.AlertSource, error) {
	deletedSource, err := s.dbRepo.DeleteAlertSource(ctx, source)
	if err != nil {
		return nil, err
	}

	s.dispatcher.DeleteAlertSource(ctx, deletedSource)

	var storeError error
	err = s.dbRepo.DeleteAlertEnrichRuleBySourceId(ctx, source.SourceID)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichConditionsBySourceId(ctx, source.SourceID)
	storeError = multierr.Append(storeError, err)
	err = s.dbRepo.DeleteAlertEnrichSchemaTargetBySourceId(ctx, source.SourceID)
	storeError = multierr.Append(storeError, err)

	return deletedSource, storeError
}
