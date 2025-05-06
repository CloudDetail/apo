package amreceiver

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
)

func (r *InnerReceivers) GetSlienceConfig(alertID string) (*slienceconfig.AlertSlienceConfig, error) {
	if cfgPtr, find := r.slientCFGMap.Load(alertID); find {
		if cfg, ok := cfgPtr.(*slienceconfig.AlertSlienceConfig); ok {
			return cfg, nil
		}
		// Impossible branch
		return nil, fmt.Errorf("unexpected SlienceConfig, expected <*SlienceConfig.AlertSlienceConfig>")
	}
	return nil, nil
}

func (r *InnerReceivers) ListSlienceConfig() ([]slienceconfig.AlertSlienceConfig, error) {
	return r.database.GetAlertSlience()
}

func (r *InnerReceivers) SetSlienceConfig(alertID string, forDuration string) error {
	duration, err := time.ParseDuration(forDuration)
	if err != nil {
		return fmt.Errorf("duration is not valid: %w", err)
	}

	now := time.Now()
	slienceconfig := &slienceconfig.AlertSlienceConfig{
		AlertID: alertID,
		For:     forDuration,
		StartAt: time.Now(),
		EndAt:   now.Add(duration),
	}

	if _, find := r.slientCFGMap.Swap(alertID, slienceconfig); find {
		return r.database.UpdateAlertSlience(slienceconfig)
	} else {
		return r.database.AddAlertSlience(slienceconfig)
	}
}

func (r *InnerReceivers) RemoveSlienceConfig(alertID string) error {
	if _, loaded := r.slientCFGMap.LoadAndDelete(alertID); loaded {
		return r.database.DeleteAlertSlience(alertID)
	}

	return fmt.Errorf("alert[%s] is not slient", alertID)
}
