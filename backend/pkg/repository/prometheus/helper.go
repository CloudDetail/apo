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
	ContentKeyPQLFilter      = "content_key="
	ContentKeyRegexPQLFilter = "content_key=~"
	ServicePQLFilter         = "svc_name="
	ServiceRegexPQLFilter    = "svc_name=~"
	ContainerIdPQLFilter     = "container_id="
	IsErrorPQLFilter         = "is_error="

	ValueExistPQLValueFilter = ".+"
	LabelExistPQLValueFilter = ".*"
)

type Granularity string

const (
	SVCGranularity      Granularity = "svc_name"
	EndpointGranularity Granularity = "svc_name, content_key"
)

// AggPQLWithFilters 生成PQL语句
// 使用vector和filterKVs生成PQL
// @ vector: 指定聚合时间范围
// @ granularity: 指定聚合粒度
// @ filterKVs: 过滤条件, 格式为 key1, value1, key2, value2
type AggPQLWithFilters func(vector string, granularity string, filterKVs []string) string

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

// MultipleValue 创建匹配多个目标值的正则表达式
// 需要配合 xxxRegexPQLFilter 使用
func MultipleValue(key ...string) string {
	escapedKeys := make([]string, len(key))
	for i, key := range key {
		escapedKeys[i] = EscapeRegexp(key)
	}
	// 使用 strings.Join 生成正则表达式模式
	return strings.Join(escapedKeys, "|")
}
