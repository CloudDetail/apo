// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"math"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"go.uber.org/zap"
)

// TODO move to prometheus package and avoid to repeated self
func (s *service) GetServicesEndpointDataWithChart(
	startTime time.Time, endTime time.Time, step time.Duration,
	filter EndpointsFilter, sortRule request.SortType,
) (res []response.ServiceEndPointsRes, err error) {
	filtersStr := filter.ExtractFilterStr()

	var opts = []prometheus.FetchEMOption{
		prometheus.WithREDMetric(),
		prometheus.WithDelaySource(),
		prometheus.WithNamespace(),
		prometheus.WithREDChart(step),
	}

	if sortRule == request.SortByLogErrorCount {
		opts = append(opts, prometheus.WithLogErrorCount())
	} else if sortRule == request.MUTATIONSORT {
		opts = append(opts, prometheus.WithRealTimeREDMetric())
	}

	endpointsMap, err := prometheus.FetchEndpointsData(
		s.promRepo, filtersStr, startTime, endTime,
		opts...,
	)

	if err != nil {
		s.logger.Error("failed to fetch endpoints data form", zap.Error(err))
		return
	}
	s.sortWithRule(sortRule, endpointsMap)

	s.sortWithRule(sortRule, endpointsMap)

	// step4 Group Endpoints by service and maintain service ordering
	services := groupEndpointsByService(endpointsMap.MetricGroupList, 3)

	// step5 Fill null values and adjust the return structure
	var servicesResMsg []response.ServiceEndPointsRes
	for _, service := range services {
		if service.ServiceName == "" {
			continue
		}
		serviceDetails := s.extractDetailWithChart(service, startTime, endTime, step)

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

func (*service) extractDetailWithChart(
	service *ServiceDetail,
	startTime time.Time, endTime time.Time, step time.Duration,
) []response.ServiceDetail {
	var newServiceDetails []response.ServiceDetail
	for _, endpoint := range service.Endpoints {
		newErrorRadio := response.Ratio{
			DayOverDay:  endpoint.REDMetrics.DOD.ErrorRate,
			WeekOverDay: endpoint.REDMetrics.WOW.ErrorRate,
		}
		newErrorRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newErrorRadio,
		}
		if endpoint.REDMetrics.Avg.ErrorRate != nil && !math.IsInf(*endpoint.REDMetrics.Avg.ErrorRate, 0) { // does not assign a value when it is infinite
			newErrorRate.Value = endpoint.REDMetrics.Avg.ErrorRate
		}
		if endpoint.ErrorRateData != nil {
			data := make(map[int64]float64)
			// Convert chartData to map
			for _, item := range endpoint.ErrorRateData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					data[timestamp] = value
				}
			}
			newErrorRate.ChartData = data
		}
		if newErrorRate.Value != nil && *newErrorRate.Value == 100 {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 100
			}
			newErrorRate.ChartData = values
		}
		if newErrorRate.Value != nil && *newErrorRate.Value == 0 {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newErrorRate.ChartData = values
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
		if endpoint.TPMData != nil {
			data := make(map[int64]float64)
			// Convert chartData to map
			for _, item := range endpoint.TPMData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					data[timestamp] = value
				}
			}
			newtpsRate.ChartData = data
		}
		// No data found, is_error = true, filled with 0
		if newErrorRate.Value == nil && newtpsRate.Value != nil {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newErrorRate.ChartData = values
			newErrorRate.Value = new(float64)
			*newErrorRate.Value = 0
		}
		if newErrorRate.Value != nil && *newErrorRate.Value == 0 {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newErrorRate.ChartData = values
		}

		newlatencyRadio := response.Ratio{
			DayOverDay:  endpoint.REDMetrics.DOD.Latency,
			WeekOverDay: endpoint.REDMetrics.WOW.Latency,
		}
		newlatencyRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newlatencyRadio,
		}
		if endpoint.REDMetrics.Avg.Latency != nil && !math.IsInf(*endpoint.REDMetrics.Avg.Latency, 0) { // does not assign a value when it is infinite
			newlatencyRate.Value = endpoint.REDMetrics.Avg.Latency
		}
		if endpoint.LatencyData != nil {
			data := make(map[int64]float64)
			// Convert chartData to map
			for _, item := range endpoint.LatencyData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					data[timestamp] = value
				}
			}
			newlatencyRate.ChartData = data
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
