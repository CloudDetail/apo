// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	pd "github.com/PagerDuty/go-pagerduty"
)

type PagerDutyDecoder struct{}

func (d PagerDutyDecoder) Decode(sourceFrom alert.SourceFrom, data []byte) ([]alert.AlertEvent, error) {
	var pdEvent PagerDutyEvent
	if err := json.Unmarshal(data, &pdEvent); err != nil {
		return nil, err
	}

	// Event.Data.ID -> Incident ID
	// Event.ID -> Event ID
	alertID := fmt.Sprintf("%s-%s", sourceFrom.SourceID[:8], pdEvent.Event.Data.ID)
	eventID := fmt.Sprintf("%s-%s", sourceFrom.SourceID[:8], pdEvent.Event.ID)

	createTime, err := time.Parse(time.RFC3339, pdEvent.Event.Data.CreatedAt)
	if err != nil {
		log.Printf("parse pagerduty event create time failed, err: %v", err)
		return nil, err
	}

	// According to the PagerDuty documentation, the UpdatedAt and ResolvedAt is always empty
	// https://developer.pagerduty.com/docs/webhooks-overview#event-types
	// if len(pdEvent.Event.Data.UpdatedAt) > 0 {
	// 	updateTime, err = time.Parse(time.RFC3339, pdEvent.Event.Data.UpdatedAt)
	// 	if err != nil {
	// 		log.Printf("parse pagerduty event update time failed, err: %v", err)
	// 		return nil, err
	// 	}
	// }
	// if len(pdEvent.Event.Data.ResolvedAt) > 0 {
	// 	endTime, err = time.Parse(time.RFC3339, pdEvent.Event.Data.ResolvedAt)
	// 	if err != nil {
	// 		log.Printf("parse pagerduty event end time failed, err: %v", err)
	// 		return nil, err
	// 	}
	// }
	var updateTime time.Time
	if len(pdEvent.Event.OccurredAt) > 0 {
		updateTime, err = time.Parse(time.RFC3339, pdEvent.Event.OccurredAt)
		if err != nil {
			log.Printf("parse pagerduty event update time failed, err: %v", err)
			return nil, err
		}
	}

	var endTime time.Time
	if pdEvent.Event.Data.Status == "resolved" {
		endTime = updateTime
	}

	var severity string
	if pdEvent.Event.Data.Priority != nil {
		severity = getPagerDutySeverity(pdEvent.Event.Data.Priority.Summary)
	} else if pdEvent.Event.Data.Urgency == "high" {
		severity = alert.SeverityWarnLevel
	}

	alertEvent := alert.AlertEvent{
		Alert: alert.Alert{
			Source:     sourceFrom.SourceName,
			SourceID:   sourceFrom.SourceID,
			AlertID:    alertID,
			Group:      getPagerDutyGroup(&pdEvent),
			Name:       pdEvent.Event.Data.Title,
			EnrichTags: map[string]string{},
			Tags:       getPagerDutySimpleTags(&pdEvent),
		},
		EventID:      eventID,
		Detail:       pdEvent.Event.Data.Title, // PagerDuty will not send eventDetail in webhook request
		CreateTime:   createTime,
		UpdateTime:   updateTime,
		EndTime:      endTime,
		ReceivedTime: time.Now(),
		Severity:     severity,
		Status:       getPagerDutyStatus(pdEvent.Event.Data.Status),
	}
	return []alert.AlertEvent{alertEvent}, nil
}

type PagerDutyEvent struct {
	Event struct {
		ID           string      `json:"id"`
		EventType    string      `json:"event_type"`
		ResourceType string      `json:"resource_type"`
		OccurredAt   string      `json:"occurred_at"`
		Agent        Agent       `json:"agent"`
		Client       Client      `json:"client"`
		Data         pd.Incident `json:"data"`
	} `json:"event"`
}

type Agent struct {
	HtmlURL string `json:"html_url"`
	ID      string `json:"id"`
	Self    string `json:"self"`
	Summary string `json:"summary"`
	Type    string `json:"type"`
}

type Client struct {
	Name string `json:"name"`
}

var PagerDutyPriorityMap = map[string]string{
	"P1": alert.SeverityCriticalLevel,
	"P2": alert.SeverityErrorLevel,
	"P3": alert.SeverityWarnLevel,
	"P4": alert.SeverityInfoLevel,
	"P5": alert.SeverityInfoLevel,
}

var PagerDutyStatusMap = map[string]string{
	"triggered": alert.StatusFiring,
	"resolved":  alert.StatusResolved,
}

func getPagerDutySeverity(pdPriority string) string {
	if severity, ok := PagerDutyPriorityMap[pdPriority]; ok {
		return severity
	}
	return alert.SeverityInfoLevel
}

func getPagerDutyStatus(pdStatus string) string {
	if status, ok := PagerDutyStatusMap[pdStatus]; ok {
		return status
	}
	return alert.StatusFiring
}

func getPagerDutyGroup(pdEvent *PagerDutyEvent) string {
	if len(pdEvent.Event.Data.Service.Summary) > 0 && pdEvent.Event.Data.Service.Type == "service_reference" {
		return string(clickhouse.APP_GROUP)
	}

	return ""
}

func getPagerDutySimpleTags(pdEvent *PagerDutyEvent) map[string]any {
	res := make(map[string]any)

	if len(pdEvent.Event.Data.Service.Summary) > 0 {
		res["service_id"] = pdEvent.Event.Data.Service.ID
		res["service"] = pdEvent.Event.Data.Service.Summary
	}

	if len(pdEvent.Event.Data.Teams) > 0 {
		var teams []string
		var teamIDs []string
		for _, team := range pdEvent.Event.Data.Teams {
			teams = append(teams, team.Summary)
			teamIDs = append(teamIDs, team.ID)
		}
		res["teams"] = strings.Join(teams, ",")
		res["team_ids"] = strings.Join(teamIDs, ",")
	}
	if len(pdEvent.Event.Agent.Summary) > 0 {
		res["agent_id"] = pdEvent.Event.Agent.ID
		res["agent"] = pdEvent.Event.Agent.Summary
	}
	if len(pdEvent.Event.Client.Name) > 0 {
		res["client"] = pdEvent.Event.Client.Name
	}
	return res
}
