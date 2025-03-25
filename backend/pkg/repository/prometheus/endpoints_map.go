// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"errors"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// EndpointsMap is used to store the query results of multiple metrics of the same granularity, using MergeMetricResults merge.
type EndpointsMap = MetricGroupMap[EndpointKey, *EndpointMetrics]

type FetchEMOption func(
	promRepo Repo,
	em *EndpointsMap,
	startTime, endTime time.Time,
	filters []string,
) error

func FetchEndpointsData(
	promRepo Repo,
	filters []string,
	startTime, endTime time.Time,
	opts ...FetchEMOption) (*EndpointsMap, error) {
	result := &EndpointsMap{
		MetricGroupList: []*EndpointMetrics{},
		MetricGroupMap:  map[EndpointKey]*EndpointMetrics{},
	}

	var errs []error
	for _, fetchFunc := range opts {
		err := fetchFunc(promRepo, result, startTime, endTime, filters)
		errs = append(errs, err)
	}

	return result, errors.Join(errs...)
}

func WithREDMetric() FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		var errs []error
		// Average RED metric over the fill time period
		if err := promRepo.FillMetric(em, AVG, startTime, endTime, filters, EndpointGranularity); err != nil {
			errs = append(errs, err)
		}
		// RED metric day-to-day-on-da during the fill period
		if err := promRepo.FillMetric(em, DOD, startTime, endTime, filters, EndpointGranularity); err != nil {
			errs = append(errs, err)
		}
		// RED metric week-on-week in the fill time period
		if err := promRepo.FillMetric(em, WOW, startTime, endTime, filters, EndpointGranularity); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}
}

func WithDelaySource() FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		metricResults, err := promRepo.QueryAggMetricsWithFilter(
			WithDefaultIFPolarisMetricExits(PQLDepLatencyRadioWithFilters, DefaultDepLatency),
			startTime.UnixMicro(), endTime.UnixMicro(),
			EndpointGranularity,
			filters...,
		)

		if err != nil {
			return err
		}

		for _, metricResult := range metricResults {
			key := EndpointKey{
				SvcName:    metricResult.Metric.SvcName,
				ContentKey: metricResult.Metric.ContentKey,
			}
			// All consolidated values contain only the results at the latest point in time, directly take the metricResult.values[0]
			value := metricResult.Values[0].Value
			if endpoint, ok := em.MetricGroupMap[key]; ok {
				endpoint.DelaySource = &value
			}
		}

		// Because the default initial value of the float64 is 0, which means that the external dependency delay ratio is 0
		// as expected, so Endpoint that are not queried to DepLatencyRadio are no longer initialized
		return nil
	}
}

func WithNamespace() FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		metricResult, err := promRepo.QueryAggMetricsWithFilter(
			PQLAvgTPSWithFilters,
			startTime.UnixMicro(), endTime.UnixMicro(),
			NSEndpointGranularity,
			filters...,
		)
		if err != nil {
			return err
		}

		for _, metric := range metricResult {
			if len(metric.Values) <= 0 {
				continue
			}
			key := EndpointKey{
				SvcName:    metric.Metric.SvcName,
				ContentKey: metric.Metric.ContentKey,
			}
			if endpoint, ok := em.MetricGroupMap[key]; ok {
				if len(metric.Metric.Namespace) > 0 {
					// Because the query granularity is namespace and svc_name, the contentKey does not need to be deduplication.
					endpoint.NamespaceList = appendIfNotExist(endpoint.NamespaceList, metric.Metric.Namespace)
				}
			}
		}
		return nil
	}
}

func appendIfNotExist(slice []string, str string) []string {
	for _, item := range slice {
		if item == str {
			return slice
		}
	}
	return append(slice, str)
}

func WithRealTimeREDMetric() FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		return promRepo.FillMetric(em, REALTIME, startTime, endTime, filters, EndpointGranularity)
	}
}

func WithREDChart(step time.Duration) FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		return promRepo.FillRangeMetric(em, AVG, startTime, endTime, step, filters, EndpointGranularity)
	}
}

func WithLogErrorCount() FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		result, err := promRepo.QueryAggMetricsWithFilter(PQLAvgLogErrorCountCombineEndpointsInfoWithFilters,
			startTime.UnixMicro(), endTime.UnixMicro(), EndpointGranularity, filters...,
		)
		if err != nil {
			return err
		}
		em.MergeMetricResults(AVG, LOG_ERROR_COUNT, result)
		return nil
	}
}

func ReverseSortWithMetrics(sortType request.SortType) func(i, j *EndpointMetrics) int {
	return func(i, j *EndpointMetrics) int {
		var itemI, itemJ *float64
		switch sortType {
		case request.SortByLatency:
			itemI = i.REDMetrics.Avg.Latency
			itemJ = j.REDMetrics.Avg.Latency
		case request.SortByErrorRate:
			itemI = i.REDMetrics.Avg.ErrorRate
			itemJ = j.REDMetrics.Avg.ErrorRate
		case request.SortByThroughput:
			itemI = i.REDMetrics.Avg.TPM
			itemJ = j.REDMetrics.Avg.TPM
		case request.SortByLogErrorCount:
			itemI = &i.AvgLogErrorCount
			itemJ = &j.AvgLogErrorCount
		}

		if itemI == nil && itemJ == nil {
			return 0
		}

		switch {
		case itemI == nil:
			return 1
		case itemJ == nil:
			return -1
		}

		switch {
		case *itemI > *itemJ:
			return -1
		case *itemI < *itemJ:
			return 1
		default:
			return 0
		}
	}
}
