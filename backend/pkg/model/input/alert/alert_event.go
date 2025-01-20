// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"

	"github.com/google/uuid"
)

type AlertEvent struct {
	ID    uuid.UUID `json:"id" ch:"id"`
	Group string    `json:"group" ch:"group" mapstructure:"group"`

	AlertID  string `json:"alertId" ch:"alert_id" mapstructure:"alertId"`
	Name     string `json:"name" ch:"name" mapstructure:"name"`
	Severity string `json:"severity" ch:"severity" mapstructure:"severity"`
	Status   string `json:"status" ch:"status" mapstructure:"status"`
	Detail   string `json:"detail" ch:"detail" mapstructure:"detail"`
	// HACK the existing clickhouse query uses `tags` as the filter field
	// so enrichTags in ch is named as 'tags' to filter new alertInput
	Tags       map[string]any    `json:"tags" ch:"raw_tags" mapstructure:"tags"`
	EnrichTags map[string]string `json:"enrich_tags" ch:"tags" mapstructure:"enrich_tags"`

	CreateTime   time.Time `ch:"createTime" json:"createTime" mapstructure:"createTime"`
	UpdateTime   time.Time `ch:"updateTime" json:"updateTime" mapstructure:"updateTime"`
	EndTime      time.Time `ch:"endTime" json:"endTime" mapstructure:"endTime"`
	ReceivedTime time.Time `ch:"receivedTime" json:"receivedTime" mapstructure:"receivedTime"`

	SourceID string `ch:"sourceId" json:"sourceId" mapstructure:"sourceId"`
}
