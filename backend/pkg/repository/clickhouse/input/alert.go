// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package input

import (
	"context"
	"encoding/json"
	"log"

	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

func (ch *chRepo) InsertAlertEvent(ctx context.Context, alertEvents []alert.AlertEvent, sourceFrom alert.SourceFrom) error {
	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO alert_event (id,name,group,severity, status, detail, alert_id, raw_tags, tags,create_time, update_time, end_time, received_time, source_id, source)
		VALUES
	`)

	if err != nil {
		return err
	}

	for _, event := range alertEvents {
		rawTagsStr := map[string]string{}
		for k, v := range event.RawTags {
			if str, ok := v.(string); ok {
				rawTagsStr[k] = str
			} else if bs, err := json.Marshal(v); err == nil {
				rawTagsStr[k] = string(bs)
			}
		}

		if err := batch.Append(
			event.ID,
			event.Name, event.Group, event.Severity, event.Status,
			event.Detail, event.AlertID,
			rawTagsStr, event.Tags,
			event.CreateTime, event.UpdateTime, event.EndTime, event.ReceivedTime,
			event.SourceID, sourceFrom.SourceName); err != nil {
			log.Println("Failed to send data:", err)
			continue
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}
