// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"fmt"
	"math"
	"strconv"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// EndpointsREDMetric query Endpoint-level RED metric results (including average value, DoD/WoW Growth Rate)
func (s *service) EndpointsREDMetric(ctx core.Context, startTime, endTime time.Time, filter prom.PQLFilter) *EndpointsMap {
	var res = &EndpointsMap{
		MetricGroupList: []*prom.EndpointMetrics{},
		MetricGroupMap:  map[prom.EndpointKey]*prom.EndpointMetrics{},
	}

	// Average RED metric over the fill time period
	s.promRepo.FillMetric(ctx, res, prom.AVG, startTime, endTime, filter, prom.EndpointGranularity)
	// RED metric day-to-day-on-da during the fill period
	s.promRepo.FillMetric(ctx, res, prom.DOD, startTime, endTime, filter, prom.EndpointGranularity)
	// RED metric week-on-week in the fill time period
	s.promRepo.FillMetric(ctx, res, prom.WOW, startTime, endTime, filter, prom.EndpointGranularity)

	return res
}

// EndpointsFilter extraction filter conditions
// Returns a string array with an even length, with the odd bits being key and the even bits being value
func (f EndpointsFilter) ExtractFilterStr() []string {
	var filters []string
	if len(f.ServiceName) > 0 {
		filters = append(filters, prom.ServicePQLFilter, f.ServiceName)
	} else if len(f.ContainsSvcName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(f.ContainsSvcName))
	}
	if len(f.ContainsEndpointName) > 0 {
		filters = append(filters, prom.ContentKeyRegexPQLFilter, prom.RegexContainsValue(f.ContainsEndpointName))
	}
	if len(f.Namespace) > 0 {
		filters = append(filters, prom.NamespacePQLFilter, f.Namespace)
	}
	if len(f.MultiNamespace) > 0 {
		filters = append(filters, prom.NamespaceRegexPQLFilter, prom.RegexMultipleValue(f.MultiNamespace...))
	}
	if len(f.MultiService) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(f.MultiService...))
	}
	if len(f.MultiEndpoint) > 0 {
		filters = append(filters, prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(f.MultiEndpoint...))
	}
	return filters
}

func (f EndpointsFilter) ExtractPQLFilterStr() prom.PQLFilter {
	filter := prom.NewFilter()

	if len(f.ServiceName) > 0 {
		filter.AddPatternFilter(prom.ServicePQLFilter, f.ServiceName)
	} else if len(f.ContainsSvcName) > 0 {
		filter.AddPatternFilter(prom.ServiceRegexPQLFilter, prom.RegexContainsValue(f.ContainsSvcName))
	}
	if len(f.ContainsEndpointName) > 0 {
		filter.AddPatternFilter(prom.ContentKeyRegexPQLFilter, prom.RegexContainsValue(f.ContainsEndpointName))
	}
	if len(f.Namespace) > 0 {
		filter.AddPatternFilter(prom.NamespacePQLFilter, f.Namespace)
	}
	if len(f.MultiNamespace) > 0 {
		filter.AddPatternFilter(prom.NamespaceRegexPQLFilter, prom.RegexMultipleValue(f.MultiNamespace...))
	}
	if len(f.MultiService) > 0 {
		filter.AddPatternFilter(prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(f.MultiService...))
	}
	if len(f.MultiEndpoint) > 0 {
		filter.AddPatternFilter(prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(f.MultiEndpoint...))
	}
	if len(f.ClusterIDs) > 0 {
		filter.RegexMatch(prom.ClusterIDKey, prom.RegexMultipleValue(f.ClusterIDs...))
	}
	return filter
}

func (s *service) EndpointsRealtimeREDMetric(ctx core.Context, filter prom.PQLFilter, endpointsMap *EndpointsMap, startTime time.Time, endTime time.Time) {
	s.promRepo.FillMetric(ctx, endpointsMap, prom.REALTIME, startTime, endTime, filter, prom.EndpointGranularity)
}

// EndpointsDelaySource fill delay source
// Based on the input Endpoints, records that do not exist in the Endpoints are discarded.
func (s *service) EndpointsDelaySource(ctx core.Context, endpoints *EndpointsMap, startTime, endTime time.Time, filter prom.PQLFilter) error {

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	metricResults, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.WithDefaultForPolarisActiveSeries(prom.PQLDepLatencyRadioWithPQLFilter, prom.DefaultDepLatency),
		startTS, endTS,
		prom.EndpointGranularity,
		filter,
	)
	if err != nil {
		return err
	}

	for _, metricResult := range metricResults {
		key := prom.EndpointKey{
			SvcName:    metricResult.Metric.SvcName,
			ContentKey: metricResult.Metric.ContentKey,
		}
		// All consolidated values contain only the results at the latest point in time, directly take the metricResult.values[0]
		value := metricResult.Values[0].Value
		if endpoint, ok := endpoints.MetricGroupMap[key]; ok {
			endpoint.DelaySource = &value
		}
	}

	// Because the default initial value of the float64 is 0, which means that the external dependency delay ratio is 0
	// as expected, so Endpoint that are not queried to DepLatencyRadio are no longer initialized
	return nil
}

func (s *service) EndpointsNamespaceInfo(ctx core.Context, endpoints *EndpointsMap, startTime, endTime time.Time, filter prom.PQLFilter) error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	metricResult, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.PQLAvgTPSWithPQLFilter,
		startTS, endTS,
		prom.NSEndpointGranularity,
		filter,
	)
	if err != nil {
		return err
	}

	for _, metric := range metricResult {
		if len(metric.Values) <= 0 {
			continue
		}
		key := prom.EndpointKey{
			SvcName:    metric.Metric.SvcName,
			ContentKey: metric.Metric.ContentKey,
		}
		if endpoint, ok := endpoints.MetricGroupMap[key]; ok {
			if len(metric.Metric.Namespace) > 0 {
				// Because the query granularity is namespace and svc_name, the contentKey does not need to be deduplication.
				endpoint.NamespaceList = append(endpoint.NamespaceList, metric.Metric.Namespace)
			}
		}
	}

	return nil
}

// EndpointRangeREDChart query graph
func (s *service) EndpointRangeREDChart(ctx core.Context, Services *[]ServiceDetail, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]ServiceDetail, error) {
	var newUrls []prom.EndpointMetrics
	var contentKeys []string
	var stepToStr string

	stepMinutes := float64(step) / float64(time.Minute)
	// Format as a string, keeping one decimal place
	stepToStr = fmt.Sprintf("%.1fm", stepMinutes)

	// Traverse the services array, get the ContentKey of each URL and store it in the slice.
	for _, service := range *Services {
		for _, Url := range service.Endpoints {
			contentKeys = append(contentKeys, Url.ContentKey)
		}
	}

	var err error
	var errorDataRes []prom.MetricResult
	// query every 300 urls
	batchSize := 300
	// Batch contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		errorDataQuery := prom.QueryEndPointRangePromql(stepToStr, duration, prom.ErrorData, batch)
		errorDataRes, err = s.promRepo.QueryRangeErrorData(ctx, startTime, endTime, errorDataQuery, step)
		for _, result := range errorDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false

			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].ErrorRateData = result.Values
					break
				}
			}
			if !found {
				newUrl := prom.EndpointMetrics{
					EndpointKey: prom.EndpointKey{
						ContentKey: contentKey,
						SvcName:    serviceName,
					},
					ErrorRateData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}
	}
	var LatencyDataRes []prom.MetricResult
	// Batch contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//LatencyDataRes, err = s.promRepo.QueryRangePrometheusLatencyLast30min(searchTime)
		latencyDataQuery := prom.QueryEndPointRangePromql(stepToStr, duration, prom.LatencyData, batch)
		LatencyDataRes, err = s.promRepo.QueryRangeLatencyData(ctx, startTime, endTime, latencyDataQuery, step)
		for _, result := range LatencyDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false
			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].LatencyData = result.Values
					break
				}
			}
			if !found {
				newUrl := prom.EndpointMetrics{
					EndpointKey: prom.EndpointKey{
						ContentKey: contentKey,
						SvcName:    serviceName,
					},
					LatencyData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}
	}
	var TPSLastDataRes []prom.MetricResult
	// Batch contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//TPSLastDataRes, err = s.promRepo.QueryRangePrometheusTPSLast30min(searchTime)
		TPSDataQuery := prom.QueryEndPointRangePromql(stepToStr, duration, prom.TPSData, batch)
		TPSLastDataRes, err = s.promRepo.QueryRangeData(ctx, startTime, endTime, TPSDataQuery, step)
		for _, result := range TPSLastDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false
			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].TPMData = result.Values
					break
				}
			}
			if !found {
				newUrl := prom.EndpointMetrics{
					EndpointKey: prom.EndpointKey{
						ContentKey: contentKey,
						SvcName:    serviceName,
					},
					TPMData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}

	}

	for _, url := range newUrls {
		serviceName := url.SvcName
		contentKey := url.ContentKey
		for j, _ := range *Services {
			if (*Services)[j].ServiceName == serviceName {
				for k, _ := range (*Services)[j].Endpoints {
					if contentKey == (*Services)[j].Endpoints[k].ContentKey {
						(*Services)[j].Endpoints[k].LatencyData = url.LatencyData
						(*Services)[j].Endpoints[k].ErrorRateData = url.ErrorRateData
						(*Services)[j].Endpoints[k].TPMData = url.TPMData
					}
				}
			}
		}
	}
	return Services, err
}

// UrlLatencySource query latency depends on
func (s *service) UrlLatencySource(ctx core.Context, Urls *[]prom.EndpointMetrics, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]prom.EndpointMetrics, error) {
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	var LatencySourceRes []prom.MetricResult
	//LatencySourceRes, err = s.promRepo.QueryPrometheusLatencyWeekOver(searchTime)
	LatencySourcequery := prom.QueryEndPointPromql(stepToStr, prom.DelaySource, serviceName)
	LatencySourceRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LatencySourcequery, step)
	for _, result := range LatencySourceRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].DelaySource = &value
				}
				break
			}
		}

	}

	return Urls, err
}

// UrlAVG1min the average value in the last minute
func (s *service) UrlAVG1min(ctx core.Context, Urls *[]prom.EndpointMetrics, serviceName string, endTime time.Time, duration string) (*[]prom.EndpointMetrics, error) {
	var Avg1minErrorRateRes []prom.MetricResult
	//Avg1minErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvg1minError := prom.QueryEndPointPromql(duration, prom.Avg1minError, serviceName)
	Avg1minErrorRateRes, err := s.promRepo.QueryErrorRateData(ctx, endTime, queryAvg1minError)
	//log.Printf("%v", Avg1minErrorRateRes)
	for _, result := range Avg1minErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.Realtime.ErrorRate = &value
				}
				break
			}
		}
		if !found {
			newUrl := prom.EndpointMetrics{
				EndpointKey: prom.EndpointKey{
					ContentKey: contentKey,
					SvcName:    serviceName,
				},
			}
			if !math.IsInf(value, 0) { // does not assign value when it is infinity
				newUrl.REDMetrics.Realtime.ErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var Avg1minLatencyRes []prom.MetricResult
	//Avg1minLatencyRes, err = s.promRepo.QueryPrometheusLatency(searchTime)
	queryAvg1minLatency := prom.QueryEndPointPromql(duration, prom.Avg1minLatency, serviceName)
	Avg1minLatencyRes, err = s.promRepo.QueryLatencyData(ctx, endTime, queryAvg1minLatency)
	for _, result := range Avg1minLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.Realtime.Latency = &value
				}
				break
			}
		}
		if !found {
			newUrl := prom.EndpointMetrics{
				EndpointKey: prom.EndpointKey{
					ContentKey: contentKey,
					SvcName:    serviceName,
				},
			}
			if !math.IsInf(value, 0) { // does not assign value when it is infinity
				newUrl.REDMetrics.Realtime.Latency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}

// EndpointsMap is used to store the query results of multiple metrics of the same granularity, using MergeMetricResults merge.
type EndpointsMap = prom.MetricGroupMap[prom.EndpointKey, *prom.EndpointMetrics]
