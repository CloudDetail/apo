// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package alert

import (
	"fmt"
	"regexp"
)

func (a *Alert) GetStringTagWithRaw(key string) string {
	if value, find := a.EnrichTags[key]; find && len(value) > 0 {
		return value
	}

	if vPtr, find := a.Tags[key]; find {
		if value, ok := vPtr.(string); ok {
			return value
		}
	}
	return ""
}

func (a *Alert) GetTargetObj() string {
	if a.EnrichTags == nil {
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

func (a *Alert) GetServiceNameTag() string {
	if serviceName, find := a.EnrichTags["svc_name"]; find && len(serviceName) > 0 {
		return serviceName
	}
	return a.EnrichTags["serviceName"]
}

func (a *Alert) GetEndpointTag() string {
	if contentKey, find := a.EnrichTags["content_key"]; find && len(contentKey) > 0 {
		return contentKey
	}
	return a.EnrichTags["endpoint"]
}

// GetLevelTag 获取级别,当前只有network告警存在
func (a *Alert) GetLevelTag() string {
	return a.EnrichTags["level"]
}

func (a *Alert) GetNetSrcNodeTag() string {
	return a.EnrichTags["node_name"]
}

func (a *Alert) GetNetSrcPidTag() string {
	return a.EnrichTags["pid"]
}

func (a *Alert) GetPidTag() string {
	return a.EnrichTags["pid"]
}

func (a *Alert) GetNetSrcPodTag() string {
	//Compatible with older versions
	if pod, find := a.EnrichTags["src_pod"]; find && len(pod) > 0 {
		return pod
	}
	return a.EnrichTags["pod"]
}

func (a *Alert) GetK8sNamespaceTag() string {
	//Compatible with older versions
	if namespace, find := a.EnrichTags["src_namespace"]; find && len(namespace) > 0 {
		return namespace
	}
	return a.EnrichTags["namespace"]
}

func (a *Alert) GetK8sPodTag() string {
	if pod, find := a.EnrichTags["pod_name"]; find && len(pod) > 0 {
		return pod
	}
	return a.EnrichTags["pod"]
}

func (a *Alert) GetContainerTag() string {
	if container, find := a.EnrichTags["container_name"]; find && len(container) > 0 {
		return container
	}

	return a.GetStringTagWithRaw("container")
}

func (a *Alert) GetContainerIDTag() string {
	if containerID, find := a.EnrichTags["container_id"]; find && len(containerID) > 0 {
		return containerID
	}
	return a.GetStringTagWithRaw("container_id")
}

func (a *Alert) GetInfraNodeTag() string {
	//Compatible with older versions
	if node, find := a.EnrichTags["instance_name"]; find && len(node) > 0 {
		return node
	}
	return a.EnrichTags["node"]
}

func (a *Alert) GetNetSrcIPTag() string {
	//Compatible with older versions
	return a.GetStringTagWithRaw("src_ip")
}

func (a *Alert) GetNetDstIPTag() string {
	//Compatible with older versions
	return a.GetStringTagWithRaw("dst_ip")
}

var dbURLRegex = regexp.MustCompile(`tcp\((.+)\)`)
var dbIPRegex = regexp.MustCompile(`tcp\((\d+\.\d+\.\d+\.\d+):.*\)`)
var dbPortRegex = regexp.MustCompile(`tcp\(.*:(\d+)\)`)

func (a *Alert) GetDatabaseURL() string {
	if dbURL, find := a.EnrichTags["dbURL"]; find && len(dbURL) > 0 {
		return dbURL
	}

	if dbURL, find := a.EnrichTags["dbHost"]; find && len(dbURL) > 0 {
		return dbURL
	}

	if a.Group == "middleware" {
		instance := a.GetStringTagWithRaw("instance")
		return dbURLRegex.FindString(instance)
	}
	return ""
}

func (a *Alert) GetDatabaseIP() string {
	if dbIP := a.EnrichTags["dbIP"]; len(dbIP) > 0 {
		return dbIP
	}

	if a.Group == "middleware" {
		instance := a.GetStringTagWithRaw("instance")
		return dbIPRegex.FindString(instance)
	}
	return ""
}

func (a *Alert) GetDatabasePort() string {
	if dbPort := a.EnrichTags["dbPort"]; len(dbPort) > 0 {
		return dbPort
	}

	if a.Group == "middleware" {
		instance := a.GetStringTagWithRaw("instance")
		return dbPortRegex.FindString(instance)
	}
	return ""
}
