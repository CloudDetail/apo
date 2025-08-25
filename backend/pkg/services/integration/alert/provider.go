// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

var providerMap = map[string]func(sourceFrom alert.SourceFrom, params alert.AlertSourceParams) provider.Provider{
	"datadog": provider.NewDatadogProvider,
}

func (s *service) KeepPullAlert(ctx core.Context, source alert.AlertSource, interval time.Duration, p provider.Provider) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	now := time.Now()

	lastPullTime := time.UnixMilli(source.LastPullMillTS)
	if lastPullTime.Add(15 * 24 * time.Hour).Before(now) {
		lastPullTime = now.Add(-15 * 24 * time.Hour)
	}

	for range ticker.C {
		now := time.Now()
		if now.Sub(lastPullTime) < interval {
			continue
		}
		events, err := p.GetAlerts(nil)
		if err != nil {
			continue
		}

		err = s.dispatcher.DispatchDecodedEvents(&source.SourceFrom, events)
		if err != nil {
			continue
		}

		lastPullTime = now
		s.difyRepo.SubmitAlertEvents(events)
		err = s.ckRepo.InsertAlertEvent(ctx, events, source.SourceFrom)
		if err == nil {
			s.dbRepo.UpdateAlertSourceLastPullTime(ctx, source.SourceID, lastPullTime)
		}
	}
}
