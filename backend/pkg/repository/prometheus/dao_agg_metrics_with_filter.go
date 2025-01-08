// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

const DefaultDepLatency int64 = -1

// FillMetric query and populate RED metric
func (repo *promRepo) FillMetric(res MetricGroupInterface, metricGroup MGroupName, startTime, endTime time.Time, filters []string, granularity Granularity) {
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

	latency, err := repo.QueryAggMetricsWithFilter(
		decorator(PQLAvgLatencyWithFilters),
		startTS, endTS,
		granularity,
		filters...,
	)
	if err != nil {
		log.Println("query latency error: ", err)
	}
	res.MergeMetricResults(metricGroup, LATENCY, latency)

	errorRate, err := repo.QueryAggMetricsWithFilter(
		decorator(PQLAvgErrorRateWithFilters),
		startTS, endTS,
		granularity,
		filters...,
	)
	if err != nil {
		log.Println("query error rate error: ", err)
	}
	res.MergeMetricResults(metricGroup, ERROR_RATE, errorRate)

	if metricGroup == REALTIME {
		return
	}
	tps, err := repo.QueryAggMetricsWithFilter(
		decorator(PQLAvgTPSWithFilters),
		startTS, endTS,
		granularity,
		filters...,
	)
	if err != nil {
		log.Println("query tps error: ", err)
	}
	res.MergeMetricResults(metricGroup, THROUGHPUT, tps)
}

func (repo *promRepo) QueryAggMetricsWithFilter(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error) {
	if len(filterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of filterKVs is not even: %d", len(filterKVs))
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}
	vector := VecFromS2E(startTime, endTime)
	pql := pqlTemplate(vector, string(granularity), filters)
	return repo.QueryData(time.UnixMicro(endTime), pql)
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
