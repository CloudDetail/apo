// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/database"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceMoreUrl(startTime time.Time, endTime time.Time, step time.Duration, serviceNames string, sortRule SortType) (res []response.ServiceDetail, err error) {
	var duration string
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"

	filter := EndpointsFilter{
		ServiceName: serviceNames,
	}

	filters := filter.ExtractFilterStr()
	endpointsMap := s.EndpointsREDMetric(startTime, endTime, filters)
	endpoints := endpointsMap.MetricGroupList

	// step2 fill delay dependency
	err = s.EndpointsDelaySource(endpointsMap, startTime, endTime, filters)
	if err != nil {
		// TODO output error log, DelaySource query failed
	}

	if len(endpoints) == 0 {
		// NOTE requests entered through moreUrl. In principle, it is impossible to fail to query data.
		// should not enter this branch
		return nil, nil
	}

	threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
	if err != nil {
		return nil, err
	}
	errorThreshold := threshold.ErrorRate
	// No throughput comparison
	//tpsThreshold := threshold.Tps
	latencyThreshold := threshold.Latency
	for i := range endpoints {
		// If the filling error rate is not equal to 0, no year-on-year comparison can be found, and the filling is the maximum value (filling is performed by judging whether there is a request and if there is a request)
		if endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[i].REDMetrics.DOD.ErrorRate == nil && endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate != 0 {
			endpoints[i].REDMetrics.DOD.ErrorRate = new(float64)
			*endpoints[i].REDMetrics.DOD.ErrorRate = RES_MAX_VALUE
		}
		if endpoints[i].REDMetrics.WOW.Latency != nil && endpoints[i].REDMetrics.WOW.ErrorRate == nil && endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate != 0 {
			endpoints[i].REDMetrics.WOW.ErrorRate = new(float64)
			*endpoints[i].REDMetrics.WOW.ErrorRate = RES_MAX_VALUE
		}

		// Filter error rate
		if endpoints[i].REDMetrics.DOD.ErrorRate != nil && *endpoints[i].REDMetrics.DOD.ErrorRate > errorThreshold {
			endpoints[i].IsErrorRateExceeded = true
			endpoints[i].AlertCount += ErrorCount
		}

		// Filter delay

		if endpoints[i].REDMetrics.DOD.Latency != nil && *endpoints[i].REDMetrics.DOD.Latency > latencyThreshold {
			endpoints[i].IsLatencyExceeded = true
			endpoints[i].AlertCount += LatencyCount
		}
		// No throughput comparison
		//// Filter TPS
		//
		//if Urls[i].TPSDayOverDay != nil && *Urls[i].TPSDayOverDay > tpsThreshold {
		//	Urls[i].IsTPSExceeded = true
		//	Urls[i].Count += TPSCount
		//}

	}
	// Sort all URLs
	switch sortRule {
	case DODThreshold: //Sort by Day-to-Year Threshold
		sortByDODThreshold(endpoints)
	}

	// Save all URLs to the corresponding service
	services := fillOneService(endpoints)

	// query all url data of the service and populate
	s.EndpointRangeREDChart(&services, startTime, endTime, duration, step)
	//(searchTime.Add(-30*time.Minute), searchTime, errorDataQuery, time.Minute)

	if len(services) == 0 {
		// NOTE requests entered through moreUrl. In principle, it is impossible to fail to query data.
		// DOUBLE CHECK
		return nil, nil
	}

	// NOTE In principle, the service that enters this entrance has a specified Service, so there will only be one
	service := services[0]
	var newServiceDetails []response.ServiceDetail
	for _, url := range service.Endpoints {
		newErrorRadio := response.Ratio{
			DayOverDay:  url.REDMetrics.DOD.ErrorRate,
			WeekOverDay: url.REDMetrics.WOW.ErrorRate,
		}
		newErrorRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newErrorRadio,
		}
		if url.REDMetrics.Avg.ErrorRate != nil && !math.IsInf(*url.REDMetrics.Avg.ErrorRate, 0) { // does not assign a value when it is infinite
			newErrorRate.Value = url.REDMetrics.Avg.ErrorRate
		}
		if url.ErrorRateData != nil {
			data := make(map[int64]float64)

			// Convert chartData to map
			for _, item := range url.ErrorRateData {
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
			DayOverDay:  url.REDMetrics.DOD.TPM,
			WeekOverDay: url.REDMetrics.WOW.TPM,
		}
		newtpsRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newtpsRadio,
		}
		if url.REDMetrics.Avg.TPM != nil && !math.IsInf(*url.REDMetrics.Avg.TPM, 0) { // is not assigned when it is infinite
			newtpsRate.Value = url.REDMetrics.Avg.TPM
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
		if url.TPMData != nil {
			data := make(map[int64]float64)
			// Convert chartData to map
			for _, item := range url.TPMData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { // does not assign value when it is infinity
					data[timestamp] = value
				}
			}
			newtpsRate.ChartData = data
		}
		if newErrorRate.Value == nil && newtpsRate.Value != nil {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newErrorRate.ChartData = values
			newErrorRate.Value = new(float64)
			*newErrorRate.Value = 0
		}
		newlatencyRadio := response.Ratio{
			DayOverDay:  url.REDMetrics.DOD.Latency,
			WeekOverDay: url.REDMetrics.WOW.Latency,
		}
		newlatencyRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newlatencyRadio,
		}
		if url.REDMetrics.Avg.Latency != nil && !math.IsInf(*url.REDMetrics.Avg.Latency, 0) { // does not assign a value when it is infinite
			newlatencyRate.Value = url.REDMetrics.Avg.Latency
		}
		if url.LatencyData != nil {
			data := make(map[int64]float64)
			// Convert chartData to map
			for _, item := range url.LatencyData {
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

		delaySource := "unknown"
		if url.DelaySource == nil {
			delaySource = "unknown"
		} else if url.DelaySource != nil && *url.DelaySource > 0.5 {
			delaySource = "dependency"
		} else {
			delaySource = "self"
		}

		newServiceDetail := response.ServiceDetail{
			Endpoint:    url.ContentKey,
			ErrorRate:   newErrorRate,
			Tps:         newtpsRate,
			Latency:     newlatencyRate,
			DelaySource: delaySource,
		}
		newServiceDetails = append(newServiceDetails, newServiceDetail)
	}

	return newServiceDetails, err
}
