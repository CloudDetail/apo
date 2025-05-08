// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package receiver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/prometheus/alertmanager/notify"

	pmodel "github.com/prometheus/common/model"
)

func (r *InnerReceivers) HandleAlertCheckRecord(ctx context.Context, record *model.WorkflowRecord) error {
	if record.WorkflowName != "AlertCheck" {
		return nil
	}
	if record.Output != "false" {
		return nil
	}

	alert, ok := record.InputRef.(*alert.AlertEvent)
	if !ok {
		return fmt.Errorf("unexpect inputRef, should be alert.AlertEvent, got %T", record.InputRef)
	}

	if _, find := r.slientCFGMap.Load(alert.AlertID); find {
		return nil
	}

	notifyRecord := &model.AlertNotifyRecord{
		AlertID:   alert.AlertID,
		CreatedAt: time.Now().UnixMicro(),
		EventID:   alert.ID.String(),
	}

	var fails []string
	var success []string
	var errs error

	ctx = notify.WithGroupKey(ctx, "alertName")
	ctx = notify.WithGroupLabels(ctx, pmodel.LabelSet{"alertName": pmodel.LabelValue(alert.Name)})
	for name, integrations := range r.receivers {
		ctx = notify.WithReceiverName(ctx, name)

		for _, integration := range integrations {
			alerts := alert.ToAMAlert(r.externalURL.String(), false)
			var err error
			var shouldRetry bool
			for retry := 3; retry > 0; retry-- {
				shouldRetry, err = integration.Notify(ctx, alerts)
				if !shouldRetry {
					break
				}
			}

			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("[%s] send alert failed: %w", name, err))
				fails = append(fails, fmt.Sprintf("%s:%s", name, integration.Name()))
			} else {
				success = append(success, fmt.Sprintf("%s:%s", name, integration.Name()))
			}
		}
	}

	if len(r.receivers) == 0 {
		// not set receiver
		notifyRecord.Failed = "no receiver set"
	} else {
		notifyRecord.Success = strings.Join(success, ";")
		notifyRecord.Failed = strings.Join(fails, ";")
	}

	err := r.ch.CreateAlertNotifyRecord(ctx, *notifyRecord)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}
