// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"fmt"
	"regexp"
)

func (a *AlertEvent) GetTargetObj() string {
	if a.Tags == nil {
		return ""
	}
	switch a.Group {
	case "app":
		return a.GetServiceNameTag()
	case "infra":
		return a.GetInfraNodeTag()
	case "network":
		return fmt.Sprintf("%s->%s", a.GetNetSrcIPTag(), a.GetNetDstIPTag())
	case "container":
		return fmt.Sprintf("%s(%s)", a.GetK8sPodTag(), a.GetContainerTag())
	case "middleware":
		return fmt.Sprintf("%s(%s:%s)",
			a.GetDatabaseURL(),
			a.GetDatabaseIP(),
			a.GetDatabasePort())
	}
	return ""
}

func (a *AlertEvent) GetServiceNameTag() string {
	if serviceName, find := a.Tags["svc_name"]; find && len(serviceName) > 0 {
		return serviceName
	}
	return a.Tags["serviceName"]
}

func (a *AlertEvent) GetEndpointTag() string {
	if contentKey, find := a.Tags["content_key"]; find && len(contentKey) > 0 {
		return contentKey
	}
	return a.Tags["endpoint"]
}

// GetLevelTag 获取级别,当前只有network告警存在
func (a *AlertEvent) GetLevelTag() string {
	return a.Tags["level"]
}

func (a *AlertEvent) GetNetSrcNodeTag() string {
	return a.Tags["node_name"]
}

func (a *AlertEvent) GetNetSrcPidTag() string {
	return a.Tags["pid"]
}

func (a *AlertEvent) GetNetSrcPodTag() string {
	//Compatible with older versions
	if pod, find := a.Tags["src_pod"]; find && len(pod) > 0 {
		return pod
	}
	return a.Tags["pod"]
}

func (a *AlertEvent) GetK8sNamespaceTag() string {
	//Compatible with older versions
	if namespace, find := a.Tags["src_namespace"]; find && len(namespace) > 0 {
		return namespace
	}
	return a.Tags["namespace"]
}

func (a *AlertEvent) GetK8sPodTag() string {
	if pod, find := a.Tags["pod_name"]; find && len(pod) > 0 {
		return pod
	}
	return a.Tags["pod"]
}

func (a *AlertEvent) GetContainerTag() string {
	if container, find := a.Tags["container_name"]; find && len(container) > 0 {
		return container
	}
	return a.RawTags["container"]
}

func (a *AlertEvent) GetInfraNodeTag() string {
	//Compatible with older versions
	if node, find := a.Tags["instance_name"]; find && len(node) > 0 {
		return node
	}
	return a.Tags["node"]
}

func (a *AlertEvent) GetNetSrcIPTag() string {
	//Compatible with older versions
	if ip, find := a.RawTags["src_ip"]; find && len(ip) > 0 {
		return ip
	}
	return a.RawTags["src_ip"]
}

func (a *AlertEvent) GetNetDstIPTag() string {
	//Compatible with older versions
	if ip, find := a.RawTags["dst_ip"]; find && len(ip) > 0 {
		return ip
	}
	return a.RawTags["dst_ip"]
}

var dbURLRegex = regexp.MustCompile(`tcp\((.+)\)`)
var dbIPRegex = regexp.MustCompile(`tcp\((\d+\.\d+\.\d+\.\d+):.*\)`)
var dbPortRegex = regexp.MustCompile(`tcp\(.*:(\d+)\)`)

func (a *AlertEvent) GetDatabaseURL() string {
	if dbURL, find := a.Tags["dbURL"]; find && len(dbURL) > 0 {
		return dbURL
	}

	if dbURL, find := a.Tags["dbHost"]; find && len(dbURL) > 0 {
		return dbURL
	}

	if a.Group == "middleware" {
		instance := a.RawTags["instance"]
		return dbURLRegex.FindString(instance)
	}
	return ""
}

func (a *AlertEvent) GetDatabaseIP() string {
	if dbIP := a.Tags["dbIP"]; len(dbIP) > 0 {
		return dbIP
	}

	if a.Group == "middleware" {
		instance := a.RawTags["instance"]
		return dbIPRegex.FindString(instance)
	}
	return ""
}

func (a *AlertEvent) GetDatabasePort() string {
	if dbPort := a.Tags["dbPort"]; len(dbPort) > 0 {
		return dbPort
	}

	if a.Group == "middleware" {
		instance := a.RawTags["instance"]
		return dbPortRegex.FindString(instance)
	}
	return ""
}
