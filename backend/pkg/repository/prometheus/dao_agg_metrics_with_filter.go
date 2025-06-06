// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

const DefaultDepLatency int64 = -1

func (repo *promRepo) FillRangeMetric(ctx core.Context, res MetricGroupInterface, metricGroup MGroupName, startTime, endTime time.Time, step time.Duration, filters []string, granularity Granularity) error {
	var decorator = func(apf AggPQLWithFilters) AggPQLWithFilters {
		return apf
	}

	switch metricGroup {
	case REALTIME:
		startTime = endTime.Add(-3 * time.Minute)
	case DOD:
		decorator = DayOnDay
	case WOW:
		decorator = WeekOnWeek
	}

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()
	stepMicro := step.Microseconds()

	var errs []error
	latency, err := repo.QueryRangeAggMetricsWithFilter(ctx,
		decorator(PQLAvgLatencyWithFilters),
		startTS, endTS, stepMicro,
		granularity,
		filters...,
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		res.MergeRangeMetricResults(metricGroup, LATENCY, latency)
	}

	errorRate, err := repo.QueryRangeAggMetricsWithFilter(ctx,
		decorator(PQLAvgErrorRateWithFilters),
		startTS, endTS, stepMicro,
		granularity,
		filters...,
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		res.MergeRangeMetricResults(metricGroup, ERROR_RATE, errorRate)
	}

	if metricGroup == REALTIME {
		return errors.Join(err)
	}
	tps, err := repo.QueryRangeAggMetricsWithFilter(ctx,
		decorator(PQLAvgTPSWithFilters),
		startTS, endTS, stepMicro,
		granularity,
		filters...,
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		res.MergeRangeMetricResults(metricGroup, THROUGHPUT, tps)
	}

	return errors.Join(errs...)
}

// FillMetric query and populate RED metric
func (repo *promRepo) FillMetric(ctx core.Context, res MetricGroupInterface, metricGroup MGroupName, startTime, endTime time.Time, filters []string, granularity Granularity) error {
	var decorator = func(apf AggPQLWithFilters) AggPQLWithFilters {
		return apf
	}

	switch metricGroup {
	case REALTIME:
		startTime = endTime.Add(-3 * time.Minute)
	case DOD:
		decorator = DayOnDay
	case WOW:
		decorator = WeekOnWeek
	}

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	var errs []error
	latency, err := repo.QueryAggMetricsWithFilter(ctx,
		decorator(PQLAvgLatencyWithFilters),
		startTS, endTS,
		granularity,
		filters...,
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		res.MergeMetricResults(metricGroup, LATENCY, latency)
	}

	errorRate, err := repo.QueryAggMetricsWithFilter(ctx,
		decorator(PQLAvgErrorRateWithFilters),
		startTS, endTS,
		granularity,
		filters...,
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		res.MergeMetricResults(metricGroup, ERROR_RATE, errorRate)
	}

	if metricGroup == REALTIME {
		return errors.Join(err)
	}
	tps, err := repo.QueryAggMetricsWithFilter(ctx,
		decorator(PQLAvgTPSWithFilters),
		startTS, endTS,
		granularity,
		filters...,
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		res.MergeMetricResults(metricGroup, THROUGHPUT, tps)
	}

	return errors.Join(errs...)
}

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
