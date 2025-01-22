// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"sort"
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

// calculate AlertID based on alertName and raw_tag
func FastAlertID(alertName string, tags map[string]any) string {
	buf := new(bytes.Buffer)
	buf.WriteString(alertName)

	keys := make([]string, 0)
	for k, v := range tags {
		if _, ok := v.(string); ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		buf.WriteString(tags[key].(string))
	}

	hash := md5.Sum(buf.Bytes())
	return fmt.Sprintf("%x", hash)
}

func FastAlertIDByStringMap(alertName string, tags map[string]string) string {
	buf := new(bytes.Buffer)
	buf.WriteString(alertName)

	keys := make([]string, 0)
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		buf.WriteString(tags[key])
	}

	hash := md5.Sum(buf.Bytes())
	return fmt.Sprintf("%x", hash)
}
