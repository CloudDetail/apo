package prometheus

import (
	"fmt"
	"strings"
)

// PQLAvgDepLatencyWithFilters 查询来自外部依赖的平均耗时
// 返回结果为 外部依赖的平均耗时
func PQLAvgDepLatencyWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	allDepNetworkLatency := `sum by (` + granularity + `) (
        increase(kindling_profiling_epoll_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
		+
        increase(kindling_profiling_net_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
	)`
	allRequestCount := `sum by (` + granularity + `) (
        increase(kindling_span_trace_duration_nanoseconds_count{` + filtersStr + `}[` + vector + `])
	)`

	return allDepNetworkLatency + "/" + allRequestCount
}

// PQLDepLatencyRadioWithFilters 查询来自外部依赖的耗时占比
// 返回结果为 外部依赖的耗时占总耗时的比例 (0~1)
func PQLDepLatencyRadioWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	allDepNetworkLatency := `sum by (` + granularity + `) (
        increase(kindling_profiling_epoll_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
		+
        increase(kindling_profiling_net_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
	)`
	allRequestLatencySum := `sum by (` + granularity + `) (
        increase(kindling_span_trace_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
	)`

	return allDepNetworkLatency + "/" + allRequestLatencySum
}

// PQLIsPolarisMetricExitsWithFilters 采用北极星指标中的onCPU耗时判断是否存在北极星指标
func PQLIsPolarisMetricExitsWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	onCpuLatency := `sum by (` + granularity + `) (
        increase(kindling_profiling_cpu_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
	)`

	return onCpuLatency
}

// PQLAvgLatencyWithFilters 查询自身平均耗时
func PQLAvgLatencyWithFilters(vector string, granularity string, filters []string) string {
	return avgLatencyWithFilters(
		"kindling_span_trace_duration_nanoseconds_sum",
		"kindling_span_trace_duration_nanoseconds_count",
		vector, granularity, filters)
}

func PQLAvgSQLLatencyWithFilters(vector string, granularity string, filters []string) string {
	return avgLatencyWithFilters(
		"kindling_db_duration_nanoseconds_sum",
		"kindling_db_duration_nanoseconds_count",
		vector, granularity, filters)
}

func avgLatencyWithFilters(sumMetric string, countMetric string, vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	durationSum := `sum by (` + granularity + `) (increase(` + sumMetric + `{` + filtersStr + `}[` + vector + `]))`
	requestCount := `sum by (` + granularity + `) (increase(` + countMetric + `{` + filtersStr + `}[` + vector + `]))`

	return durationSum + "/" + requestCount
}

// PQLAvgErrorRateWithFilters 查询SQL请求的平均错误率
func PQLAvgErrorRateWithFilters(vector string, granularity string, filters []string) string {
	return avgErrorRateWithFilters(
		"kindling_span_trace_duration_nanoseconds_count",
		vector, granularity, filters)
}

// PQLAvgSQLErrorRateWithFilters 查询SQL请求的平均错误率
func PQLAvgSQLErrorRateWithFilters(vector string, granularity string, filters []string) string {
	return avgErrorRateWithFilters(
		"kindling_db_duration_nanoseconds_count",
		vector, granularity, filters)
}

func avgErrorRateWithFilters(metric string, vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")

	var filterWithError string
	if len(filters) > 0 {
		filterWithError = filtersStr + `, is_error="true"`
	} else {
		filterWithError = `is_error="true"`
	}

	errorCount := `sum by (` + granularity + `) (increase(` + metric + `{` + filterWithError + `}[` + vector + `]))`
	requestCount := `sum by (` + granularity + `) (increase(` + metric + `{` + filtersStr + `}[` + vector + `]))`

	// ( errorCount or requestCount * 0 ) / requestCount
	// 用于保留requestCount中存在而errorCount中不存在的标签,记录该标签的请求失败率为0
	return "(" + errorCount + "/" + requestCount + ") or (" + requestCount + " * 0)"
}

// PQLAvgTPSWithFilters 查询平均TPS
func PQLAvgTPSWithFilters(vector string, granularity string, filters []string) string {
	return avgTPSWithFilters(
		"kindling_span_trace_duration_nanoseconds_count",
		vector, granularity, filters)
}

// PQLAvgTPSWithFilters 查询平均TPS
func PQLAvgSQLTPSWithFilters(vector string, granularity string, filters []string) string {
	return avgTPSWithFilters(
		"kindling_db_duration_nanoseconds_count",
		vector, granularity, filters)
}

func avgTPSWithFilters(metric string, vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	return `sum(rate(` + metric + `{` + filtersStr + `}[` + vector + `])) by(` + granularity + `)`
}

func PQLAvgLogErrorCountWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")

	var filterWithError string
	if len(filters) > 0 {
		filterWithError = filtersStr + `, level=~"error|critical"`
	} else {
		filterWithError = `level=~"error|critical"`
	}

	errorLevelCount := `sum by (` + granularity + `) (increase(originx_logparser_level_count_total{` + filterWithError + `}[` + vector + `]))`
	exceptionCount := `sum by (` + granularity + `) (increase(originx_logparser_exception_count_total{` + filtersStr + `}[` + vector + `]))`

	// ( errorLevelCount + exceptionCount ) or errorLevelCount or exceptionCount
	return "((" + errorLevelCount + ") + (" + exceptionCount + ")) or (" + errorLevelCount + ") or (" + exceptionCount + ")"
}

// PQLNormalLogCountWithFilters 检查有无正常日志
func PQLNormalLogCountWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")

	var filterWithLevel string
	if len(filters) > 0 {
		filterWithLevel = filtersStr + `, level=~".*"`
	} else {
		filterWithLevel = `level=~".*"`
	}
	errorLevelCount := `sum by (` + granularity + `) (increase(originx_logparser_level_count_total{` + filterWithLevel + `}[` + vector + `]))`
	exceptionCount := `sum by (` + granularity + `) (increase(originx_logparser_exception_count_total{` + filtersStr + `}[` + vector + `]))`

	return "((" + errorLevelCount + ") + (" + exceptionCount + ")) or (" + errorLevelCount + ") or (" + exceptionCount + ")"
}

// PQLMonitorStatus uptime-kuma监控项状态
func PQLMonitorStatus(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	return `last_over_time(monitor_status{` + filtersStr + `}[` + vector + `])`
}

// PQLInstanceLog 获取instance级别log指标的pql pod or vm
func PQLInstanceLog(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, podFilterKVs, vmFilterKVs []string) (string, error) {
	if len(podFilterKVs)%2 != 0 {
		return "", fmt.Errorf("size of podFilterKVs is not even: %d", len(podFilterKVs))
	}

	if len(vmFilterKVs)%2 != 0 {
		return "", fmt.Errorf("size of vmFilterKVs is not even: %d", len(vmFilterKVs))
	}
	var podFilters []string
	for i := 0; i+1 < len(podFilterKVs); i += 2 {
		podFilters = append(podFilters, fmt.Sprintf("%s\"%s\"", podFilterKVs[i], podFilterKVs[i+1]))
	}

	var vmFilters []string
	for i := 0; i+1 < len(vmFilterKVs); i += 2 {
		vmFilters = append(vmFilters, fmt.Sprintf("%s\"%s\"", vmFilterKVs[i], vmFilterKVs[i+1]))
	}

	vector := VecFromS2E(startTime, endTime)
	podPql := pqlTemplate(vector, string(granularity), podFilters)
	vmPql := pqlTemplate(vector, string(granularity), vmFilters)
	return `(` + podPql + `) or (` + vmPql + `)`, nil
}
