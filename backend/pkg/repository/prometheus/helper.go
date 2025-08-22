// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var regexpEscape = regexp.MustCompile(`([\\\^\$\.\|\?\*\+\(\)\[\]\{\}])`)

func EscapeRegexp(s string) string {
	return regexpEscape.ReplaceAllString(s, `\\$1`)
}

const (
	ContentKeyPQLFilter     = "content_key="
	ServicePQLFilter        = "svc_name="
	NamespacePQLFilter      = "namespace="
	NamespaceRegexPQLFilter = "namespace=~"
	ContainerIdPQLFilter    = "container_id="
	IsErrorPQLFilter        = "is_error="
	PodPQLFilter            = "pod="
	PidPQLFilter            = "pid="
	NodeNamePQLFilter       = "node_name="

	ContentKeyRegexPQLFilter = "content_key=~"
	ServiceRegexPQLFilter    = "svc_name=~"
	DBNameRegexPQLFilter     = "name=~"

	ValueExistPQLValueFilter = ".+"
	LabelExistPQLValueFilter = ".*"

	PodRegexPQLFilter           = "pod=~"
	LogMetricPodRegexPQLFilter  = "pod_name=~"
	LogMetricNodeRegexPQLFilter = "host_name=~"
	LogMetricPidRegexPQLFilter  = "pid=~"
	ContainerIdRegexPQLFilter   = "container_id=~"
	PidRegexPQLFilter           = "pid=~"

	ClusterIDKey   = "cluster_id"
	ServiceNameKey = "svc_name"
	ContentKeyKey  = "content_key"
	NamespaceKey   = "namespace"
)

type Granularity string

const (
	SVCGranularity              Granularity = "svc_name"
	EndpointGranularity         Granularity = "svc_name, content_key"
	NSEndpointGranularity       Granularity = "namespace, svc_name, content_key"
	InstanceEndpointGranularity Granularity = "svc_name, content_key, container_id, node_name, pid, pod, namespace, node_ip, cluster_id"
	InstanceGranularity         Granularity = "svc_name, container_id, node_name, pid, pod, namespace, node_ip, cluster_id"
	LogGranularity              Granularity = "pid,host_name,host_ip,container_id,pod_name,namespace"
	DBOperationGranularity      Granularity = "svc_name, db_system, db_name, name, db_url"

	DBInstanceGranularity Granularity = "db_url,container_id,node_name,pid,cluster_id"
)

/*
VecFromS2E 根据起止时间戳获取时间范围
用于PromQL查询时的内置聚合

e.g. avg (xxx[${vec}])
*/
func VecFromS2E(startTime int64, endTime int64) (vec string) {
	durationNS := (endTime - startTime) * int64(time.Microsecond)
	if durationNS > int64(time.Minute) {
		vec = strconv.FormatInt(durationNS/int64(time.Minute), 10) + "m"
	} else {
		vec = strconv.FormatInt(durationNS/int64(time.Second), 10) + "s"
	}
	return vec
}

func VecFromDuration(duration time.Duration) (vec string) {
	durationNS := duration.Nanoseconds()
	if durationNS > int64(time.Minute) {
		vec = strconv.FormatInt(durationNS/int64(time.Minute), 10) + "m"
	} else {
		vec = strconv.FormatInt(durationNS/int64(time.Second), 10) + "s"
	}
	return vec
}

// RegexMultipleValue create a regular expression that matches multiple target values
// need to be used with xxxRegexPQLFilter
func RegexMultipleValue(key ...string) string {
	escapedKeys := make([]string, len(key))
	for i, key := range key {
		escapedKeys[i] = EscapeRegexp(key)
	}
	// Generate regex patterns using strings.Join
	return strings.Join(escapedKeys, "|")
}

// RegexContainsValue create a regular expression that matches a single target value
// need to be used with xxxRegexPQLFilter
func RegexContainsValue(key string) string {
	escapedKey := EscapeRegexp(key)
	return ".*" + escapedKey + ".*"
}
