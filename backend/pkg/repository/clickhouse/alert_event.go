// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

func (ch *chRepo) InsertBatchAlertEvents(ctx context.Context, events []*model.AlertEvent) error {
	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO alert_event (source, id, alert_id, create_time, update_time, end_time, received_time, severity, group,
		                         name, detail, tags, status)
		VALUES
	`)
	if err != nil {
		return err
	}
	for _, event := range events {
		alertId := alert.FastAlertIDByStringMap(event.Name, event.Tags)
		if err := batch.Append(event.Source, event.ID, alertId, event.CreateTime, event.UpdateTime, event.EndTime,
			event.ReceivedTime, int8(event.Severity), event.Group, event.Name, event.Detail, event.Tags, int8(event.Status)); err != nil {
			log.Println("Failed to send data:", err)
			continue
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}

// ReadAlertEvent implement the Read method of the AlertEventDAO interface
func (ch *chRepo) ReadAlertEvent(ctx context.Context, id uuid.UUID) (*model.AlertEvent, error) {
	var event model.AlertEvent
	query := `
		SELECT source, id, create_time, update_time, end_time, received_time, severity
		       ,group, name, detail, tags, status
		FROM alert_event
		WHERE id = ?
	`
	err := ch.conn.QueryRow(ctx, query, id).Scan(
		&event.Source, &event.ID, &event.CreateTime, &event.UpdateTime, &event.EndTime,
		&event.ReceivedTime, &event.Severity, &event.Group, &event.Name, &event.Detail, &event.Tags, &event.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to read event: %w", err)
	}
	return &event, nil
}
