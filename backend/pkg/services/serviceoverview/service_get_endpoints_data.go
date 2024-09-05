package serviceoverview

import (
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetServicesEndPointData(startTime time.Time, endTime time.Time, step time.Duration, filter EndpointsFilter, sortRule SortType) (res []response.ServiceEndPointsRes, err error) {
	var duration string
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"

	// step1 查询满足Filter的Endpoint,并返回对应的RED指标
	// RED指标包含了选定时间段内的平均值,日同比变化率和周同比变化率
	endpointsMap := s.EndpointsREDMetric(startTime, endTime, filter)

	// step2 填充延时依赖关系
	err = s.EndpointsDelaySource(endpointsMap, startTime, endTime, filter)
	if err != nil {
		// TODO 输出错误日志, DelaySource查询失败
	}

	// step2.. 填充Namespace信息
	err = s.EndpointsNamespaceInfo(endpointsMap, startTime, endTime, filter)
	if err != nil {
		// TODO 输出错误日志, Namespace查询失败
	}

	// step3 根据排序规则对URL进行排序, 并填充前期未查询到的数据
	if sortRule == MUTATIONSORT {
		// 填充实时RED指标用于排序(当前时间往前3分钟之间的情况)
		s.EndpointsRealtimeREDMetric(filter, endpointsMap, startTime, endTime)
	}
	// 根据排序规则对endpoints进行排序并填充部分未查询到的数据
	err = s.sortWithRule(sortRule, endpointsMap)

	// step4 将Endpoints按service分组,并维持service排序
	services := fillServices(endpointsMap.Endpoints)

	// step5 填充每个service分组前三url的RED图表数据
	s.EndpointRangeREDChart(&services, startTime, endTime, duration, step)

	// step6 填充空值并调整返回结构
	var servicesResMsg []response.ServiceEndPointsRes
	for _, service := range services {
		if service.ServiceName == "" {
			continue
		}
		serviceDetails := s.extractDetail(service, startTime, endTime, step)

		if serviceDetails == nil {
			continue
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
	case DODThreshold: //按照日同比阈值排序
		threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
		if err != nil {
			return err
		}
		errorThreshold := threshold.ErrorRate
		//不对吞吐量进行比较
		//tpsThreshold := threshold.Tps
		latencyThreshold := threshold.Latency
		for i := range endpointsMap.Endpoints {
			endpoint := endpointsMap.Endpoints[i]

			//填充错误率不等于0，且有请求时查不出同比，填充为最大值（通过判断是否有请求，有请求进行填充）
			if endpoint.LatencyDayOverDay != nil && endpoint.ErrorRateDayOverDay == nil && endpoint.AvgErrorRate != nil && *endpoint.AvgErrorRate != 0 {
				endpoint.ErrorRateDayOverDay = new(float64)
				*endpoint.ErrorRateDayOverDay = RES_MAX_VALUE
			}
			if endpoint.LatencyWeekOverWeek != nil && endpoint.ErrorRateWeekOverWeek == nil && endpoint.AvgErrorRate != nil && *endpoint.AvgErrorRate != 0 {
				endpoint.ErrorRateWeekOverWeek = new(float64)
				*endpoint.ErrorRateWeekOverWeek = RES_MAX_VALUE
			}
			//过滤错误率
			if endpoint.ErrorRateDayOverDay != nil && *endpoint.ErrorRateDayOverDay > errorThreshold {
				endpoint.IsErrorRateExceeded = true
				endpoint.Count += ErrorCount
			}
			//过滤延时
			if endpoint.LatencyDayOverDay != nil && *endpoint.LatencyDayOverDay > latencyThreshold {
				endpoint.IsLatencyExceeded = true
				endpoint.Count += LatencyCount
			}
			////过滤TPS 不对吞吐量进行比较
			//if Urls[i].TPSDayOverDay != nil && *Urls[i].TPSDayOverDay > tpsThreshold {
			//	Urls[i].IsTPSExceeded = true
			//	Urls[i].Count += TPSCount
			//}
		}
		sortByDODThreshold(endpointsMap.Endpoints)
	case MUTATIONSORT: // 按照实时突变率排序
		sortByMutation(endpointsMap.Endpoints)
	}

	return nil
}

func (*service) extractDetail(service ServiceDetail, startTime time.Time, endTime time.Time, step time.Duration) []response.ServiceDetail {
	var newServiceDetails []response.ServiceDetail
	for _, endpoint := range service.Endpoints {
		newErrorRadio := response.Ratio{
			DayOverDay:  endpoint.ErrorRateDayOverDay,
			WeekOverDay: endpoint.ErrorRateWeekOverWeek,
		}
		newErrorRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newErrorRadio,
		}
		if endpoint.AvgErrorRate != nil && !math.IsInf(*endpoint.AvgErrorRate, 0) { //为无穷大时则不赋值
			newErrorRate.Value = endpoint.AvgErrorRate
		}
		if endpoint.ErrorRateData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range endpoint.ErrorRateData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
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
			DayOverDay:  endpoint.TPMDayOverDay,
			WeekOverDay: endpoint.TPMWeekOverWeek,
		}
		newtpsRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newtpsRadio,
		}
		if endpoint.AvgTPM != nil && !math.IsInf(*endpoint.AvgTPM, 0) { //为无穷大时则不赋值
			newtpsRate.Value = endpoint.AvgTPM
		}
		if endpoint.TPMData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range endpoint.TPMData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					data[timestamp] = value
				}
			}
			newtpsRate.ChartData = data
		}
		//没有查询到数据，is_error=true，填充为0
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
			DayOverDay:  endpoint.LatencyDayOverDay,
			WeekOverDay: endpoint.LatencyWeekOverWeek,
		}
		newlatencyRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newlatencyRadio,
		}
		if endpoint.AvgLatency != nil && !math.IsInf(*endpoint.AvgLatency, 0) { //为无穷大时则不赋值
			newlatencyRate.Value = endpoint.AvgLatency
		}
		if endpoint.LatencyData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range endpoint.LatencyData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					data[timestamp] = value
				}
			}
			newlatencyRate.ChartData = data
		}
		//填充错误率等于0查不出同比，统一填充为0（通过判断是否有请求，有请求进行填充）
		if newlatencyRadio.DayOverDay != nil && newErrorRadio.DayOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value == 0 {
			newErrorRate.Ratio.DayOverDay = new(float64)
			*newErrorRate.Ratio.DayOverDay = 0
		}
		if newlatencyRadio.WeekOverDay != nil && newErrorRadio.WeekOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value == 0 {
			newErrorRate.Ratio.WeekOverDay = new(float64)
			*newErrorRate.Ratio.WeekOverDay = 0
		}
		//填充错误率不等于0查不出同比，填充为最大值（通过判断是否有请求，有请求进行填充）
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
