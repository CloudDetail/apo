// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"time"

	ainput "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

type JsonDecoder struct{}

func (d JsonDecoder) Decode(sourceFrom ainput.SourceFrom, data []byte) ([]ainput.AlertEvent, error) {
	var event map[string]any
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	alertEvent, err := d.convertAlertEvent(event)
	if err != nil {
		return nil, err
	}
	alertEvent.ID = uuid.New()
	if len(alertEvent.AlertID) == 0 {
		alertEvent.AlertID = fastAlertID(alertEvent.Name, alertEvent.Tags)
	}
	alertEvent.SourceID = sourceFrom.SourceID
	alertEvent.Severity = ainput.ConvertSeverity(sourceFrom.SourceType, alertEvent.Severity)
	alertEvent.Status = ainput.ConvertStatus(sourceFrom.SourceType, alertEvent.Status)
	alertEvent.ReceivedTime = time.Now()
	return []ainput.AlertEvent{*alertEvent}, nil
}

func (d JsonDecoder) convertAlertEvent(rawMap map[string]any) (*ainput.AlertEvent, error) {
	var alertEvent ainput.AlertEvent
	err := mapstructure.Decode(rawMap, &alertEvent)
	if err != nil {
		return nil, err
	}
	if alertEvent.Tags == nil {
		alertEvent.Tags = make(map[string]any)
	}
	if len(alertEvent.AlertID) == 0 {
		alertEvent.AlertID = fastAlertID(alertEvent.Name, alertEvent.Tags)
	}
	if alertEvent.Tags == nil {
		alertEvent.Tags = map[string]any{}
	}
	alertEvent.EnrichTags = map[string]string{}
	return &alertEvent, err
}
