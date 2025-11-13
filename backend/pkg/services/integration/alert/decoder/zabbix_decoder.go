// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"encoding/json"
	"math"
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

	now := time.Now()
	// parse in local timezone by default
	loc := time.Local

	// decide if the timestamp is in UTC by comparing updateTime with now
	if updateTimeStr, find := event["updateTime"]; find {
		updateTime, err := time.ParseInLocation(zabbixTimeLayout, updateTimeStr.(string), loc)
		if err == nil {
			// Compatible with legacy UTC timestamps
			if math.Abs(float64(updateTime.Sub(now))) > float64(time.Hour) {
				updateTime, _ = time.ParseInLocation(zabbixTimeLayout, updateTimeStr.(string), time.UTC)
				loc = time.UTC
			}
			event["updateTime"] = updateTime
			if duration, find := event["duration"]; find {
				duration := parseDuration(duration.(string))
				event["createTime"] = updateTime.Add(-duration)
			}
		}
	}

	if endTimeStr, find := event["endTime"]; find {
		endTime, err := time.ParseInLocation(zabbixTimeLayout, endTimeStr.(string), loc)
		if err == nil {
			event["endTime"] = endTime
		}
	}

	alertEvent, err := d.jsDecoder.convertAlertEvent(event)
	if err != nil {
		return nil, err
	}
	if len(alertEvent.EventID) == 0 {
		alertEvent.EventID = uuid.NewString()
	}

	alertEvent.SourceID = sourceFrom.SourceID
	alertEvent.Severity = alert.ConvertSeverity(sourceFrom.SourceType, alertEvent.Severity)
	alertEvent.Status = alert.ConvertStatus(sourceFrom.SourceType, alertEvent.Status)
	alertEvent.ReceivedTime = now

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
	var durationSeconds int64
	for _, part := range durationParts {
		switch part[len(part)-1] {
		case 'M':
			if month, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += int64(month) * int64(time.Hour) * 30 * 24
			}
		case 'd':
			if day, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += int64(day) * int64(time.Hour) * 24
			}
		case 'h':
			if hour, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += int64(hour) * int64(time.Hour)
			}
		case 'm':
			if minute, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += int64(minute) * int64(time.Minute)
			}
		case 's':
			if second, err := strconv.Atoi(part[:len(part)-1]); err == nil {
				durationSeconds += int64(second) * int64(time.Second)
			}
		}
	}

	return time.Duration(durationSeconds)
}
