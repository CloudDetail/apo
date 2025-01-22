// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"reflect"
	"regexp"
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
	Group   string            `ch:"group" json:"group,omitempty"`
	Name    string            `ch:"name" json:"name,omitempty"`
	Detail  string            `ch:"detail" json:"detail,omitempty"`
	Tags    map[string]string `ch:"tags" json:"tags,omitempty"`
	RawTags map[string]string `ch:"raw_tags" json:"raw_tags,omitempty"`
	Status  Status            `ch:"status" json:"status,omitempty"`
}

func (a *AlertEvent) GetTargetObj() string {
	if a.Tags == nil {
		return ""
	}
	switch a.Group {
	case "app":
		return a.GetServiceNameTag()
	case "infra":
		return a.GetInfraNodeTag()
	case "network":
		return fmt.Sprintf("%s->%s", a.GetNetSrcIPTag(), a.GetNetDstIPTag())
	case "container":
		return fmt.Sprintf("%s(%s)", a.GetK8sPodTag(), a.GetContainerTag())
	case "middleware":
		return fmt.Sprintf("%s(%s:%s)",
			a.GetDatabaseURL(),
			a.GetDatabaseIP(),
			a.GetDatabasePort())
	}
	return ""
}

func (a *AlertEvent) GetServiceNameTag() string {
	if serviceName, find := a.Tags["svc_name"]; find && len(serviceName) > 0 {
		return serviceName
	}
	return a.Tags["serviceName"]
}

func (a *AlertEvent) GetEndpointTag() string {
	if contentKey, find := a.Tags["content_key"]; find && len(contentKey) > 0 {
		return contentKey
	}
	return a.Tags["endpoint"]
}

// GetLevelTag 获取级别,当前只有network告警存在
func (a *AlertEvent) GetLevelTag() string {
	return a.Tags["level"]
}

func (a *AlertEvent) GetNetSrcNodeTag() string {
	return a.Tags["node_name"]
}

func (a *AlertEvent) GetNetSrcPidTag() string {
	return a.Tags["pid"]
}

func (a *AlertEvent) GetNetSrcPodTag() string {
	//Compatible with older versions
	if pod, find := a.Tags["src_pod"]; find && len(pod) > 0 {
		return pod
	}
	return a.Tags["pod"]
}

func (a *AlertEvent) GetK8sNamespaceTag() string {
	//Compatible with older versions
	if namespace, find := a.Tags["src_namespace"]; find && len(namespace) > 0 {
		return namespace
	}
	return a.Tags["namespace"]
}

func (a *AlertEvent) GetK8sPodTag() string {
	if pod, find := a.Tags["pod_name"]; find && len(pod) > 0 {
		return pod
	}
	return a.Tags["pod"]
}

func (a *AlertEvent) GetContainerTag() string {
	if container, find := a.Tags["container_name"]; find && len(container) > 0 {
		return container
	}
	return a.RawTags["container"]
}

func (a *AlertEvent) GetInfraNodeTag() string {
	//Compatible with older versions
	if node, find := a.Tags["instance_name"]; find && len(node) > 0 {
		return node
	}
	return a.Tags["node"]
}

func (a *AlertEvent) GetNetSrcIPTag() string {
	//Compatible with older versions
	if ip, find := a.RawTags["src_ip"]; find && len(ip) > 0 {
		return ip
	}
	return a.RawTags["src_ip"]
}

func (a *AlertEvent) GetNetDstIPTag() string {
	//Compatible with older versions
	if ip, find := a.RawTags["dst_ip"]; find && len(ip) > 0 {
		return ip
	}
	return a.RawTags["dst_ip"]
}

var dbURLRegex = regexp.MustCompile(`tcp\((.+)\)`)
var dbIPRegex = regexp.MustCompile(`tcp\((\d+\.\d+\.\d+\.\d+):.*\)`)
var dbPortRegex = regexp.MustCompile(`tcp\(.*:(\d+)\)`)

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
