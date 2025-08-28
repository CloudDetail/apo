// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider/lifecycle"
)

type DatadogDecoder struct{}

type DatadogPayload struct {
	AlertID         string `json:"alert_id"`
	AggRegKey       string `json:"agg_reg_key"`
	AlertType       string `json:"alert_type"`
	AlertQuery      string `json:"alert_query"`
	AlertTitle      string `json:"alert_title"`
	AlertCircleKey  string `json:"alert_cycle_key"`
	AlertTransition string `json:"alert_transition"`
	ID              string `json:"id"` // EventID
	EventTitle      string `json:"event_title"`
	EventType       string `json:"event_type"`
	EventMsg        string `json:"event_msg"`
	LastUpdated     string `json:"last_updated"`
	AlertPriority   string `json:"alert_priority"`
	Tags            string `json:"tags"`
	Date            string `json:"date"`
	OrgID           string `json:"org_id"`
	OrgName         string `json:"org_name"`
}

func (d DatadogDecoder) Decode(sourceFrom alert.SourceFrom, data []byte) ([]alert.AlertEvent, error) {
	var payload DatadogPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}

	// AlertID: Same as monitor_id. Identifies the monitor rule and stays the same across all scopes.
	// AggRegKey: Unique per scope within the same monitor. Identifies a specific monitor+scope alert instance.
	// AlertCycleKey: Unique for each trigger-to-resolve cycle within the same monitor and scope. Changes on every new alert cycle.
	// ID: Unique identifier for each event. Different events always have different IDs, even within the same cycle.
	alertID := fmt.Sprintf("%s-%s", sourceFrom.SourceID[:8], payload.AggRegKey)
	eventID := fmt.Sprintf("%s-%s", sourceFrom.SourceID[:8], payload.ID)

	// Only add eventID to the cache; no need to check since webhook is the primary event
	lifecycle.AlertLifeCycle.CheckEventSeen(eventID)

	status := transDDStatus(payload.AlertTransition)
	var updateTime, endTime time.Time
	dateTS, err := strconv.ParseInt(payload.Date, 10, 64)
	if err != nil {
		return nil, err
	}
	updateTime = time.UnixMilli(dateTS)
	if status == alert.StatusResolved {
		endTime = updateTime
	}

	createTime, find := lifecycle.AlertLifeCycle.CacheEventStatus(alertID, status, updateTime)
	if !find {
		createTime = updateTime
	}

	tagStrs := strings.Split(payload.Tags, ",")
	var tags = make(map[string]any, len(tagStrs))
	for _, tag := range tagStrs {
		if strings.Contains(tag, ":") {
			tagKV := strings.SplitN(tag, ":", 2)
			tags[tagKV[0]] = tagKV[1]
		} else {
			tags[tag] = ""
		}
	}
	tags["org_id"] = payload.OrgID
	tags["org_name"] = payload.OrgName

	detail := map[string]string{
		"query":       payload.AlertQuery,
		"summary":     payload.EventTitle,
		"description": payload.EventMsg,
		"tags":        payload.Tags,
	}

	detailJsonStr, _ := json.Marshal(detail)

	alertEvent := alert.AlertEvent{
		Alert: alert.Alert{
			Source:            sourceFrom.SourceName,
			SourceID:          sourceFrom.SourceID,
			AlertID:           alertID,
			Group:             "",
			Name:              payload.AlertTitle,
			EnrichTags:        map[string]string{},
			EnrichTagsDisplay: []alert.TagDisplay{},
			Tags:              tags,
		},
		EventID:      eventID,
		Detail:       string(detailJsonStr),
		CreateTime:   createTime,
		UpdateTime:   updateTime,
		EndTime:      endTime,
		ReceivedTime: time.Now(),
		Severity:     transDDSeverity(payload.AlertPriority),
		Status:       status,
	}

	return []alert.AlertEvent{alertEvent}, nil
}

var DDPriorityMap = map[string]string{
	"P1": alert.SeverityCriticalLevel,
	"P2": alert.SeverityErrorLevel,
	"P3": alert.SeverityWarnLevel,
	"P4": alert.SeverityInfoLevel,
}

func transDDSeverity(priority string) string {
	if severity, ok := DDPriorityMap[priority]; ok {
		return severity
	}
	return alert.SeverityInfoLevel
}

func transDDStatus(transition string) string {
	if transition == "Recovered" {
		return alert.StatusResolved
	}
	// Include Triggered/Re-Triggered, No Data/Re-No Data, Warn/Re-Warn, Renotify
	return alert.StatusFiring
}
