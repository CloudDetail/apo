package prometheus

import (
	"fmt"
	"strings"
)

const (
	// 此处没有乘以100， 由其他地方乘以100
	SQL_ERROR_RATE_INSTANCE = "(" +
		"(sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s, is_error='true'}[%s])) or 0)" + // or 0补充缺失数据场景
		"/sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))" +
		") or (sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s])) * 0)" // or * 0补充中间缺失数据的场景
)

// PQLAvgDepLatencyWithFilters 查询来自外部依赖的平均耗时
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

// PQLAvgLatencyWithFilters 查询自身平均耗时
func PQLAvgLatencyWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	return `sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
) by(` + granularity + `)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{` + filtersStr + `}[` + vector + `]
  )
) by(` + granularity + `)`
}

// PQLAvgErrorRateWithFilters 查询平均错误率
func PQLAvgErrorRateWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")

	return fmt.Sprintf(SQL_ERROR_RATE_INSTANCE,
		granularity, filtersStr, vector,
		granularity, filtersStr, vector,
		granularity, filtersStr, vector,
	)
}

// PQLAvgTPSWithFilters 查询平均TPS
func PQLAvgTPSWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	return `avg(
	  rate(kindling_span_trace_duration_nanoseconds_count{` + filtersStr + `}[` + vector + `])
	) by(` + granularity + `)`
}
