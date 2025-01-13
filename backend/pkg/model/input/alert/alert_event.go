// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"

	"github.com/google/uuid"
)

type AlertEvent struct {
	ID uuid.UUID `json:"id" ch:"id"` // 告警记录ID

	Group string `json:"group" ch:"group"` // 告警对象类型

	AlertID  string            `json:"alertId" ch:"alert_id"`  // 告警事件ID
	Name     string            `json:"name" ch:"name"`         // 告警事件名称
	Severity string            `json:"severity" ch:"severity"` // 告警事件级别
	Status   string            `json:"status" ch:"status"`     // 告警事件状态
	Detail   string            `json:"detail" ch:"detail"`     // 告警事件内容
	RawTags  map[string]any    `json:"raw_tags" ch:"raw_tags"` // 原始的标签信息
	Tags     map[string]string `json:"tags" ch:"tags"`         // 丰富后的标签信息

	CreateTime   time.Time `ch:"create_time" json:"createTime"`     // 故障触发时间
	UpdateTime   time.Time `ch:"update_time" json:"updateTime"`     // 故障最后一次发生时间
	EndTime      time.Time `ch:"end_time" json:"endTime"`           // 故障恢复时间（仅恢复时存在）
	ReceivedTime time.Time `ch:"received_time" json:"receivedTime"` // 故障事件接收时间（用于记录数据对接，无业务含义）

	SourceID string `ch:"source_id"`
}

type RelatedService struct {
	ServiceName string `json:"serviceName"`
	Endpoint    string `json:"endpoint"`
}
