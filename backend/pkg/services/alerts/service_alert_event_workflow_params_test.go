// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func TestFillWorkflowParamsKeepsOnlyRunOnceInputFields(t *testing.T) {
	now := time.Unix(1710000000, 0)
	record := &alert.AEventWithWRecord{
		AlertEvent: alert.AlertEvent{
			Alert: alert.Alert{
				AlertID: "alert-001",
				Name:    "HighCPU",
				EnrichTags: map[string]string{
					"container_id":  "container-001",
					"instance_name": "node-from-enrich",
					"namespace":     "default",
					"node":          "node-001",
					"pid":           "1234",
					"pod_name":      "checkout-7c9",
					"serviceName":   "checkout",
					"unused":        "large-unused-tag",
				},
				Tags: alert.RawTags{
					"instance_name": "node-from-raw",
					"unused_raw":    "large-unused-raw-tag",
				},
			},
			EventID:    "event-001",
			Detail:     "large alert detail",
			Status:     alert.StatusFiring,
			CreateTime: now.Add(-time.Hour),
			UpdateTime: now,
		},
		WorkflowDetail: alert.WorkflowDetail{
			Input:    "checked-event-001",
			Validity: "unknown",
		},
	}

	(&service{}).fillWorkflowParams(core.EmptyCtx(), record)

	var params map[string]string
	if err := json.Unmarshal([]byte(record.WorkflowParams.Params), &params); err != nil {
		t.Fatalf("workflow params JSON = %q, unmarshal error = %v", record.WorkflowParams.Params, err)
	}

	want := map[string]string{
		"alertEventId": "checked-event-001",
		"alertName":    "HighCPU",
		"containerId":  "container-001",
		"instanceName": "node-from-enrich",
		"namespace":    "default",
		"nodeName":     "node-001",
		"pid":          "1234",
		"pod":          "checkout-7c9",
		"service":      "checkout",
	}
	if !reflect.DeepEqual(params, want) {
		t.Fatalf("workflow params = %#v, want %#v", params, want)
	}
}
