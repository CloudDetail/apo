// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package receiver

import (
	"fmt"
	"time"

	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
)

func (r *InnerReceivers) GetSlienceConfigByAlertID(alertID string) (*sc.AlertSlienceConfig, error) {
	if cfgPtr, find := r.slientCFGMap.Load(alertID); find {
		return cfgPtr.(*sc.AlertSlienceConfig), nil
	}
	return nil, nil
}

func (r *InnerReceivers) ListSlienceConfig() ([]sc.AlertSlienceConfig, error) {
	return r.database.GetAlertSlience()
}

func (r *InnerReceivers) SetSlienceConfigByAlertID(alertID string, forDuration string) error {
	duration, err := time.ParseDuration(forDuration)
	if err != nil {
		return fmt.Errorf("duration is not valid: %w", err)
	}

	now := time.Now()
	slienceconfig := &sc.AlertSlienceConfig{
		AlertID: alertID,
		For:     forDuration,
		StartAt: time.Now(),
		EndAt:   now.Add(duration),
	}

	if oldCFGPtr, find := r.slientCFGMap.Swap(alertID, slienceconfig); find {
		cfg := oldCFGPtr.(*sc.AlertSlienceConfig)
		slienceconfig.ID = cfg.ID
		slienceconfig.AlertName = cfg.AlertName
		slienceconfig.Group = cfg.Group
		slienceconfig.Tags = cfg.Tags
		return r.database.UpdateAlertSlience(slienceconfig)
	} else {
		event, err := r.ch.GetLatestAlertEventByAlertID(alertID)
		if err == nil && event != nil {
			slienceconfig.AlertName = event.Name
			slienceconfig.Tags = event.EnrichTags
			slienceconfig.Group = event.Group
		}
		return r.database.AddAlertSlience(slienceconfig)
	}
}

func (r *InnerReceivers) RemoveSlienceConfigByAlertID(alertID string) error {
	if cfgPtr, loaded := r.slientCFGMap.LoadAndDelete(alertID); loaded {
		cfg := cfgPtr.(*sc.AlertSlienceConfig)
		return r.database.DeleteAlertSlience(cfg.ID)
	}

	return fmt.Errorf("alert[%s] is not slient", alertID)
}
