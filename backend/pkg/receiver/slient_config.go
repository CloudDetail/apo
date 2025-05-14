// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package receiver

import (
	"fmt"
	"time"

	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (r *InnerReceivers) GetSlienceConfigByAlertID(ctx_core core.Context, alertID string) (*sc.AlertSlienceConfig, error) {
	if cfgPtr, find := r.slientCFGMap.Load(alertID); find {
		return cfgPtr.(*sc.AlertSlienceConfig), nil
	}
	return nil, nil
}

func (r *InnerReceivers) ListSlienceConfig(ctx_core core.Context) ([]sc.AlertSlienceConfig, error) {
	return r.database.GetAlertSlience(ctx_core)
}

func (r *InnerReceivers) SetSlienceConfigByAlertID(ctx_core core.Context, alertID string, forDuration string) error {
	duration, err := time.ParseDuration(forDuration)
	if err != nil {
		return fmt.Errorf("duration is not valid: %w", err)
	}

	now := time.Now()
	slienceconfig := &sc.AlertSlienceConfig{
		AlertID:	alertID,
		For:		forDuration,
		StartAt:	time.Now(),
		EndAt:		now.Add(duration),
	}

	if oldCFGPtr, find := r.slientCFGMap.Swap(alertID, slienceconfig); find {
		cfg := oldCFGPtr.(*sc.AlertSlienceConfig)
		slienceconfig.ID = cfg.ID
		slienceconfig.AlertName = cfg.AlertName
		slienceconfig.Group = cfg.Group
		slienceconfig.Tags = cfg.Tags
		return r.database.UpdateAlertSlience(ctx_core, slienceconfig)
	} else {
		event, err := r.ch.GetLatestAlertEventByAlertID(ctx_core, alertID)
		if err == nil && event != nil {
			slienceconfig.AlertName = event.Name
			slienceconfig.Tags = event.EnrichTags
			slienceconfig.Group = event.Group
		}
		return r.database.AddAlertSlience(ctx_core, slienceconfig)
	}
}

func (r *InnerReceivers) RemoveSlienceConfigByAlertID(ctx_core core.Context, alertID string) error {
	if cfgPtr, loaded := r.slientCFGMap.LoadAndDelete(alertID); loaded {
		cfg := cfgPtr.(*sc.AlertSlienceConfig)
		return r.database.DeleteAlertSlience(ctx_core, cfg.ID)
	}

	return fmt.Errorf("alert[%s] is not slient", alertID)
}
