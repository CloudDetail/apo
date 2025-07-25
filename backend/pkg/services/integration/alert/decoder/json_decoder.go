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
	alertEvent.SetPayloadRef(event)
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
	return func(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch v := data.(type) {
		case string:
			if ts, err := time.ParseInLocation(time.DateTime, v, time.Local); err == nil {
				return ts, nil
			}
			return time.Parse(time.RFC3339, v)
		case float64:
			return inferTimeFromNumber(int64(v))
		case int64:
			return inferTimeFromNumber(v)
		default:
			return data, nil
		}
	}
}

func inferTimeFromNumber(v int64) (time.Time, error) {
	switch {
	case v > 1e18:
		return time.Unix(0, v), nil
	case v > 1e15:
		return time.UnixMicro(v), nil
	case v > 1e12:
		return time.Unix(0, v*int64(time.Millisecond)), nil
	case v > 1e9:
		return time.Unix(v, 0), nil
	default:
		return time.Unix(v, 0), nil
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
