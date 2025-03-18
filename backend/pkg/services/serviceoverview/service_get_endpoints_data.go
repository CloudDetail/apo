// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"math"
	"slices"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServicesEndPointData(
	startTime, endTime time.Time, step time.Duration,
	filter EndpointsFilter,
	sortRule SortType,
) ([]response.ServiceEndPointsRes, error) {
	filterStrs := filter.ExtractFilterStr()

	var opts = []prometheus.FetchEMOption{
		prometheus.WithREDMetric(),
		prometheus.WithDelaySource(),
		prometheus.WithNamespace(),
	}

	if sortRule == SortByLogErrorCount {
		opts = append(opts, prometheus.WithLogErrorCount())
	} else if sortRule == MUTATIONSORT {
		opts = append(opts, prometheus.WithRealTimeREDMetric())
	}

	endpointsMap, err := prometheus.FetchEndpointsData(
		s.promRepo, filterStrs, startTime, endTime,
		opts...,
	)

	s.sortWithRule(sortRule, endpointsMap)

	services := groupEndpointsByService(endpointsMap.MetricGroupList, 3)
	var servicesResMsg []response.ServiceEndPointsRes
	for _, service := range services {
		if service.ServiceName == "" {
			continue
		}
		serviceDetails := s.extractDetail(service, startTime, endTime, step)

		if serviceDetails == nil {
			continue
		}

		// endpoint namespaceList to remove weight
		tmpSet := make(map[string]struct{})
		nsList := make([]string, 0)
		for _, endpoint := range service.Endpoints {
			for _, ns := range endpoint.NamespaceList {
				if _, find := tmpSet[ns]; find {
					continue
				}
				tmpSet[ns] = struct{}{}
				nsList = append(nsList, ns)
			}
		}

		newServiceRes := response.ServiceEndPointsRes{
			ServiceName:    service.ServiceName,
			Namespaces:     nsList,
			EndpointCount:  service.EndpointCount,
			ServiceDetails: serviceDetails,
		}

		servicesResMsg = append(servicesResMsg, newServiceRes)
	}
	return servicesResMsg, err
}

func (s *service) sortWithRule(sortRule SortType, endpointsMap *EndpointsMap) error {
	switch sortRule {
	case SortByLatency, SortByErrorRate, SortByThroughput, SortByLogErrorCount:
		slices.SortStableFunc(endpointsMap.MetricGroupList, sortWithMetrics(sortRule))
	case DODThreshold: //Sort by Day-to-Year Threshold
		threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
		if err != nil {
			return err
		}
		errorThreshold := threshold.ErrorRate
		// No throughput comparison
		//tpsThreshold := threshold.Tps
		latencyThreshold := threshold.Latency
		for i, _ := range endpointsMap.MetricGroupList {
			endpoint := endpointsMap.MetricGroupList[i]

			// The filling error rate is not equal to 0, and the year-on-year comparison cannot be found when there is a request. The filling is the maximum value (filling is performed by determining whether there is a request and when there is a request)
			if endpoint.REDMetrics.DOD.Latency != nil && endpoint.REDMetrics.DOD.ErrorRate == nil && endpoint.REDMetrics.Avg.ErrorRate != nil && *endpoint.REDMetrics.Avg.ErrorRate != 0 {
				endpoint.REDMetrics.DOD.ErrorRate = new(float64)
				*endpoint.REDMetrics.DOD.ErrorRate = RES_MAX_VALUE
			}
			if endpoint.REDMetrics.WOW.Latency != nil && endpoint.REDMetrics.WOW.ErrorRate == nil && endpoint.REDMetrics.Avg.ErrorRate != nil && *endpoint.REDMetrics.Avg.ErrorRate != 0 {
				endpoint.REDMetrics.WOW.ErrorRate = new(float64)
				*endpoint.REDMetrics.WOW.ErrorRate = RES_MAX_VALUE
			}
			// Filter error rate
			if endpoint.REDMetrics.DOD.ErrorRate != nil && *endpoint.REDMetrics.DOD.ErrorRate > errorThreshold {
				endpoint.IsErrorRateExceeded = true
				endpoint.AlertCount += ErrorCount
			}
			// Filter delay
			if endpoint.REDMetrics.DOD.Latency != nil && *endpoint.REDMetrics.DOD.Latency > latencyThreshold {
				endpoint.IsLatencyExceeded = true
				endpoint.AlertCount += LatencyCount
			}
		}
		sortByDODThreshold(endpointsMap.MetricGroupList)
	case MUTATIONSORT: //Sort by real-time mutation rate
		sortByMutation(endpointsMap.MetricGroupList)
	}

	return nil
}

func sortWithMetrics(sortType SortType) func(i, j *prometheus.EndpointMetrics) int {
	return func(i, j *prometheus.EndpointMetrics) int {
		var itemI, itemJ *float64
		switch sortType {
		case SortByLatency:
			itemI = i.REDMetrics.Avg.Latency
			itemJ = j.REDMetrics.Avg.Latency
		case SortByErrorRate:
			itemI = i.REDMetrics.Avg.ErrorRate
			itemJ = j.REDMetrics.Avg.ErrorRate
		case SortByThroughput:
			itemI = i.REDMetrics.Avg.TPM
			itemJ = j.REDMetrics.Avg.TPM
		case SortByLogErrorCount:
			itemI = &i.AvgLogErrorCount
			itemJ = &j.AvgLogErrorCount
		}

		if itemI == nil && itemJ == nil {
			return 0
		}

		switch {
		case itemI == nil:
			return -1
		case itemJ == nil:
			return 1
		}

		switch {
		case *itemI > *itemJ:
			return 1
		case *itemI < *itemJ:
			return -1
		default:
			return 0
		}
	}
}

func (*service) extractDetail(
	service *ServiceDetail,
	startTime, endTime time.Time, step time.Duration,
) []response.ServiceDetail {
	var newServiceDetails []response.ServiceDetail
	for _, endpoint := range service.Endpoints {
		newErrorRadio := response.Ratio{
			DayOverDay:  endpoint.REDMetrics.DOD.ErrorRate,
			WeekOverDay: endpoint.REDMetrics.WOW.ErrorRate,
		}
		newErrorRate := response.TempChartObject{
			Ratio: newErrorRadio,
		}
		if endpoint.REDMetrics.Avg.ErrorRate != nil && !math.IsInf(*endpoint.REDMetrics.Avg.ErrorRate, 0) { // does not assign a value when it is infinite
			newErrorRate.Value = endpoint.REDMetrics.Avg.ErrorRate
		}

		newtpsRadio := response.Ratio{
			DayOverDay:  endpoint.REDMetrics.DOD.TPM,
			WeekOverDay: endpoint.REDMetrics.WOW.TPM,
		}
		newtpsRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newtpsRadio,
		}
		if endpoint.REDMetrics.Avg.TPM != nil && !math.IsInf(*endpoint.REDMetrics.Avg.TPM, 0) { // is not assigned when it is infinite
			newtpsRate.Value = endpoint.REDMetrics.Avg.TPM
		}

		newlatencyRadio := response.Ratio{
			DayOverDay:  endpoint.REDMetrics.DOD.Latency,
			WeekOverDay: endpoint.REDMetrics.WOW.Latency,
		}
		newlatencyRate := response.TempChartObject{
			Ratio: newlatencyRadio,
		}
		if endpoint.REDMetrics.Avg.Latency != nil && !math.IsInf(*endpoint.REDMetrics.Avg.Latency, 0) { // does not assign a value when it is infinite
			newlatencyRate.Value = endpoint.REDMetrics.Avg.Latency
		}

		// The filling error rate is equal to 0 and cannot be found year-on-year. The uniform filling is 0 (filling is performed by judging whether there is a request and if there is a request)
		if newlatencyRadio.DayOverDay != nil && newErrorRadio.DayOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value == 0 {
			newErrorRate.Ratio.DayOverDay = new(float64)
			*newErrorRate.Ratio.DayOverDay = 0
		}
		if newlatencyRadio.WeekOverDay != nil && newErrorRadio.WeekOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value == 0 {
			newErrorRate.Ratio.WeekOverDay = new(float64)
			*newErrorRate.Ratio.WeekOverDay = 0
		}
		// If the filling error rate is not equal to 0, no year-on-year comparison can be found, and the filling is the maximum value (filling is performed by judging whether there is a request and if there is a request)
		if newlatencyRadio.DayOverDay != nil && newErrorRadio.DayOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value != 0 {
			newErrorRate.Ratio.DayOverDay = new(float64)
			*newErrorRate.Ratio.DayOverDay = RES_MAX_VALUE
		}
		if newlatencyRadio.WeekOverDay != nil && newErrorRadio.WeekOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value != 0 {
			newErrorRate.Ratio.WeekOverDay = new(float64)
			*newErrorRate.Ratio.WeekOverDay = RES_MAX_VALUE
		}
		newServiceDetail := response.ServiceDetail{
			Endpoint:  endpoint.ContentKey,
			ErrorRate: newErrorRate,
			Tps:       newtpsRate,
			Latency:   newlatencyRate,
		}
		if endpoint.DelaySource == nil {
			newServiceDetail.DelaySource = "unknown"
		} else if endpoint.DelaySource != nil && *endpoint.DelaySource > 0.5 {
			newServiceDetail.DelaySource = "dependency"
		} else {
			newServiceDetail.DelaySource = "self"
		}
		if newServiceDetail.ErrorRate.Value == nil && newServiceDetail.Latency.Value == nil {
			continue
		}
		newServiceDetails = append(newServiceDetails, newServiceDetail)
	}
	return newServiceDetails
}
