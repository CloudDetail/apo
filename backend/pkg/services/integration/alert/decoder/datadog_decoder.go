// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type DatadogDecoder struct {
}

func (d *DatadogDecoder) Decode(sourceFrom alert.SourceFrom, data []byte) ([]alert.AlertEvent, error) {
	var event map[string]any
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	alertEvent := alert.AlertEvent{
		Alert: alert.Alert{
			Source:     sourceFrom.SourceName,
			SourceID:   sourceFrom.SourceID,
			AlertID:    getJson[string](event, "alert_id"),
			Group:      getJson[string](event, "group"),
			Name:       getJson[string](event, "name"),
			EnrichTags: make(map[string]string),
			Tags:       getJson[map[string]any](event, "tags"),
		},
		// ID:      uuid.UUID{},
		EventID: getJson[string](event, "event_id"),
		Detail:  getJson[string](event, "message"),
		// CreateTime:   time.Time{},
		// UpdateTime:   time.Time{},
		// EndTime:      time.Time{},
		// ReceivedTime: time.Time{},
		// Severity:     "",
		// Status:       "",
	}
	return []alert.AlertEvent{alertEvent}, nil
}

func getJson[T any](data map[string]any, key string) T {
	var result T
	if value, ok := data[key]; ok {
		if value, ok := value.(T); ok {
			return value
		}
	}
	return result
}
