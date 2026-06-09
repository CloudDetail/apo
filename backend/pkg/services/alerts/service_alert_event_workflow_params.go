// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type alertRunOnceParams struct {
	AlertName    string `json:"alertName"`
	Pod          string `json:"pod"`
	Namespace    string `json:"namespace"`
	NodeName     string `json:"nodeName"`
	ContainerID  string `json:"containerId"`
	Pid          string `json:"pid"`
	InstanceName string `json:"instanceName"`
	Service      string `json:"service"`
	AlertEventID string `json:"alertEventId"`
}

func buildAlertRunOnceParams(event *alert.AlertEvent, alertEventID string, service string) string {
	params := alertRunOnceParams{
		AlertName:    event.Name,
		Pod:          getAlertParamValue(event, "pod", "src_pod", "pod_name"),
		Namespace:    getAlertParamValue(event, "namespace", "src_namespace"),
		NodeName:     getAlertParamValue(event, "nodename", "node_name", "nodeName", "node", "instance_name"),
		ContainerID:  getAlertParamValue(event, "containerId", "container_id", "containerID"),
		Pid:          getAlertParamValue(event, "pid", "process_id", "processId"),
		InstanceName: getAlertParamValue(event, "instanceName", "instance_name"),
		Service:      service,
		AlertEventID: alertEventID,
	}

	jsonStr, err := json.Marshal(params)
	if err != nil {
		return "{}"
	}
	return string(jsonStr)
}

func getAlertParamValue(event *alert.AlertEvent, keys ...string) string {
	for _, key := range keys {
		if value := event.EnrichTags[key]; value != "" {
			return value
		}
	}
	for _, key := range keys {
		if value, ok := event.Tags[key]; ok && value != nil {
			if str, ok := value.(string); ok {
				return str
			}
			return fmt.Sprint(value)
		}
	}
	return ""
}
