package prometheus

import (
	"strings"
)

// PQLAvgDepLatencyWithFilters 查询来自外部依赖的平均耗时
// 返回结果为 外部依赖的评价耗时
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
	filtersStr := strings.Join(filters, ",")

	durationSum := `sum by (` + granularity + `) (increase(kindling_span_trace_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `]))`
	requestCount := `sum by (` + granularity + `) (increase(kindling_span_trace_duration_nanoseconds_count{` + filtersStr + `}[` + vector + `]))`

	return durationSum + "/" + requestCount
}

// PQLAvgErrorRateWithFilters 查询平均错误率
func PQLAvgErrorRateWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")

	var filterWithError string
	if len(filters) > 0 {
		filterWithError = filtersStr + `, is_error="true"`
	} else {
		filterWithError = `is_error="true"`
	}

	errorCount := `sum by (` + granularity + `) (increase(kindling_span_trace_duration_nanoseconds_count{` + filterWithError + `}[` + vector + `]))`
	requestCount := `sum by (` + granularity + `) (increase(kindling_span_trace_duration_nanoseconds_count{` + filtersStr + `}[` + vector + `]))`

	// ( errorCount or requestCount * 0 ) / requestCount
	// 用于保留requestCount中存在而errorCount中不存在的标签,记录该标签的请求失败率为0
	return "(" + errorCount + "/" + requestCount + ") or (" + requestCount + " * 0)"
}

// PQLAvgTPSWithFilters 查询平均TPS
func PQLAvgTPSWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	return `sum(rate(kindling_span_trace_duration_nanoseconds_count{` + filtersStr + `}[` + vector + `])) by(` + granularity + `)`
}
