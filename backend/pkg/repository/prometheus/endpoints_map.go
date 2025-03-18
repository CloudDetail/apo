// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"errors"
	"time"
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
	filter EndpointsFilter,
	startTime, endTime time.Time,
	opts ...FetchEMOption) *EndpointsMap {
	result := &EndpointsMap{
		MetricGroupList: []*EndpointMetrics{},
		MetricGroupMap:  map[EndpointKey]*EndpointMetrics{},
	}

	filterStrs := filter.ExtractFilterStr()
	for _, fetchFunc := range opts {
		fetchFunc(promRepo, result, startTime, endTime, filterStrs)
	}

	return result
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
					endpoint.NamespaceList = append(endpoint.NamespaceList, metric.Metric.Namespace)
				}
			}
		}
		return nil
	}
}

func WithRealTimeREDMetric() FetchEMOption {
	return func(promRepo Repo, em *EndpointsMap, startTime, endTime time.Time, filters []string) error {
		return promRepo.FillMetric(em, REALTIME, startTime, endTime, filters, EndpointGranularity)
	}
}

func WithREDChart(step time.Duration) FetchEMOption {

}

type EndpointsFilter struct {
	ContainsSvcName      string   // SvcName, containing matches
	ContainsEndpointName string   // EndpointName, containing matches
	Namespace            string   // Namespace, exact match
	ServiceName          string   // Specify the service name, exact match
	MultiService         []string // multiple service names, exact match
	MultiNamespace       []string // multiple namespace, exact match
	MultiEndpoint        []string // multiple service endpoints, exact match
}

// EndpointsFilter extraction filter conditions
// Returns a string array with an even length, with the odd bits being key and the even bits being value
func (f EndpointsFilter) ExtractFilterStr() []string {
	var filters []string
	if len(f.ServiceName) > 0 {
		filters = append(filters, ServicePQLFilter, f.ServiceName)
	} else if len(f.ContainsSvcName) > 0 {
		filters = append(filters, ServiceRegexPQLFilter, RegexContainsValue(f.ContainsSvcName))
	}
	if len(f.ContainsEndpointName) > 0 {
		filters = append(filters, ContentKeyRegexPQLFilter, RegexContainsValue(f.ContainsEndpointName))
	}
	if len(f.Namespace) > 0 {
		filters = append(filters, NamespacePQLFilter, f.Namespace)
	}
	if len(f.MultiNamespace) > 0 {
		filters = append(filters, NamespaceRegexPQLFilter, RegexMultipleValue(f.MultiNamespace...))
	}
	if len(f.MultiService) > 0 {
		filters = append(filters, ServiceRegexPQLFilter, RegexMultipleValue(f.MultiService...))
	}
	if len(f.MultiEndpoint) > 0 {
		filters = append(filters, ContentKeyRegexPQLFilter, RegexMultipleValue(f.MultiEndpoint...))
	}
	return filters
}
