// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/google/uuid"
)

type ZabbixDecoder struct {
	jsDecoder JsonDecoder
}

const zabbixTimeLayout = "2006.01.02 15:04:05"

func (d ZabbixDecoder) Decode(sourceFrom alert.SourceFrom, data []byte) ([]alert.AlertEvent, error) {
	var event map[string]any
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	if endTimeStr, find := event["endTime"]; find {
		endTime, err := time.Parse(zabbixTimeLayout, endTimeStr.(string))
		if err == nil {
			event["endTime"] = endTime
		}
	}

	if updateTimeStr, find := event["updateTime"]; find {
		updateTime, err := time.Parse(zabbixTimeLayout, updateTimeStr.(string))
		if err == nil {
			event["updateTime"] = updateTime
			if duration, find := event["duration"]; find {
				duration := parseDuration(duration.(string))
				event["createTime"] = updateTime.Add(-duration)
			}
		}
	}

	alertEvent, err := d.jsDecoder.convertAlertEvent(event)
	alertEvent.ID = uuid.New()
	alertEvent.SourceID = sourceFrom.SourceID
	alertEvent.Severity = alert.ConvertSeverity(sourceFrom.SourceType, alertEvent.Severity)
	alertEvent.Status = alert.ConvertStatus(sourceFrom.SourceType, alertEvent.Status)
	alertEvent.ReceivedTime = time.Now()

	alertEvent.Tags["alert_id"] = alertEvent.AlertID
	alertEvent.Tags["name"] = alertEvent.Name
	alertEvent.Tags["severity"] = alertEvent.Severity
	alertEvent.Tags["status"] = alertEvent.Status
	alertEvent.Tags["group"] = alertEvent.Group

	return []alert.AlertEvent{*alertEvent}, err
}

// M d h m s
func parseDuration(duration string) time.Duration {
	durationParts := strings.Split(duration, " ")
	var durationSeconds = 0
	for _, part := range durationParts {
		switch part[len(part)-1] {
		case 'M':
			if month, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += month * int(time.Hour) * 30 * 24
			}
		case 'd':
			if day, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += day * int(time.Hour) * 24
			}
		case 'h':
			if hour, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += hour * int(time.Hour)
			}
		case 'm':
			if minute, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += minute * int(time.Minute)
			}
		case 's':
			if second, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += second * int(time.Second)
			}
		}
	}

	return time.Duration(durationSeconds)
}
