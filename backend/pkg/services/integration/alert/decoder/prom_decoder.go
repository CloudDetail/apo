// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/google/uuid"
	"go.uber.org/multierr"
)

type PrometheusDecoder struct{}

func (d PrometheusDecoder) Decode(sourceFrom alert.SourceFrom, data []byte) ([]alert.AlertEvent, error) {
	var promAlertList map[string]any
	err := json.Unmarshal(data, &promAlertList)
	if err != nil {
		return nil, err
	}

	var decodeErrs error
	events := promAlertList["alerts"].([]any)
	var alertEvents []alert.AlertEvent

	receivedTime := time.Now()
	for _, event := range events {
		rawMap := event.(map[string]any)
		alertEvent, err := d.convertAlertEvent(rawMap, receivedTime)
		if err != nil {
			decodeErrs = multierr.Append(decodeErrs, err)
			continue
		}
		if len(alertEvent.AlertID) == 0 {
			alertEvent.AlertID = alert.FastAlertID(alertEvent.Name, alertEvent.Tags)
		}
		alertEvent.ID = uuid.New()
		alertEvent.SourceID = sourceFrom.SourceID
		alertEvent.Severity = alert.ConvertSeverity(sourceFrom.SourceType, alertEvent.Severity)
		alertEvent.Status = alert.ConvertStatus(sourceFrom.SourceType, alertEvent.Status)
		alertEvent.ReceivedTime = receivedTime
		alertEvent.SetPayloadRef(promAlertList)
		alertEvents = append(alertEvents, *alertEvent)
	}

	return alertEvents, decodeErrs
}

func (d PrometheusDecoder) convertAlertEvent(rawMap map[string]any, receivedTime time.Time) (*alert.AlertEvent, error) {
	var promAlert request.Alert
	err := DecodeEvent(rawMap, &promAlert)
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

	var updateTime time.Time
	if promAlert.Status == model.StatusResolved.ToString() {
		updateTime = endsAt
	} else {
		updateTime = receivedTime
	}
	var alertEvent = alert.AlertEvent{
		Alert: alert.Alert{
			Name:       promAlert.Labels["alertname"],
			Group:      promAlert.Labels["group"],
			EnrichTags: map[string]string{},
			Tags:       tags,
		},
		Detail:     string(annotationsJson),
		CreateTime: startsAt,
		UpdateTime: updateTime,
		EndTime:    endsAt,
		Status:     promAlert.Status,
		Severity:   promAlert.Labels["severity"],
	}
	if alertEvent.Tags == nil {
		alertEvent.Tags = map[string]any{}
	}
	return &alertEvent, nil
}
