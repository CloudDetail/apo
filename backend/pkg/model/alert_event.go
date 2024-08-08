package model

import (
	"time"

	"github.com/google/uuid"
)

// AlertEvent 表示alert_event表中的一个事件
type AlertEvent struct {
	Source string    `ch:"source" json:"source,omitempty"`
	ID     uuid.UUID `ch:"id" json:"id,omitempty"`
	// 故障触发时间
	CreateTime time.Time `ch:"create_time" json:"create_time"`
	// 故障最后一次发生时间
	UpdateTime time.Time `ch:"update_time" json:"update_time"`
	// 故障恢复时间（仅恢复时存在）
	EndTime time.Time `ch:"end_time" json:"end_time"`
	// 故障事件接收时间（用于记录数据对接，无业务含义）
	ReceivedTime time.Time     `ch:"received_time" json:"received_time"`
	Severity     SeverityLevel `ch:"severity" json:"severity,omitempty"`
	// 故障所属分组信息
	Group  string            `ch:"group" json:"group,omitempty"`
	Name   string            `ch:"name" json:"name,omitempty"`
	Detail string            `ch:"detail" json:"detail,omitempty"`
	Tags   map[string]string `ch:"tags" json:"tags,omitempty"`
	Status Status            `ch:"status" json:"status,omitempty"`
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

// Status 定义了事件的状态
type Status int8

const (
	StatusResolved Status = iota
	StatusFiring
)
