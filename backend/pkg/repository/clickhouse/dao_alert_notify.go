// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (ch *chRepo) CreateAlertNotifyRecord(ctx_core core.Context, ctx context.Context, record model.AlertNotifyRecord) error {
	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO alert_notify_record (alert_id, created_at, event_id, success,failed)
		VALUES
	`)
	if err != nil {
		return err
	}
	if err := batch.Append(
		record.AlertID,
		time.UnixMicro(record.CreatedAt),
		record.EventID,
		record.Success,
		record.Failed,
	); err != nil {
		return err
	}
	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}
