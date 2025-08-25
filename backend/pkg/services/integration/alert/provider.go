// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"log"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

func (s *service) GetAlertProviderParamsSpec(sourceType string) *response.GetAlertProviderParamsSpecResponse {
	pType, find := provider.ProviderRegistry[sourceType]
	if !find {
		return &response.GetAlertProviderParamsSpecResponse{
			ParamSpec: &provider.ParamSpec{
				Name: "root",
				Type: provider.JSONTypeObject,
			},
			WithPullOptions: false,
		}
	}
	return &response.GetAlertProviderParamsSpecResponse{ParamSpec: &pType.ParamSpec}
}

func (s *service) KeepPullAlert(ctx core.Context, source alert.AlertSource, interval time.Duration, p provider.Provider) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	now := time.Now()

	lastPullTime := time.UnixMilli(source.LastPullMillTS)
	if lastPullTime.Add(15 * 24 * time.Hour).Before(now) {
		lastPullTime = now.Add(-15 * 24 * time.Hour)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if now.Sub(lastPullTime) < interval {
				continue
			}
			events, err := p.PullAlerts(provider.GetAlertParams{
				From: lastPullTime,
				To:   now,
			})
			if err != nil {
				log.Printf("failed to pull alerts,err: %v", err)
				continue
			}

			err = s.dispatcher.DispatchDecodedEvents(&source.SourceFrom, events)
			if err != nil {
				log.Printf("failed to dispatch events,err: %v", err)
				continue
			}

			lastPullTime = now
			s.difyRepo.SubmitAlertEvents(events)

			if err = s.ckRepo.InsertAlertEvent(ctx, events, source.SourceFrom); err != nil {
				log.Printf("failed to insert alert event,err: %v", err)
				continue
			}
			if err = s.dbRepo.UpdateAlertSourceLastPullTime(ctx, source.SourceID, lastPullTime); err != nil {
				log.Printf("failed to update alert source last pull time,err: %v", err)
			}
		}
	}
}
