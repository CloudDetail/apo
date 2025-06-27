// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// AggPQLWithFilters generate PQL statements
// Generate PQL using vector and filterKVs
// @vector: Specify the aggregation time range
// @granularity: Specify aggregation granularity
// @filterKVs: filter condition, in the format of key1, value1, key2, and value2
type AggPQLWithFilters func(vector string, granularity string, filterKVs []string) string

func (repo *promRepo) QueryAggMetricsWithFilter(ctx core.Context, pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error) {
	if len(filterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of filterKVs is not even: %d", len(filterKVs))
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}
	vector := VecFromS2E(startTime, endTime)
	pql := pqlTemplate(vector, string(granularity), filters)
	return repo.QueryData(ctx, time.UnixMicro(endTime), pql)
}

func (repo *promRepo) QueryRangeAggMetricsWithFilter(ctx core.Context, pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, stepMicroS int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error) {
	if len(filterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of filterKVs is not even: %d", len(filterKVs))
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}

	step := time.Duration(stepMicroS) * time.Microsecond
	vector := VecFromDuration(step)
	pql := pqlTemplate(vector, string(granularity), filters)
	return repo.QueryRangeData(
		ctx,
		time.UnixMicro(startTime), time.UnixMicro(endTime),
		pql,
		step)
}

// #################### Decorator ######################

// Calculate the Day-over-Day Growth Rate rate of the metric.
func DayOnDay(pqlTemplate AggPQLWithFilters) AggPQLWithFilters {
	return func(vector string, granularity string, filterKVs []string) string {
		nowPql := pqlTemplate(vector, granularity, filterKVs)

		return `(` + nowPql + `) / ((` + nowPql + `) offset 24h )`
	}
}

// Calculate Week-over-Week Growth Rate of the metric.
func WeekOnWeek(pqlTemplate AggPQLWithFilters) AggPQLWithFilters {
	return func(vector string, granularity string, filterKVs []string) string {
		nowPql := pqlTemplate(vector, granularity, filterKVs)

		return `(` + nowPql + `) / ((` + nowPql + `) offset 7d )`
	}
}

func WithDefaultIFPolarisMetricExits(pqlTemplate AggPQLWithFilters, defaultValue int64) AggPQLWithFilters {
	return func(vector string, granularity string, filterKVs []string) string {
		pql := pqlTemplate(vector, granularity, filterKVs)
		checkPql := PQLIsPolarisMetricExitsWithFilters(vector, granularity, filterKVs)
		defaultV := strconv.FormatInt(defaultValue, 10)
		return `(` + pql + `) or ( ` + checkPql + ` * 0 + ` + defaultV + `)`
	}
}

// #################### PQL Templates #####################

// Average time spent on PQLAvgDepLatencyWithFilters queries from external dependencies
// Average time taken to return results for external dependencies
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

// Percentage of time spent by PQLDepLatencyRadioWithFilters queries from external dependencies.
// The percentage of the returned result that is externally dependent time to the total time consumed (0~1)
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

// PQLIsPolarisMetricExitsWithFilters uses the onCPU time in the Polaris indicator to determine whether the Polaris metric exists.
func PQLIsPolarisMetricExitsWithFilters(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	onCpuLatency := `sum by (` + granularity + `) (
        increase(kindling_profiling_cpu_duration_nanoseconds_sum{` + filtersStr + `}[` + vector + `])
	)`

	return onCpuLatency
}

// Average time consumption of PQLAvgLatencyWithFilters query
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

// Average error rate of PQLAvgErrorRateWithFilters query SQL requests
func PQLAvgErrorRateWithFilters(vector string, granularity string, filters []string) string {
	return avgErrorRateWithFilters(
		"kindling_span_trace_duration_nanoseconds_count",
		vector, granularity, filters)
}

// Average error rate of PQLAvgSQLErrorRateWithFilters query SQL requests
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
	// Used to retain a tag that exists in the requestCount but does not exist in the errorCount. Record that the request failure rate of this tag is 0
	return "(" + errorCount + "/" + requestCount + ") or (" + requestCount + " * 0)"
}

// Average TPS for PQLAvgTPSWithFilters query
func PQLAvgTPSWithFilters(vector string, granularity string, filters []string) string {
	return avgTPSWithFilters(
		"kindling_span_trace_duration_nanoseconds_count",
		vector, granularity, filters)
}

// Average TPS for PQLAvgTPSWithFilters query
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

/*
Using `* on` to join logparser_level_count/logparser_exception_count and span_trace_duration_count

It is mainly composed of the following exprs:

	( logparser_level_count + span_trace_duration_count ) left_join on(pod) span_trace_duration_count
	or
	( logparser_level_count + span_trace_duration_count ) left_join on(node,pid) span_trace_duration_count
*/
func PQLAvgLogErrorCountCombineEndpointsInfoWithFilters(vector string, granularity string, filters []string) string {
	errorLevelCount := `sum by (pod,node,pid) (increase(originx_logparser_level_count_total{level=~"error|critical"}[` + vector + `]))`
	exceptionCount := `sum by (pod,node,pid) (increase(originx_logparser_exception_count_total{}[` + vector + `]))`

	// ( errorLevelCount + exceptionCount ) or errorLevelCount or exceptionCount
	logErrorCount := "(((" + errorLevelCount + ") + (" + exceptionCount + ")) or (" + errorLevelCount + ") or (" + exceptionCount + "))"

	filtersStr := strings.Join(filters, ",")

	k8sSVCGroup := `group by(svc_name,content_key,pod) (last_over_time(kindling_span_trace_duration_nanoseconds_count{pod!="",` + filtersStr + `})[` + vector + `])`
	vmSVCGroup := `group by(svc_name,content_key,node,pid) (last_over_time(kindling_span_trace_duration_nanoseconds_count{pid!="",` + filtersStr + `})[` + vector + `])`

	return `sum by (` + granularity + `) ((` + logErrorCount + ` * on (pod) group_left (svc_name,content_key) ` + k8sSVCGroup + `) or ` +
		`(` + logErrorCount + ` * on (node,pid) group_left (svc_name,content_key) ` + vmSVCGroup + `))`
}

// PQLNormalLogCountWithFilters check for normal logs
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

// PQLMonitorStatus uptime-kuma monitoring item status
func PQLMonitorStatus(vector string, granularity string, filters []string) string {
	filtersStr := strings.Join(filters, ",")
	return `last_over_time(monitor_status{` + filtersStr + `}[` + vector + `])`
}
