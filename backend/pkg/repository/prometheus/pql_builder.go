// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"log"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	prom_m "github.com/prometheus/common/model"
)

// this file is using @PQLFilter and @PQLTemplate to build more complex PQL statements

type QueryWithPQLFilter interface {
	QueryMetricsWithPQLFilter(ctx core.Context, pqlTpl PQLTemplate, startTime int64, endTime int64, gran Granularity, filter PQLFilter) ([]MetricResult, error)
	QueryRangeMetricsWithPQLFilter(ctx core.Context, pqlTpl PQLTemplate, startTime int64, endTime int64, stepMicroS int64, gran Granularity, filter PQLFilter) ([]MetricResult, error)

	// QuerySeriesWithPQLFilter(ctx core.Context, startTime int64, endTime int64, filter PQLFilter, metric ...Metric) ([]prom_m.LabelSet, error)
	FillMetric(ctx core.Context, res MetricGroupInterface, metricGroup MGroupName, startTime, endTime time.Time, filter PQLFilter, granularity Granularity) error
	FillRangeMetric(ctx core.Context, res MetricGroupInterface, metricGroup MGroupName, startTime, endTime time.Time, step time.Duration, filter PQLFilter, granularity Granularity) error

	GetInstanceListByPQLFilter(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) (*model.ServiceInstances, error)
	GetMultiSVCInstanceListByPQLFilter(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) (map[string]*model.ServiceInstances, error)

	// Query the db instance for specified service
	GetDescendantDatabase(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) ([]model.MiddlewareInstance, error)
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

func (repo *promRepo) QuerySeriesWithPQLFilter(ctx core.Context, startTime int64, endTime int64, filter PQLFilter, metric ...Metric) ([]prom_m.LabelSet, error) {
	vectors := []string{}
	for _, m := range metric {
		vectors = append(vectors, vector(m, filter).String())
	}
	labelSets, warnings, err := repo.GetApi().Series(ctx.GetContext(), vectors, time.UnixMicro(startTime), time.UnixMicro(endTime))
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}
	return labelSets, err
}

// ####################### PQL Template #######################

type Metric string

const (
	PROFILING_EPOLL_DURATION_SUM Metric = "kindling_profiling_epoll_duration_nanoseconds_sum"
	PROFILING_NET_DURATION_SUM   Metric = "kindling_profiling_net_duration_nanoseconds_sum"
	PROFILING_CPU_DURATION_SUM   Metric = "kindling_profiling_cpu_duration_nanoseconds_sum"

	SPAN_TRACE_COUNT        Metric = "kindling_span_trace_duration_nanoseconds_count"
	SPAN_TRACE_DURATION_SUM Metric = "kindling_span_trace_duration_nanoseconds_sum"

	SPAN_DB_COUNT        Metric = "kindling_db_duration_nanoseconds_count"
	SPAN_DB_DURATION_SUM Metric = "kindling_db_duration_nanoseconds_sum"

	LOG_LEVEL_COUNT     Metric = "originx_logparser_level_count_total"
	LOG_EXCEPTION_COUNT Metric = "originx_logparser_exception_count_total"

	MONITOR_STATUS Metric = "monitor_status"
)

func PQLMetricSeries(metric ...Metric) PQLTemplate {
	return func(rng, gran string, filter PQLFilter, _ string) string {
		// using lastOverTime to get full series exited during range
		var metrics []string
		for _, m := range metric {
			metrics = append(metrics, groupBy(gran, lastOverTime(rangeVec(m, filter, rng, ""))))
		}
		return or(metrics...)
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
	filterWithError := Clone(filter).Equal("is_error", "true")

	errorCount := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filterWithError, rng, offset)))
	requestCount := sumBy(gran, increase(rangeVec(SPAN_TRACE_COUNT, filter, rng, offset)))

	// ( errorCount or requestCount * 0 ) / requestCount
	// Used to retain a tag that exists in the requestCount but does not exist in the errorCount. Record that the request failure rate of this tag is 0
	return divWithDef(errorCount, requestCount, requestCount, "0")
}

// Average error rate of PQLAvgSQLErrorRateWithFilters query SQL requests
func PQLAvgSQLErrorRateWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	filterWithError := Clone(filter).Equal("is_error", "true")

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
	filterWithError := Clone(filter).RegexMatch("level", "error|critical")

	errorLevelCount := sumBy(gran, increase(rangeVec(LOG_LEVEL_COUNT, filterWithError, rng, offset)))
	exceptionCount := sumBy(gran, increase(rangeVec(LOG_EXCEPTION_COUNT, filter, rng, offset)))

	// ( errorLevelCount + exceptionCount ) or errorLevelCount or exceptionCount
	return addWithDef(errorLevelCount, exceptionCount, errorLevelCount, exceptionCount)
}

// WARNING: LogErrorCount without service will not return
func PQLAvgLogErrorCountCombineEndpointsInfoWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	grans := strings.Split(gran, ",")
	grans = append(grans, "container_id", "node_name", "pid")
	granWithCombineKey := strings.Join(grans, ",")

	filter, svcFilter := Clone(filter).SplitFilters([]string{ServiceNameKey, ContentKeyKey})
	errorLevelCount := sumBy(granWithCombineKey,
		increase(rangeVec(LOG_LEVEL_COUNT, Clone(filter).RegexMatch("level", "error|critical"), rng, offset)))

	exceptionCount := sumBy(granWithCombineKey,
		increase(rangeVec(LOG_EXCEPTION_COUNT, filter, rng, offset)))

	// ( errorLevelCount + exceptionCount ) or errorLevelCount or exceptionCount
	logErrorCount := addWithDef(errorLevelCount, exceptionCount, errorLevelCount, exceptionCount)

	k8sSVCGroup := groupBy("svc_name,content_key,container_id",
		lastOverTime(rangeVec(SPAN_TRACE_COUNT, Clone(svcFilter).NotEqual("container_id", ""), rng, offset)))

	vmSVCGroup := groupBy("svc_name,content_key,node_name,pid",
		lastOverTime(rangeVec(SPAN_TRACE_COUNT, Clone(svcFilter).Equal("container_id", "").NotEqual("pid", ""), rng, offset)))

	return addWithDef(
		sumBy(gran, labelLeftOn(logErrorCount, "pod", "svc_name,content_key", k8sSVCGroup)),
		sumBy(gran, labelLeftOn(logErrorCount, "node_name,pid", "svc_name,content_key", vmSVCGroup)),
		sumBy(gran, labelLeftOn(logErrorCount, "pod", "svc_name,content_key", k8sSVCGroup)),
		sumBy(gran, labelLeftOn(logErrorCount, "node_name,pid", "svc_name,content_key", vmSVCGroup)),
	)
}

func LogCountSeriesCombineSvcInfoWithPQLFilter(rng string, gran string, filter PQLFilter, offset string) string {
	grans := strings.Split(gran, ",")
	grans = append(grans, "container_id", "node_name", "pid")
	granWithCombineKey := strings.Join(grans, ",")

	filter, svcFilter := Clone(filter).SplitFilters([]string{ServiceNameKey})

	logCount := groupBy(granWithCombineKey, lastOverTime(rangeVec(LOG_LEVEL_COUNT, filter, rng, offset)))
	k8sSVCGroup := groupBy("svc_name,container_id",
		lastOverTime(rangeVec(SPAN_TRACE_COUNT, Clone(svcFilter).NotEqual("container_id", ""), rng, offset)))

	vmSVCGroup := groupBy("svc_name,node_name,pid",
		lastOverTime(rangeVec(SPAN_TRACE_COUNT, Clone(svcFilter).Equal("container_id", "").NotEqual("pid", ""), rng, offset)))

	return or(
		sumBy(gran, labelLeftOn(logCount, "container_id", "svc_name", k8sSVCGroup)),
		sumBy(gran, labelLeftOn(logCount, "node_name,pid", "svc_name", vmSVCGroup)),
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
	__name__   Metric
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
		return strings.Join(v.instantVec._strictPQL(string(v.__name__), v.modifier, v.offset, ""), " or ")
	}

	var sb strings.Builder
	sb.WriteString(string(v.__name__))
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
		return strings.Join(v.instantVec._strictPQL(string(v.__name__), v.modifier, v.offset, v.rangeVec), " or ")
	}

	var sb strings.Builder
	sb.WriteString(string(v.__name__))
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

func vector(metric Metric, filter PQLFilter) _vector {
	return _vector{
		__name__:   metric,
		instantVec: filter,
	}
}

func offsetVec(metric Metric, filter PQLFilter, offset string) _vector {
	return _vector{
		__name__:   metric,
		instantVec: filter,
		offset:     offset,
	}
}

func modifyVec(metric Metric, filter PQLFilter, modifier string) _vector {
	return _vector{
		__name__:   metric,
		instantVec: filter,
		modifier:   modifier,
	}
}

func rangeVec(metric Metric, filter PQLFilter, rangeVec string, offset string) _rangeVec {
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

// a * on ( @on ) group_left (@gran) (group by (@on,@gran) (b or a))
//
// a left join B on @on, append @gran labels into exprA result
func labelLeftOn(exprA string, on string, gran string, exprB string) string {
	return fmt.Sprintf("(%s) * on (%s) group_left (%s) group by (%s,%s) (%s)", exprA, on, gran, on, gran, exprB)
}

func labelReplace(expr string, sourceLabel, targetLabel string) string {
	return fmt.Sprintf(`label_replace(%s,"%s","%s", "%s","%s")`, expr, sourceLabel, "$1", targetLabel, "(.*)")
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

func and(exprA, exprB string) string { return _op(exprA, exprB, "and") }

func or(expr ...string) string {
	if len(expr) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(expr[0])
	sb.WriteByte(')')
	for i := 1; i < len(expr); i++ {
		sb.WriteString(" or ")
		sb.WriteByte('(')
		sb.WriteString(expr[i])
		sb.WriteByte(')')
	}
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
