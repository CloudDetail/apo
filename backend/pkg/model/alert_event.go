// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// AlertEvent indicates an event in the alert_event table
type AlertEvent struct {
	Source string    `ch:"source" json:"source,omitempty"`
	ID     uuid.UUID `ch:"id" json:"id,omitempty"`
	// fault trigger time
	CreateTime time.Time `ch:"create_time" json:"createTime"`
	// Last time the fault occurred
	UpdateTime time.Time `ch:"update_time" json:"updateTime"`
	// Recovery time (only present at recovery)
	EndTime time.Time `ch:"end_time" json:"endTime"`
	// Failure event reception time (used to record data connection, no business meaning)
	ReceivedTime time.Time     `ch:"received_time" json:"receivedTime"`
	Severity     SeverityLevel `ch:"severity" json:"severity,omitempty"`
	// Fault group information
	Group  string            `ch:"group" json:"group,omitempty"`
	Name   string            `ch:"name" json:"name,omitempty"`
	Detail string            `ch:"detail" json:"detail,omitempty"`
	Tags   map[string]string `ch:"tags" json:"tags,omitempty"`
	Status Status            `ch:"status" json:"status,omitempty"`
}

func (a *AlertEvent) GetTargetObj() string {
	if a.Tags == nil {
		return ""
	}
	switch a.Group {
	case "app":
		return a.Tags["svc_name"]
	case "infra":
		return a.Tags["instance_name"]
	case "network":
		return fmt.Sprintf("%s->%s", a.Tags["src_ip"], a.Tags["dst_ip"])
	case "container":
		return fmt.Sprintf("%s(%s)", a.Tags["pod"], a.Tags["container"])
	}
	return ""
}

func GenUUID() uuid.UUID {
	return uuid.New()
}

const (
	AlertManagerSource = "alertManager"
	ZabbixSource       = "zabbix"
)

type SeverityLevel uint8

const (
	SeverityLevelUnknown SeverityLevel = iota
	SeverityLevelInfo
	SeverityLevelWarning
	SeverityLevelError
	SeverityLevelCritical
)

// Scan implements sql.Scanner so SeverityLevel can be read from databases transparently.
// Currently, database types that map to uint8 and []byte are supported.
func (s *SeverityLevel) Scan(src interface{}) error {
	switch v := src.(type) {
	case uint8:
		*s = SeverityLevel(v)
	case uint64:
		*s = SeverityLevel(v)
	case string:
		switch v {
		case "info":
			*s = SeverityLevelInfo
		case "warning":
			*s = SeverityLevelWarning
		case "error":
			*s = SeverityLevelError
		case "critical":
			*s = SeverityLevelCritical
		default:
			*s = SeverityLevelUnknown
		}
	default:
		return fmt.Errorf("can not covert %v to SeverityLevel", reflect.TypeOf(src))
	}
	return nil
}

func Str2SeverityLevel(levelText string) SeverityLevel {
	switch levelText {
	case "info":
		return SeverityLevelInfo
	case "warning":
		return SeverityLevelWarning
	case "error":
		return SeverityLevelError
	case "critical":
		return SeverityLevelCritical
	default:
		return SeverityLevelUnknown
	}
}

func (s SeverityLevel) toString() string {
	switch s {
	case SeverityLevelUnknown:
		return "unknown"
	case SeverityLevelInfo:
		return "info"
	case SeverityLevelWarning:
		return "warning"
	case SeverityLevelError:
		return "error"
	case SeverityLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Status defines the status of the event
type Status int8

const (
	StatusResolved Status = iota
	StatusFiring
)

func (s *Status) Scan(src interface{}) error {
	switch v := src.(type) {
	case uint8:
		*s = Status(v)
	case uint64:
		*s = Status(v)
	case string:
		switch v {
		case "resolved":
			*s = StatusResolved
		case "firing":
			*s = StatusFiring
		default:
			*s = StatusResolved
		}
	default:
		return fmt.Errorf("can not covert %v to Status", reflect.TypeOf(src))
	}
	return nil
}

type AlertEventCount struct {
	Severity SeverityLevel     `ch:"severity" json:"severity,omitempty"`
	Group    string            `ch:"group" json:"group,omitempty"`
	Tags     map[string]string `ch:"tags" json:"tags,omitempty"`

	Rn         uint64 `ch:"rn" json:"-"`
	AlarmCount uint64 `ch:"alarm_count" json:"alarmCount"`
}
