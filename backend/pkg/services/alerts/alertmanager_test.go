// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"testing"
	"time"
)

type AlertManagerData struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Status string `json:"status"`
		Labels struct {
			Alertgroup string `json:"alertgroup"`
			Alertname  string `json:"alertname"`
			Device     string `json:"device"`
			Fstype     string `json:"fstype"`
			Instance   string `json:"instance"`
			Job        string `json:"job"`
			Mountpoint string `json:"mountpoint"`
			Nodename   string `json:"nodename"`
			Severity   string `json:"severity"`
		} `json:"labels"`
		Annotations struct {
			Description string `json:"description"`
			Summary     string `json:"summary"`
		} `json:"annotations"`
		StartsAt     time.Time `json:"startsAt"`
		EndsAt       time.Time `json:"endsAt"`
		GeneratorURL string    `json:"generatorURL"`
		Fingerprint  string    `json:"fingerprint"`
	} `json:"alerts"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	CommonLabels struct {
		Alertgroup string `json:"alertgroup"`
		Alertname  string `json:"alertname"`
		Fstype     string `json:"fstype"`
		Job        string `json:"job"`
		Severity   string `json:"severity"`
	} `json:"commonLabels"`
	CommonAnnotations struct {
	} `json:"commonAnnotations"`
	ExternalURL     string `json:"externalURL"`
	Version         string `json:"version"`
	GroupKey        string `json:"groupKey"`
	TruncatedAlerts int    `json:"truncatedAlerts"`
}

func TestMarshal(t *testing.T) {
	raw := "{\n    \"receiver\": \"alert-collector\",\n    \"status\": \"firing\",\n    \"alerts\": [\n        {\n            \"status\": \"firing\",\n            \"labels\": {\n                \"alertgroup\": \"NodeGroup\",\n                \"alertname\": \"HostOutOfDiskSpace\",\n                \"device\": \"/dev/mapper/centos_node--56-root\",\n                \"fstype\": \"xfs\",\n                \"instance\": \"192.168.1.56:9100\",\n                \"job\": \"node-exporter\",\n                \"mountpoint\": \"/\",\n                \"nodename\": \"node-56\",\n                \"severity\": \"warning\"\n            },\n            \"annotations\": {\n                \"description\": \"Disk is almost full (\\u003c 50% left)\\n  VALUE = 37.151284654372255\\n  LABELS = map[alertgroup:NodeGroup alertname:HostOutOfDiskSpace device:/dev/mapper/centos_node--56-root fstype:xfs instance:192.168.1.56:9100 job:node-exporter mountpoint:/ nodename:node-56 severity:warning]\",\n                \"summary\": \"Host out of disk space (instance 192.168.1.56:9100)\"\n            },\n            \"startsAt\": \"2024-07-23T17:27:00+08:00\",\n            \"endsAt\": \"0001-01-01T00:00:00Z\",\n            \"generatorURL\": \"http://Loyalty-Mac.local:8880/vmalert/alert?group_id=11181553531327151470\\u0026alert_id=6770942266677250940\",\n            \"fingerprint\": \"1195d06c3684af9b\"\n        },\n        {\n            \"status\": \"firing\",\n            \"labels\": {\n                \"alertgroup\": \"NodeGroup\",\n                \"alertname\": \"HostOutOfDiskSpace\",\n                \"device\": \"/dev/mapper/centos_node--56-root\",\n                \"fstype\": \"xfs\",\n                \"instance\": \"192.168.1.56:9100\",\n                \"job\": \"node-exporter\",\n                \"mountpoint\": \"/var/odigos\",\n                \"nodename\": \"node-56\",\n                \"severity\": \"warning\"\n            },\n            \"annotations\": {\n                \"description\": \"Disk is almost full (\\u003c 50% left)\\n  VALUE = 37.151284654372255\\n  LABELS = map[alertgroup:NodeGroup alertname:HostOutOfDiskSpace device:/dev/mapper/centos_node--56-root fstype:xfs instance:192.168.1.56:9100 job:node-exporter mountpoint:/var/odigos nodename:node-56 severity:warning]\",\n                \"summary\": \"Host out of disk space (instance 192.168.1.56:9100)\"\n            },\n            \"startsAt\": \"2024-07-23T17:27:00+08:00\",\n            \"endsAt\": \"0001-01-01T00:00:00Z\",\n            \"generatorURL\": \"http://Loyalty-Mac.local:8880/vmalert/alert?group_id=11181553531327151470\\u0026alert_id=14309186901432850141\",\n            \"fingerprint\": \"05742fd81562dd6a\"\n        },\n        {\n            \"status\": \"firing\",\n            \"labels\": {\n                \"alertgroup\": \"NodeGroup\",\n                \"alertname\": \"HostOutOfDiskSpace\",\n                \"device\": \"/dev/mapper/centos_large--server--6-root\",\n                \"fstype\": \"xfs\",\n                \"instance\": \"192.168.1.6:9100\",\n                \"job\": \"node-exporter\",\n                \"mountpoint\": \"/\",\n                \"nodename\": \"large-server-6\",\n                \"severity\": \"warning\"\n            },\n            \"annotations\": {\n                \"description\": \"Disk is almost full (\\u003c 50% left)\\n  VALUE = 45.78545737664876\\n  LABELS = map[alertgroup:NodeGroup alertname:HostOutOfDiskSpace device:/dev/mapper/centos_large--server--6-root fstype:xfs instance:192.168.1.6:9100 job:node-exporter mountpoint:/ nodename:large-server-6 severity:warning]\",\n                \"summary\": \"Host out of disk space (instance 192.168.1.6:9100)\"\n            },\n            \"startsAt\": \"2024-07-23T17:27:00+08:00\",\n            \"endsAt\": \"0001-01-01T00:00:00Z\",\n            \"generatorURL\": \"http://Loyalty-Mac.local:8880/vmalert/alert?group_id=11181553531327151470\\u0026alert_id=15853309511096313482\",\n            \"fingerprint\": \"8713d6093267d5f7\"\n        },\n        {\n            \"status\": \"firing\",\n            \"labels\": {\n                \"alertgroup\": \"NodeGroup\",\n                \"alertname\": \"HostOutOfDiskSpace\",\n                \"device\": \"/dev/mapper/centos_large--server--6-root\",\n                \"fstype\": \"xfs\",\n                \"instance\": \"192.168.1.6:9100\",\n                \"job\": \"node-exporter\",\n                \"mountpoint\": \"/var/odigos\",\n                \"nodename\": \"large-server-6\",\n                \"severity\": \"warning\"\n            },\n            \"annotations\": {\n                \"description\": \"Disk is almost full (\\u003c 50% left)\\n  VALUE = 45.78545737664876\\n  LABELS = map[alertgroup:NodeGroup alertname:HostOutOfDiskSpace device:/dev/mapper/centos_large--server--6-root fstype:xfs instance:192.168.1.6:9100 job:node-exporter mountpoint:/var/odigos nodename:large-server-6 severity:warning]\",\n                \"summary\": \"Host out of disk space (instance 192.168.1.6:9100)\"\n            },\n            \"startsAt\": \"2024-07-23T17:27:00+08:00\",\n            \"endsAt\": \"0001-01-01T00:00:00Z\",\n            \"generatorURL\": \"http://Loyalty-Mac.local:8880/vmalert/alert?group_id=11181553531327151470\\u0026alert_id=15216902388315119165\",\n            \"fingerprint\": \"943ffd72d4c8c718\"\n        }\n    ],\n    \"groupLabels\": {\n        \"alertname\": \"HostOutOfDiskSpace\"\n    },\n    \"commonLabels\": {\n        \"alertgroup\": \"NodeGroup\",\n        \"alertname\": \"HostOutOfDiskSpace\",\n        \"fstype\": \"xfs\",\n        \"job\": \"node-exporter\",\n        \"severity\": \"warning\"\n    },\n    \"commonAnnotations\": {},\n    \"externalURL\": \"http://Loyalty-Mac.local:9093\",\n    \"version\": \"4\",\n    \"groupKey\": \"{}:{alertname=\\\"HostOutOfDiskSpace\\\"}\",\n    \"truncatedAlerts\": 0\n}"
	var structedData AlertManagerData
	err := json.Unmarshal([]byte(raw), &structedData)
	if err != nil {
		t.Errorf("Error unmarshalling raw json to structed data: %v", err)
	}
	t.Logf("Structured data: %v", structedData)

	var mapData AlertManagerEvent
	err = json.Unmarshal([]byte(raw), &mapData)
	if err != nil {
		t.Errorf("Error unmarshalling raw json to mapped data: %v", err)
	}
	t.Logf("Mapped data: %v", mapData)
}
