// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServicesEndpointDataByFilter(startTime time.Time, endTime time.Time, step time.Duration, filter EndpointsFilter, sortRule SortType) (res []response.ServiceEndPointsRes, err error) {
	var duration string
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
	filters := filter.ExtractFilterStr()
	// step1 Query that meets the Endpoint of the Filter and return the corresponding RED metric
	// RED metric contains the average, day-to-year rate of change and week-to-week rate of change over the selected time period
	endpointsMap := s.EndpointsREDMetric(startTime, endTime, filters)

	// step2 fill delay dependency
	err = s.EndpointsDelaySource(endpointsMap, startTime, endTime, filters)
	if err != nil {
		// TODO output error log, DelaySource query failed
	}

	// step2.. Fill Namespace information
	err = s.EndpointsNamespaceInfo(endpointsMap, startTime, endTime, filters)
	if err != nil {
		// TODO output error log, Namespace query failed
	}

	// step3 Sort the URL according to the sorting rule and fill in the data that has not been queried in the previous period.
	if sortRule == MUTATIONSORT {
		// Fill the real-time RED metric for sorting (the case between 3 minutes before the current time)
		s.EndpointsRealtimeREDMetric(filter, endpointsMap, startTime, endTime)
	}
	// Sort the endpoints according to the sorting rule and fill some unqueried data
	err = s.sortWithRule(sortRule, endpointsMap)

	// step4 Group Endpoints by service and maintain service ordering
	services := fillServices(endpointsMap.MetricGroupList)

	// step5 fills the RED chart data of the first three urls of each service group
	s.EndpointRangeREDChart(&services, startTime, endTime, duration, step)

	// step6 Fill null values and adjust the return structure
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

// GetServicesEndpointDataByEndpoints 根据输入的Endpoint填充端点总览信息,不按Service分组,并保留指标为空的Endpoint
// 当前只在告警分析时使用
func (s *service) GetServicesEndpointDataByEndpoints(startTime time.Time, endTime time.Time, step time.Duration, endpoints []prom.EndpointKey, sortRule SortType) (res []response.ServiceEndPointsRes, err error) {
	var duration string
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"

	filters := extractEndpointFiltersByEndpoints(endpoints)

	// step1 查询满足Filter的Endpoint,并返回对应的RED指标
	// RED指标包含了选定时间段内的平均值,日同比变化率和周同比变化率
	endpointsMap := s.EndpointsREDMetric(startTime, endTime, filters)

	// 为未能查到数据的endpoint填充空结构
	for _, endpoint := range endpoints {
		if _, find := endpointsMap.CheckKeyExist(endpoint); !find {
			endpointsMap.AppendEmptyGroup(endpoint)
		}
	}

	// step2 填充延时依赖关系
	err = s.EndpointsDelaySource(endpointsMap, startTime, endTime, filters)
	if err != nil {
		// TODO 输出错误日志, DelaySource查询失败
	}

	// step2.. 填充Namespace信息
	err = s.EndpointsNamespaceInfo(endpointsMap, startTime, endTime, filters)
	if err != nil {
		// TODO 输出错误日志, Namespace查询失败
	}

	// step3 根据排序规则对URL进行排序, 并填充前期未查询到的数据
	if sortRule == MUTATIONSORT {
		// 填充实时RED指标用于排序(当前时间往前3分钟之间的情况)
		s.EndpointsRealtimeREDMetric(filters, endpointsMap, startTime, endTime)
	}
	// 根据排序规则对endpoints进行排序并填充部分未查询到的数据
	err = s.sortWithRule(sortRule, endpointsMap)

	// step4 将Endpoints按service分组,并维持service排序
	services := fillEndpointsWithEmptyEndpoint(endpointsMap.MetricGroupList)

	// step5 填充RED图表数据
	s.EndpointRangeREDChart(&services, startTime, endTime, duration, step)

	// step6 填充空值并调整返回结构
	var servicesResMsg []response.ServiceEndPointsRes
	for _, service := range services {
		if service.ServiceName == "" {
			continue
		}
		serviceDetails := s.extractDetail(service, startTime, endTime, step)

		if serviceDetails == nil {
			serviceDetails = make([]response.ServiceDetail, 0)
			for _, endpoint := range service.Endpoints {
				serviceDetails = append(serviceDetails, response.ServiceDetail{
					Endpoint: endpoint.ContentKey,
				})
			}
		}

		// endpoint的namespaceList去重
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
			//// Filter TPS does not compare throughput
			//if Urls[i].TPSDayOverDay != nil && *Urls[i].TPSDayOverDay > tpsThreshold {
			//	Urls[i].IsTPSExceeded = true
			//	Urls[i].Count += TPSCount
			//}
		}
		sortByDODThreshold(endpointsMap.MetricGroupList)
	case MUTATIONSORT: //Sort by real-time mutation rate
		sortByMutation(endpointsMap.MetricGroupList)
	}

	return nil
}

func (*service) extractDetail(service ServiceDetail, startTime time.Time, endTime time.Time, step time.Duration) []response.ServiceDetail {
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
