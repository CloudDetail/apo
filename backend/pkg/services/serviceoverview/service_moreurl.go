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

	endpointsMap := s.EndpointsREDMetric(startTime, endTime, filter)
	endpoints := endpointsMap.Endpoints

	// step2 填充延时依赖关系
	err = s.EndpointsDelaySource(endpointsMap, startTime, endTime, filter)
	if err != nil {
		// TODO 输出错误日志, DelaySource查询失败
	}

	if len(endpoints) == 0 {
		// NOTE 通过moreUrl进入的请求,原则上不可能出现未查询到数据的情况
		// 不应该进入该分支
		return nil, nil
	}

	threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
	if err != nil {
		return nil, err
	}
	errorThreshold := threshold.ErrorRate
	//不对吞吐量进行比较
	//tpsThreshold := threshold.Tps
	latencyThreshold := threshold.Latency
	for i := range endpoints {
		//填充错误率不等于0查不出同比，填充为最大值（通过判断是否有请求，有请求进行填充）
		if endpoints[i].LatencyDayOverDay != nil && endpoints[i].ErrorRateDayOverDay == nil && endpoints[i].AvgErrorRate != nil && *endpoints[i].AvgErrorRate != 0 {
			endpoints[i].ErrorRateDayOverDay = new(float64)
			*endpoints[i].ErrorRateDayOverDay = RES_MAX_VALUE
		}
		if endpoints[i].LatencyWeekOverWeek != nil && endpoints[i].ErrorRateWeekOverWeek == nil && endpoints[i].AvgErrorRate != nil && *endpoints[i].AvgErrorRate != 0 {
			endpoints[i].ErrorRateWeekOverWeek = new(float64)
			*endpoints[i].ErrorRateWeekOverWeek = RES_MAX_VALUE
		}

		//过滤错误率
		if endpoints[i].ErrorRateDayOverDay != nil && *endpoints[i].ErrorRateDayOverDay > errorThreshold {
			endpoints[i].IsErrorRateExceeded = true
			endpoints[i].Count += ErrorCount
		}

		//过滤延时

		if endpoints[i].LatencyDayOverDay != nil && *endpoints[i].LatencyDayOverDay > latencyThreshold {
			endpoints[i].IsLatencyExceeded = true
			endpoints[i].Count += LatencyCount
		}
		//不对吞吐量进行比较
		////过滤TPS
		//
		//if Urls[i].TPSDayOverDay != nil && *Urls[i].TPSDayOverDay > tpsThreshold {
		//	Urls[i].IsTPSExceeded = true
		//	Urls[i].Count += TPSCount
		//}

	}
	//对所有的url进行排序
	switch sortRule {
	case DODThreshold: //按照日同比阈值排序
		sortByDODThreshold(endpoints)
	}

	//将所有url存到对应的service中
	services := fillOneService(endpoints)

	//查询service的所有url数据,并填充
	s.EndpointRangeREDChart(&services, startTime, endTime, duration, step)
	//(searchTime.Add(-30*time.Minute), searchTime, errorDataQuery, time.Minute)

	if len(services) == 0 {
		// NOTE 通过moreUrl进入的请求,原则上不可能出现未查询到数据的情况
		// DOUBLE CHECK
		return nil, nil
	}

	// NOTE 原则上进入这个入口的服务指定了Service,所以只会有一个
	service := services[0]
	var newServiceDetails []response.ServiceDetail
	for _, url := range service.Endpoints {
		newErrorRadio := response.Ratio{
			DayOverDay:  url.ErrorRateDayOverDay,
			WeekOverDay: url.ErrorRateWeekOverWeek,
		}
		newErrorRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newErrorRadio,
		}
		if url.AvgErrorRate != nil && !math.IsInf(*url.AvgErrorRate, 0) { //为无穷大时则不赋值
			newErrorRate.Value = url.AvgErrorRate
		}
		if url.ErrorRateData != nil {
			data := make(map[int64]float64)

			// 将chartData转换为map
			for _, item := range url.ErrorRateData {
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
			DayOverDay:  url.TPMDayOverDay,
			WeekOverDay: url.TPMWeekOverWeek,
		}
		newtpsRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newtpsRadio,
		}
		if url.AvgTPM != nil && !math.IsInf(*url.AvgTPM, 0) { //为无穷大时则不赋值
			newtpsRate.Value = url.AvgTPM
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
		if url.TPMData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range url.TPMData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
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
			DayOverDay:  url.LatencyDayOverDay,
			WeekOverDay: url.LatencyWeekOverWeek,
		}
		newlatencyRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newlatencyRadio,
		}
		if url.AvgLatency != nil && !math.IsInf(*url.AvgLatency, 0) { //为无穷大时则不赋值
			newlatencyRate.Value = url.AvgLatency
		}
		if url.LatencyData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range url.LatencyData {
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
			Endpoint:    url.ContentKey,
			ErrorRate:   newErrorRate,
			Tps:         newtpsRate,
			Latency:     newlatencyRate,
			DelaySource: "unknown",
		}
		newServiceDetails = append(newServiceDetails, newServiceDetail)
	}

	return newServiceDetails, err

}
