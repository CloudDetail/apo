// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"fmt"
	"math"
	"strconv"
	"time"

	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) UrlAVG(Urls *[]prom.EndpointMetrics, serviceName string, endTime time.Time, duration string) (*[]prom.EndpointMetrics, error) {
	var AvgErrorRateRes []prom.MetricResult
	//AvgErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvgError := prom.QueryEndPointPromql(duration, prom.AvgError, serviceName)
	AvgErrorRateRes, err := s.promRepo.QueryErrorRateData(endTime, queryAvgError)
	for _, result := range AvgErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.Avg.ErrorRate = &value
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
				newUrl.REDMetrics.Avg.ErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var AvgLatencyRes []prom.MetricResult
	//AvgLatencyRes, err = s.promRepo.QueryPrometheusLatency(searchTime)
	queryAvgLatency := prom.QueryEndPointPromql(duration, prom.AvgLatency, serviceName)
	AvgLatencyRes, err = s.promRepo.QueryLatencyData(endTime, queryAvgLatency)
	for _, result := range AvgLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.Avg.Latency = &value
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
				newUrl.REDMetrics.Avg.Latency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var AvgTPSRes []prom.MetricResult
	//AvgTPSRes, err = s.promRepo.QueryPrometheusTPS(searchTime)
	queryAvgTPS := prom.QueryEndPointPromql(duration, prom.AvgTPS, serviceName)
	AvgTPSRes, err = s.promRepo.QueryData(endTime, queryAvgTPS)
	for _, result := range AvgTPSRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.Avg.TPM = &value
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
				newUrl.REDMetrics.Avg.TPM = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}

// EndpointsREDMetric query Endpoint-level RED metric results (including average value, DoD/WoW Growth Rate)
func (s *service) EndpointsREDMetric(startTime, endTime time.Time, filters []string) *EndpointsMap {
	var res = &EndpointsMap{
		MetricGroupList: []*prom.EndpointMetrics{},
		MetricGroupMap:  map[prom.EndpointKey]*prom.EndpointMetrics{},
	}

	// Average RED metric over the fill time period
	s.promRepo.FillMetric(res, prom.AVG, startTime, endTime, filters, prom.EndpointGranularity)
	// RED metric day-to-day-on-da during the fill period
	s.promRepo.FillMetric(res, prom.DOD, startTime, endTime, filters, prom.EndpointGranularity)
	// RED metric week-on-week in the fill time period
	s.promRepo.FillMetric(res, prom.WOW, startTime, endTime, filters, prom.EndpointGranularity)

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

func (s *service) EndpointsRealtimeREDMetric(filter EndpointsFilter, endpointsMap *EndpointsMap, startTime time.Time, endTime time.Time) {
	filters := filter.ExtractFilterStr()
	s.promRepo.FillMetric(endpointsMap, prom.REALTIME, startTime, endTime, filters, prom.EndpointGranularity)
}

// EndpointsDelaySource fill delay source
// Based on the input Endpoints, records that do not exist in the Endpoints are discarded.
func (s *service) EndpointsDelaySource(endpoints *EndpointsMap, startTime, endTime time.Time, filters []string) error {

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	metricResults, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.WithDefaultIFPolarisMetricExits(prom.PQLDepLatencyRadioWithFilters, prom.DefaultDepLatency),
		startTS, endTS,
		prom.EndpointGranularity,
		filters...,
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

func (s *service) EndpointsNamespaceInfo(endpoints *EndpointsMap, startTime, endTime time.Time, filters []string) error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	metricResult, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLAvgTPSWithFilters,
		startTS, endTS,
		prom.NSEndpointGranularity,
		filters...,
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

func (s *service) UrlDOD(Urls *[]prom.EndpointMetrics, serviceName string, endTime time.Time, duration string) (*[]prom.EndpointMetrics, error) {
	latencyDODquery := prom.QueryEndPointPromql(duration, prom.LatencyDOD, serviceName)
	latencyDoDres, err := s.promRepo.QueryData(endTime, latencyDODquery)
	for _, result := range latencyDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.DOD.Latency = &value
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
				newUrl.REDMetrics.DOD.Latency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}

	errorDODquery := prom.QueryEndPointPromql(duration, prom.ErrorDOD, serviceName)
	errorDoDres, err := s.promRepo.QueryData(endTime, errorDODquery)
	// Update the contents of the wrongUrls
	for _, result := range errorDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // if value is infinity, the error rate is assigned to MaxFloat64
					(*Urls)[i].REDMetrics.DOD.ErrorRate = &value
				} else {
					var value float64
					value = RES_MAX_VALUE
					pointer := &value
					(*Urls)[i].REDMetrics.DOD.ErrorRate = pointer
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
				newUrl.REDMetrics.DOD.ErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	//tpsResults, err := s.promRepo.QueryPrometheusTPSDayOver(searchTime)
	tpsDODquery := prom.QueryEndPointPromql(duration, prom.TPSDOD, serviceName)
	tpsResults, err := s.promRepo.QueryData(endTime, tpsDODquery)
	for _, result := range tpsResults {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.DOD.TPM = &value
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
				newUrl.REDMetrics.DOD.TPM = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}
func (s *service) UrlWOW(Urls *[]prom.EndpointMetrics, serviceName string, endTime time.Time, duration string) (*[]prom.EndpointMetrics, error) {

	var LatencyWoWRes []prom.MetricResult
	//LatencyWoWRes, err = s.promRepo.QueryPrometheusLatencyWeekOver(searchTime)
	latencyWOWquery := prom.QueryEndPointPromql(duration, prom.LatencyWOW, serviceName)
	LatencyWoWRes, err := s.promRepo.QueryData(endTime, latencyWOWquery)
	for _, result := range LatencyWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.WOW.Latency = &value
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
				newUrl.REDMetrics.WOW.Latency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var TPSWoWRes []prom.MetricResult
	//TPSWoWRes, err = s.promRepo.QueryPrometheusTPSWeekOver(searchTime)
	TPSWOWquery := prom.QueryEndPointPromql(duration, prom.TPSWOW, serviceName)
	TPSWoWRes, err = s.promRepo.QueryData(endTime, TPSWOWquery)
	for _, result := range TPSWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					(*Urls)[i].REDMetrics.WOW.TPM = &value
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
				newUrl.REDMetrics.WOW.TPM = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var ErrorWoWRes []prom.MetricResult
	//ErrorWoWRes, err = s.promRepo.QueryPrometheusErrorWeekOver(searchTime)
	errorWoWquery := prom.QueryEndPointPromql(duration, prom.ErrorWOW, serviceName)
	ErrorWoWRes, err = s.promRepo.QueryData(endTime, errorWoWquery)
	for _, result := range ErrorWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { // if value is infinity, the error rate is assigned to MaxFloat64
					(*Urls)[i].REDMetrics.WOW.ErrorRate = &value
				} else {
					var value float64
					value = RES_MAX_VALUE
					pointer := &value
					(*Urls)[i].REDMetrics.WOW.ErrorRate = pointer
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
				newUrl.REDMetrics.WOW.ErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}

// EndpointRangeREDChart query graph
func (s *service) EndpointRangeREDChart(Services *[]ServiceDetail, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]ServiceDetail, error) {
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
		errorDataRes, err = s.promRepo.QueryRangeErrorData(startTime, endTime, errorDataQuery, step)
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
		LatencyDataRes, err = s.promRepo.QueryRangeLatencyData(startTime, endTime, latencyDataQuery, step)
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
		TPSLastDataRes, err = s.promRepo.QueryRangeData(startTime, endTime, TPSDataQuery, step)
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
func (s *service) UrlLatencySource(Urls *[]prom.EndpointMetrics, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]prom.EndpointMetrics, error) {
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
	LatencySourceRes, err := s.promRepo.QueryRangeData(startTime, endTime, LatencySourcequery, step)
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
func (s *service) UrlAVG1min(Urls *[]prom.EndpointMetrics, serviceName string, endTime time.Time, duration string) (*[]prom.EndpointMetrics, error) {
	var Avg1minErrorRateRes []prom.MetricResult
	//Avg1minErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvg1minError := prom.QueryEndPointPromql(duration, prom.Avg1minError, serviceName)
	Avg1minErrorRateRes, err := s.promRepo.QueryErrorRateData(endTime, queryAvg1minError)
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
	Avg1minLatencyRes, err = s.promRepo.QueryLatencyData(endTime, queryAvg1minLatency)
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
