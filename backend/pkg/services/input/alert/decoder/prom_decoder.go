// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"time"

	inputa "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/multierr"
)

type PrometheusDecoder struct{}

func (d PrometheusDecoder) Decode(sourceFrom inputa.SourceFrom, data []byte) ([]inputa.AlertEvent, error) {
	var promAlertList map[string]any
	err := json.Unmarshal(data, &promAlertList)
	if err != nil {
		return nil, err
	}

	var decodeErrs error
	events := promAlertList["alerts"].([]any)
	var alertEvents []inputa.AlertEvent

	receivedTime := time.Now()
	for _, event := range events {
		rawMap := event.(map[string]any)
		alertEvent, err := d.convertAlertEvent(rawMap)
		if err != nil {
			decodeErrs = multierr.Append(decodeErrs, err)
			continue
		}
		alertEvent.ID = uuid.New()
		if len(alertEvent.AlertID) == 0 {
			alertEvent.AlertID = fastAlertID(alertEvent.Name, alertEvent.RawTags)
		}
		alertEvent.SourceID = sourceFrom.SourceID
		alertEvent.Severity = inputa.ConvertSeverity(sourceFrom.SourceType, alertEvent.Severity)
		alertEvent.Status = inputa.ConvertStatus(sourceFrom.SourceType, alertEvent.Status)
		alertEvent.ReceivedTime = receivedTime
		alertEvents = append(alertEvents, *alertEvent)
	}

	return alertEvents, decodeErrs
}

func (d PrometheusDecoder) convertAlertEvent(rawMap map[string]any) (*inputa.AlertEvent, error) {
	var promAlert request.Alert
	err := mapstructure.Decode(rawMap, &promAlert)
	if err != nil {
		return nil, err
	}
	annotationsJson, err := json.Marshal(promAlert.Annotations)
	if err != nil {
		return nil, err
	}
	startsAt, err := time.Parse(time.RFC3339, promAlert.StartsAt)
	if err != nil {
		return nil, err
	}
	endsAt, err := time.Parse(time.RFC3339, promAlert.EndsAt)
	if err != nil {
		return nil, err
	}

	tags := map[string]any{}
	for k, v := range promAlert.Labels {
		tags[k] = v
	}
	var alertEvent = inputa.AlertEvent{
		Name:       promAlert.Labels["alertname"],
		Detail:     string(annotationsJson),
		RawTags:    tags,
		Tags:       map[string]string{},
		CreateTime: startsAt,
		UpdateTime: startsAt,
		EndTime:    endsAt,
	}
	if alertEvent.RawTags == nil {
		alertEvent.RawTags = map[string]any{}
	}
	return &alertEvent, nil
}
