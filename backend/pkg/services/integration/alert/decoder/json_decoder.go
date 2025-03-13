// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

type JsonDecoder struct{}

func (d JsonDecoder) Decode(sourceFrom alert.SourceFrom, data []byte) ([]alert.AlertEvent, error) {
	var event map[string]any
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	alertEvent, err := d.convertAlertEvent(event)
	if err != nil {
		return nil, err
	}

	if len(alertEvent.AlertID) == 0 {
		alertEvent.AlertID = alert.FastAlertID(alertEvent.Name, alertEvent.Tags)
	}
	alertEvent.ID = uuid.New()
	alertEvent.SourceID = sourceFrom.SourceID
	alertEvent.Severity = alert.ConvertSeverity(sourceFrom.SourceType, alertEvent.Severity)
	alertEvent.Status = alert.ConvertStatus(sourceFrom.SourceType, alertEvent.Status)
	alertEvent.ReceivedTime = time.Now()
	return []alert.AlertEvent{*alertEvent}, nil
}

func (d JsonDecoder) convertAlertEvent(rawMap map[string]any) (*alert.AlertEvent, error) {
	var alertEvent alert.AlertEvent

	err := DecodeEvent(rawMap, &alertEvent)
	if err != nil {
		return nil, err
	}
	if alertEvent.Tags == nil {
		alertEvent.Tags = make(map[string]any)
	}
	if len(alertEvent.AlertID) == 0 {
		alertEvent.AlertID = alert.FastAlertID(alertEvent.Name, alertEvent.Tags)
	}
	if alertEvent.Tags == nil {
		alertEvent.Tags = map[string]any{}
	}
	alertEvent.EnrichTags = map[string]string{}
	return &alertEvent, err
}

func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func DecodeEvent(input map[string]interface{}, result interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			ToTimeHookFunc()),
		Result: result,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}
	return err
}
