// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

type AlertEvent struct {
	Alert `mapstructure:",squash"`

	ID uuid.UUID `json:"id" ch:"id"`

	Detail string `json:"detail" ch:"detail" mapstructure:"detail"`

	CreateTime   time.Time `json:"createTime" ch:"create_time" mapstructure:"createTime"`
	UpdateTime   time.Time `json:"updateTime" ch:"update_time" mapstructure:"updateTime"`
	EndTime      time.Time `json:"endTime" ch:"end_time" mapstructure:"endTime"`
	ReceivedTime time.Time `json:"receivedTime" ch:"received_time"  mapstructure:"receivedTime"`

	Severity string `json:"severity" ch:"severity" mapstructure:"severity"`
	Status   string `json:"status" ch:"status" mapstructure:"status"`
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

func (e *AlertEvent) TagsInStr() string {
	param := e.EnrichTags
	param["status"] = e.Status
	// param["describe"] =
	bytes, err := json.Marshal(e.Tags)
	if err != nil {
		return "{}"
	}

	return string(bytes)
}

func (e *AlertEvent) ToAMAlert(timeout bool) *types.Alert {
	var convertLabels = make(model.LabelSet)
	for k, v := range e.EnrichTags {
		convertLabels[model.LabelName(k)] = model.LabelValue(v)
	}

	var convertAnnos = make(model.LabelSet)
	err := convertAnnos.UnmarshalJSON([]byte(e.Detail))
	if err != nil {
		convertAnnos["detail"] = model.LabelValue(e.Detail)
	}
	for k, v := range e.Tags {
		if vStr, ok := v.(string); ok {
			convertAnnos[model.LabelName(k)] = model.LabelValue(vStr)
		}
	}

	return &types.Alert{
		Alert: model.Alert{
			Labels:       convertLabels,
			Annotations:  convertAnnos,
			StartsAt:     e.CreateTime,
			EndsAt:       e.EndTime,
			GeneratorURL: "", // TODO from static config
		},
		UpdatedAt: e.UpdateTime,
		Timeout:   timeout,
	}
}
