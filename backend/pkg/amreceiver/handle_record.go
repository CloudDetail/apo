package amreceiver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (r *InnerReceivers) HandleAlertCheckRecord(ctx context.Context, record *model.WorkflowRecord) error {
	if record.WorkflowName != "AlertCheck" {
		return nil
	}
	if record.Output != "false" {
		return nil
	}

	alert, ok := record.InputRef.(alert.AlertEvent)
	if !ok {
		return fmt.Errorf("unexpect inputRef, should be alert.AlertEvent, got %T", record.InputRef)
	}

	if cfgPtr, find := r.slientCFGMap.Load(alert.AlertID); find {
		if cfg, ok := cfgPtr.(*slienceconfig.AlertSlienceConfig); ok && cfg.IsSlient(&alert) {
			return nil
		}
	}

	alertNotifyRecord := &model.AlertNotifyRecord{
		AlertID:  alert.AlertID,
		CreateAt: time.Now().UnixMicro(),
		EventID:  alert.ID.String(),
	}

	var fails []string
	var success []string
	var errs error
	for name, integrations := range r.receivers {
		for _, integration := range integrations {
			// TODO set timeout and retry
			_, err := integration.Notify(ctx, alert.ToAMAlert(false))
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("[%s] send alert failed: %w", name, err))
				fails = append(fails, fmt.Sprintf("%s:%s", name, integration.Name()))
			} else {
				success = append(success, fmt.Sprintf("%s:%s", name, integration.Name()))
			}
		}
	}

	alertNotifyRecord.Success = strings.Join(success, ";")
	alertNotifyRecord.Failed = strings.Join(fails, ";")

	err := r.ch.CreateAlertNotifyRecord(ctx, *alertNotifyRecord)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}
