// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"

	"github.com/google/uuid"
)

type AlertEvent struct {
	ID uuid.UUID `json:"id" ch:"id"`

	Group string `json:"group" ch:"group"`

	AlertID  string            `json:"alertId" ch:"alert_id"`
	Name     string            `json:"name" ch:"name"`
	Severity string            `json:"severity" ch:"severity"`
	Status   string            `json:"status" ch:"status"`
	Detail   string            `json:"detail" ch:"detail"`
	RawTags  map[string]any    `json:"raw_tags" ch:"raw_tags"`
	Tags     map[string]string `json:"tags" ch:"tags"`

	CreateTime   time.Time `ch:"create_time" json:"createTime"`
	UpdateTime   time.Time `ch:"update_time" json:"updateTime"`
	EndTime      time.Time `ch:"end_time" json:"endTime"`
	ReceivedTime time.Time `ch:"received_time" json:"receivedTime"`

	SourceID string `ch:"source_id"`
}

type RelatedService struct {
	ServiceName string `json:"serviceName"`
	Endpoint    string `json:"endpoint"`
}
