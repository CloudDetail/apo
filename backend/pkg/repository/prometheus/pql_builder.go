// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// this file is using @PQLFilter and @PQLTemplate to build more complex PQL statements

type QueryWithPQLFilter interface {
	QueryMetricsWithPQLFilter(ctx core.Context, pqlTpl PQLTemplate, startTime int64, endTime int64, gran Granularity, filter PQLFilter) ([]MetricResult, error)
	QueryRangeMetricsWithPQLFilter(ctx core.Context, pqlTpl PQLTemplate, startTime int64, endTime int64, stepMicroS int64, gran Granularity, filter PQLFilter) ([]MetricResult, error)
}

type PQLTemplate func(vector string, gran string, filter PQLFilter, offset string) string

func (repo *promRepo) QueryMetricsWithPQLFilter(ctx core.Context, pqlTpl PQLTemplate, startTime int64, endTime int64, gran Granularity, filter PQLFilter) ([]MetricResult, error) {
	rng := VecFromS2E(startTime, endTime)
	pql := pqlTpl(rng, string(gran), filter, "")
	return repo.QueryData(ctx, time.UnixMicro(endTime), pql)
}

func (repo *promRepo) QueryRangeMetricsWithPQLFilter(ctx core.Context, pqlTpl PQLTemplate, startTime int64, endTime int64, stepMicroS int64, gran Granularity, filter PQLFilter) ([]MetricResult, error) {
	step := time.Duration(stepMicroS) * time.Microsecond
	vector := VecFromDuration(step)
	pql := pqlTpl(vector, string(gran), filter, "")
	return repo.QueryRangeData(ctx, time.UnixMicro(startTime), time.UnixMicro(endTime), pql, step)
}

// ####################### PQL Template #######################

const (
	PROFILING_EPOLL_DURATION_SUM = "kindling_profiling_epoll_duration_nanoseconds_sum"
	PROFILING_NET_DURATION_SUM   = "kindling_profiling_net_duration_nanoseconds_sum"
	PROFILING_CPU_DURATION_SUM   = "kindling_profiling_cpu_duration_nanoseconds_sum"

	SPAN_TRACE_COUNT        = "kindling_span_trace_duration_nanoseconds_count"
	SPAN_TRACE_DURATION_SUM = "kindling_span_trace_duration_nanoseconds_sum"

	SPAN_DB_COUNT        = "kindling_db_duration_nanoseconds_count"
	SPAN_DB_DURATION_SUM = "kindling_db_duration_nanoseconds_sum"

	LOG_LEVEL_COUNT     = "originx_logparser_level_count_total"
	LOG_EXCEPTION_COUNT = "originx_logparser_exception_count_total"

	MONITOR_STATUS = "monitor_status"
)

func PQLMetricSeries(metric string) PQLTemplate {
	return func(rng, gran string, filter PQLFilter, _ string) string {
		// using lastOverTime to get full series exited during range
		return groupBy(gran, lastOverTime(rangeVec(metric, filter, rng, "")))
	}
}

func PQLAvgDepLatencyWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	allDepNetworkLatency := sumBy(gran,
		add(
			increase(rangeVec(PROFILING_EPOLL_DURATION_SUM, filter, rng, offset)),
			increase(rangeVec(PROFILING_NET_DURATION_SUM, filter, rng, offset)),
		),
	)
	allRequestCount := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filter, rng, offset)))
	return divWithDef(allDepNetworkLatency, allRequestCount, allRequestCount, "0")
}

// Percentage of time spent by PQLDepLatencyRadioWithFilters queries from external dependencies.
// The percentage of the returned result that is externally dependent time to the total time consumed (0~1)
func PQLDepLatencyRadioWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	allDepNetworkLatency := sumBy(gran,
		add(
			increase(rangeVec(PROFILING_EPOLL_DURATION_SUM, filter, rng, offset)),
			increase(rangeVec(PROFILING_NET_DURATION_SUM, filter, rng, offset)),
		),
	)
	allRequestLatencySum := sumBy(gran,
		increase(rangeVec(SPAN_TRACE_DURATION_SUM, filter, rng, offset)),
	)
	return divWithDef(allDepNetworkLatency, allRequestLatencySum, allRequestLatencySum, "0")
}

// PQLIsPolarisMetricExitsWithFilters uses the onCPU time in the Polaris indicator to determine whether the Polaris metric exists.
func PQLPolarisActiveSeries(rng string, gran string, filters PQLFilter, offset string) string {
	return sumBy(gran,
		increase(rangeVec(PROFILING_CPU_DURATION_SUM, filters, rng, offset)),
	)
}

// Average time consumption of PQLAvgLatencyWithFilters query
func PQLAvgLatencyWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	durationSum := sumBy(gran, increase(rangeVec(SPAN_TRACE_DURATION_SUM, filter, rng, offset)))
	requestCount := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filter, rng, offset)))
	return div(durationSum, requestCount)
}

func PQLAvgSQLLatencyWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	durationSum := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filter, rng, offset)))
	requestCount := sumBy(gran, increase(rangeVec(SPAN_DB_COUNT, filter, rng, offset)))
	return div(durationSum, requestCount)
}

// Average error rate of PQLAvgErrorRateWithFilters query SQL requests
func PQLAvgErrorRateWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	filterWithError := filter.Clone().Equal("is_error", "true")

	errorCount := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filterWithError, rng, offset)))
	requestCount := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filter, rng, offset)))

	// ( errorCount or requestCount * 0 ) / requestCount
	// Used to retain a tag that exists in the requestCount but does not exist in the errorCount. Record that the request failure rate of this tag is 0
	return divWithDef(errorCount, requestCount, requestCount, "0")
}

// Average error rate of PQLAvgSQLErrorRateWithFilters query SQL requests
func PQLAvgSQLErrorRateWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	filterWithError := filter.Clone().Equal("is_error", "true")

	errorCount := sumBy(gran, increase(rangeVec(SPAN_DB_COUNT, filterWithError, rng, offset)))
	requestCount := sumBy(gran, increase(rangeVec(SPAN_DB_COUNT, filter, rng, offset)))

	// ( errorCount or requestCount * 0 ) / requestCount
	// Used to retain a tag that exists in the requestCount but does not exist in the errorCount. Record that the request failure rate of this tag is 0
	return divWithDef(errorCount, requestCount, requestCount, "0")
}

// Average TPS for PQLAvgTPSWithFilters query
func PQLAvgTPSWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	return sumBy(gran, rate(rangeVec(SPAN_TRACE_COUNT, filter, rng, offset)))
}

// Average TPS for PQLAvgTPSWithFilters query
func PQLAvgSQLTPSWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	return sumBy(gran, rate(rangeVec(SPAN_DB_COUNT, filter, rng, offset)))
}

func PQLAvgLogErrorCountWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	filterWithError := filter.Clone().RegexMatch("level", "error|critical")

	errorLevelCount := sumBy(gran, increase(rangeVec(LOG_LEVEL_COUNT, filterWithError, rng, offset)))
	exceptionCount := sumBy(gran, increase(rangeVec(LOG_EXCEPTION_COUNT, filter, rng, offset)))

	// ( errorLevelCount + exceptionCount ) or errorLevelCount or exceptionCount
	return addWithDef(errorLevelCount, exceptionCount, errorLevelCount, exceptionCount)
}

func PQLAvgLogErrorCountCombineEndpointsInfoWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	errorLevelCount := sumBy("pod,node,pid",
		increase(rangeVec(LOG_LEVEL_COUNT, RegexMatchFilter("level", "error|critical"), rng, offset)))

	exceptionCount := sumBy("pod,node,pid",
		increase(rangeVec(LOG_EXCEPTION_COUNT, nil, rng, offset)))

	// ( errorLevelCount + exceptionCount ) or errorLevelCount or exceptionCount
	logErrorCount := addWithDef(errorLevelCount, exceptionCount, errorLevelCount, exceptionCount)

	k8sSVCGroup := groupBy("svc_name,content_key,pod",
		lastOverTime(rangeVec(SPAN_TRACE_COUNT, filter.Clone().NotEqual("pod", ""), rng, offset)))

	vmSVCGroup := groupBy("svc_name,content_key,node,pid",
		lastOverTime(rangeVec(SPAN_TRACE_COUNT, filter.Clone().Equal("pod", "").NotEqual("pid", ""), rng, offset)))

	return add(
		sumBy(gran, labelLeftOn(logErrorCount, "pod", "svc_name,content_key", k8sSVCGroup)),
		sumBy(gran, labelLeftOn(logErrorCount, "node,pid", "svc_name,content_key", vmSVCGroup)),
	)
}

func PQLNormalLogCountWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	logCount := sumBy(gran, increase(rangeVec(LOG_LEVEL_COUNT, filter, rng, offset)))
	exceptionCount := sumBy(gran, increase(rangeVec(LOG_EXCEPTION_COUNT, filter, rng, offset)))
	return addWithDef(logCount, exceptionCount, logCount, exceptionCount)
}

func PQLMonitorStatusWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	return lastOverTime(rangeVec(MONITOR_STATUS, filter, rng, offset))
}

// ####################### vector Expr #########################

// using VM extended syntax by default
var StrictPQL = false

func EnableStrictPQL() {
	StrictPQL = true
}

type _vector struct {
	__name__   string
	instantVec PQLFilter
	modifier   string
	offset     string
}

type _rangeVec struct {
	_vector
	rangeVec string
}

func (v _vector) String() string {
	if StrictPQL {
		return strings.Join(v.instantVec._strictPQL(v.__name__, v.modifier, v.offset, ""), " or ")
	}

	var sb strings.Builder
	sb.WriteString(v.__name__)
	if v.instantVec != nil {
		sb.WriteByte('{')
		sb.WriteString(v.instantVec.String())
		sb.WriteByte('}')
	}
	if len(v.modifier) > 0 {
		sb.WriteString(" @ ")
		sb.WriteString(v.modifier)
	}
	if len(v.offset) > 0 {
		sb.WriteString(" ")
		sb.WriteString(v.offset)
	}
	return sb.String()
}

func (v _rangeVec) String() string {
	if StrictPQL {
		return strings.Join(v.instantVec._strictPQL(v.__name__, v.modifier, v.offset, v.rangeVec), " or ")
	}

	var sb strings.Builder
	sb.WriteString(v.__name__)
	if v.instantVec != nil {
		sb.WriteByte('{')
		sb.WriteString(v.instantVec.String())
		sb.WriteByte('}')
	}
	if len(v.rangeVec) > 0 {
		sb.WriteByte('[')
		sb.WriteString(v.rangeVec)
		sb.WriteByte(']')
	}
	if len(v.modifier) > 0 {
		sb.WriteString(" @ ")
		sb.WriteString(v.modifier)
	}
	if len(v.offset) > 0 {
		sb.WriteString(" ")
		sb.WriteString(v.offset)
	}
	return sb.String()
}

func vector(metric string, filter PQLFilter) _vector {
	return _vector{
		__name__:   metric,
		instantVec: filter,
	}
}

func offsetVec(metric string, filter PQLFilter, offset string) _vector {
	return _vector{
		__name__:   metric,
		instantVec: filter,
		offset:     offset,
	}
}

func modifyVec(metric string, filter PQLFilter, modifier string) _vector {
	return _vector{
		__name__:   metric,
		instantVec: filter,
		modifier:   modifier,
	}
}

func rangeVec(metric string, filter PQLFilter, rangeVec string, offset string) _rangeVec {
	return _rangeVec{
		_vector: _vector{
			__name__:   metric,
			instantVec: filter,
			offset:     offset,
		},
		rangeVec: rangeVec,
	}
}

func increase(vec _rangeVec) string {
	var sb strings.Builder
	sb.WriteString("increase(")
	sb.WriteString(vec.String())
	sb.WriteString(")")
	return sb.String()
}

func rate(vec _rangeVec) string {
	var sb strings.Builder
	sb.WriteString("rate(")
	sb.WriteString(vec.String())
	sb.WriteString(")")
	return sb.String()
}

func lastOverTime(vec _rangeVec) string {
	var sb strings.Builder
	sb.WriteString("last_over_time(")
	sb.WriteString(vec.String())
	sb.WriteString(")")
	return sb.String()
}

// sum by (@gran) (a)
func sumBy(gran string, expr string) string {
	return fmt.Sprintf("sum by (%s) (%s)", gran, expr)
}

// (expr) > (value)
func greater(expr, value string) string {
	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(expr)
	sb.WriteByte(')')
	sb.WriteString(" > ")
	sb.WriteByte('(')
	sb.WriteString(value)
	sb.WriteByte(')')
	return sb.String()
}

// group by (@gran) (a)
func groupBy(gran string, expr string) string {
	return fmt.Sprintf("group by (%s) (%s)", gran, expr)
}

// a * on ( @on ) group_left (@gran) (group by (@on,@gran) (b))
//
//	a left join B on @on, append @gran labels into exprA result
func labelLeftOn(exprA string, on string, gran string, exprB string) string {
	return fmt.Sprintf("(%s) * on (%s) group_left (%s) group by (%s,%s) (%s)", exprA, on, gran, on, gran, exprB)
}

func _op(exprA, exprB, operator string) string {
	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(exprA)
	sb.WriteByte(')')
	sb.WriteString(operator)
	sb.WriteByte('(')
	sb.WriteString(exprB)
	sb.WriteByte(')')
	return sb.String()
}

// (a) / (b)
func div(exprA, exprB string) string { return _op(exprA, exprB, "/") }

// (a) * (b)
func mul(exprA, exprB string) string { return _op(exprA, exprB, "*") }

// (a) + (b)
func add(exprA, exprB string) string { return _op(exprA, exprB, "+") }

// (a) - (b)
func sub(exprA, exprB string) string { return _op(exprA, exprB, "-") }

// (a) or (b * 0 + def)
//
// e.g. increaseRate with default value: ((a / b)) or (a * 0 + 1)
func withDef(exprA string, series string, def string) string {
	return fmt.Sprintf("(%s) or ((%s) * 0 + %s)", exprA, series, def)
}

// (a + b) or (series1) or (series2) ...
func addWithDef(exprA, exprB string, series ...string) string {
	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(add(exprA, exprB))
	sb.WriteByte(')')
	for _, s := range series {
		sb.WriteString(" or ")
		sb.WriteByte('(')
		sb.WriteString(s)
		sb.WriteByte(')')
	}
	return sb.String()
}

// (a / b) or (series * 0 + def)
func divWithDef(exprA, exprB, series string, def string) string {
	return withDef(div(exprA, exprB), series, def)
}
