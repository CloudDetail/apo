// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"log"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

type AlertManagerEvent struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type AlertManagerHandler struct{}

func (h *AlertManagerHandler) handle(data *AlertManagerEvent) []*model.AlertEvent {
	events := make([]*model.AlertEvent, 0)
	for _, a := range data.Alerts {
		startsAt, _ := time.Parse(time.RFC3339, a.StartsAt)
		endsAt, err := time.Parse(time.RFC3339, a.EndsAt)
		if err != nil {
			log.Println("[AlertManager] Error parsing end time: ", err)
		}
		annotationsJson, err := json.Marshal(a.Annotations)
		if err != nil {
			log.Println("[AlertManager] Error marshaling annotations: ", err)
			continue
		}
		now := time.Now()
		alertEvent := &model.AlertEvent{
			ID:           model.GenUUID(),
			Source:       model.AlertManagerSource,
			Name:         a.Labels["alertname"],
			Severity:     h.convertSeverity(a.Labels["severity"]),
			Status:       h.convertStatus(a.Status),
			Group:        a.Labels["group"],
			CreateTime:   startsAt,
			UpdateTime:   now,
			ReceivedTime: now,
			EndTime:      endsAt,
			Detail:       string(annotationsJson),
			Tags:         a.Labels,
		}

		events = append(events, alertEvent)
	}
	return events
}
func (h *AlertManagerHandler) convertSeverity(severity string) model.SeverityLevel {
	switch severity {
	case "critical":
		return model.SeverityLevelCritical
	case "error":
		return model.SeverityLevelError
	case "warning":
		return model.SeverityLevelWarning
	case "info":
		return model.SeverityLevelInfo
	default:
		return model.SeverityLevelWarning
	}
}

func (h *AlertManagerHandler) convertStatus(status string) model.Status {
	switch status {
	case "firing":
		return model.StatusFiring
	case "resolved":
		return model.StatusResolved
	default:
		return model.StatusFiring
	}
}
