// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func transferAlertManager(data *request.InputAlertManagerRequest) []*model.AlertEvent {
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
			ID:		model.GenUUID(),
			Source:		model.AlertManagerSource,
			Name:		a.Labels["alertname"],
			Severity:	convertSeverity(a.Labels["severity"]),
			Status:		convertStatus(a.Status),
			Group:		a.Labels["group"],
			CreateTime:	startsAt,
			UpdateTime:	now,
			ReceivedTime:	now,
			EndTime:	endsAt,
			Detail:		string(annotationsJson),
			Tags:		a.Labels,
		}

		events = append(events, alertEvent)
	}
	return events
}
func convertSeverity(severity string) model.SeverityLevel {
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

func convertStatus(status string) model.Status {
	switch status {
	case "firing":
		return model.StatusFiring
	case "resolved":
		return model.StatusResolved
	default:
		return model.StatusFiring
	}
}

func (s *service) InputAlertManager(ctx_core core.Context, req *request.InputAlertManagerRequest) error {
	events := transferAlertManager(req)
	err := s.chRepo.InsertBatchAlertEvents(context.Background(), events)
	if err != nil {
		log.Println("[AlertManager] Error inserting data: ", err)
		return err
	}
	return nil
}
