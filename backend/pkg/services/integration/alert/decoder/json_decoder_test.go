// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func TestJsonDecoder_Decode(t *testing.T) {
	createTime, _ := time.Parse(time.RFC3339, "2025-01-21T15:04:05+00:00")
	type args struct {
		sourceFrom alert.SourceFrom
		data       []byte
	}
	tests := []struct {
		name    string
		d       JsonDecoder
		args    args
		want    []alert.AlertEvent
		wantErr bool
	}{
		{
			name: "basicAlert",
			d:    JsonDecoder{},
			args: args{
				sourceFrom: alert.SourceFrom{
					SourceID: "abcd",
					SourceInfo: alert.SourceInfo{
						SourceName: "externalAlert",
						SourceType: "json",
					},
				},
				data: jsonMarshal(
					map[string]any{
						"name":    "service延时增加",
						"detail":  "服务出现异常",
						"alertId": "1234567890",
						"tags": map[string]string{
							"pod":       "ts-price-service-5fb976df54-k8m6m",
							"namespace": "train-ticket",
							"node":      "larger-server-6",
						},
						"createTime": "2025-01-21T15:04:05+00:00",
						"updateTime": int64(1737514800000),
						"endTime":    int64(1737514800000),
						"severity":   "error",
						"status":     "firing",
					},
				),
			},
			want: []alert.AlertEvent{
				{
					Alert: alert.Alert{
						SourceID:   "abcd",
						AlertID:    "1234567890",
						Group:      "",
						Name:       "service延时增加",
						EnrichTags: map[string]string{},
						Tags: alert.RawTags{
							"pod":       "ts-price-service-5fb976df54-k8m6m",
							"namespace": "train-ticket",
							"node":      "larger-server-6",
						},
					},
					EventID:      "",
					Detail:       "服务出现异常",
					CreateTime:   createTime,
					UpdateTime:   time.UnixMilli(1737514800000),
					EndTime:      time.UnixMilli(1737514800000),
					ReceivedTime: time.UnixMilli(0),
					Severity:     "error",
					Status:       "firing",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := JsonDecoder{}
			got, err := d.Decode(tt.args.sourceFrom, tt.args.data)
			for i := range got {
				got[i].EventID = ""
				got[i].ReceivedTime = time.Unix(0, 0)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("JsonDecoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonDecoder.Decode() = \n%+v \n want: \n%+v", got, tt.want)
			}
		})
	}
}

func jsonMarshal(data map[string]any) []byte {
	bytes, _ := json.Marshal(data)
	return bytes
}
