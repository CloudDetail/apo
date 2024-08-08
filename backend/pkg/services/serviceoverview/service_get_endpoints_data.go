package serviceoverview

import (
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/database"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServicesEndPointData(startTime time.Time, endTime time.Time, step time.Duration, serviceNames string, sortRule SortType) (res []response.ServiceEndPointsRes, err error) {
	var Urls []Url
	var duration string
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"

	//start := time.Now()
	// 查询日同比，填充相应数据,传入servicesName不为空时则只查servicesName对应的url
	_, err = s.UrlDOD(&Urls, serviceNames, endTime, duration)
	//// 获取代码执行完后的时间
	//end := time.Now()
	//// 计算耗时
	//elapsed := end.Sub(start)
	//// 打印耗时
	//fmt.Printf("查询日同比代码执行时间: %s\n", elapsed)

	//start1 := time.Now()
	//查询所有url的平均值,传入servicesName不为空时则只查servicesName对应的url
	_, err = s.UrlAVG(&Urls, serviceNames, endTime, duration)
	//// 获取代码执行完后的时间
	//end1 := time.Now()
	//// 计算耗时
	//elapsed1 := end1.Sub(start1)
	//// 打印耗时
	//fmt.Printf("查询平均值代码执行时间: %s\n", elapsed1)

	//start2 := time.Now()
	//查询所有url的周同比,传入servicesName不为空时则只查servicesName对应的url
	_, err = s.UrlWOW(&Urls, serviceNames, endTime, duration)
	//// 获取代码执行完后的时间
	//end2 := time.Now()
	//// 计算耗时
	//elapsed2 := end2.Sub(start2)
	//// 打印耗时
	//fmt.Printf("查询周同比代码执行时间: %s\n", elapsed2)

	//start3 := time.Now()
	//查询延时时间，判断延时依赖关系
	_, err = s.UrlLatencySource(&Urls, serviceNames, startTime, endTime, duration, step)
	//// 获取代码执行完后的时间
	//end3 := time.Now()
	//// 计算耗时
	//elapsed3 := end3.Sub(start3)
	//// 打印耗时
	//fmt.Printf("查询延时时间代码执行时间: %s\n", elapsed3)

	//start4 := time.Now()
	//对所有的url进行排序
	switch sortRule {
	case DODThreshold: //按照日同比阈值排序
		threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
		if err != nil {

			return nil, err
		}
		errorThreshold := threshold.ErrorRate
		//不对吞吐量进行比较
		//tpsThreshold := threshold.Tps
		latencyThreshold := threshold.Latency
		for i := range Urls {
			//填充错误率不等于0，且有请求时查不出同比，填充为最大值（通过判断是否有请求，有请求进行填充）
			if Urls[i].LatencyDayOverDay != nil && Urls[i].ErrorRateDayOverDay == nil && Urls[i].AvgErrorRate != nil && *Urls[i].AvgErrorRate != 0 {
				Urls[i].ErrorRateDayOverDay = new(float64)
				*Urls[i].ErrorRateDayOverDay = RES_MAX_VALUE
			}
			if Urls[i].LatencyWeekOverWeek != nil && Urls[i].ErrorRateWeekOverWeek == nil && Urls[i].AvgErrorRate != nil && *Urls[i].AvgErrorRate != 0 {
				Urls[i].ErrorRateWeekOverWeek = new(float64)
				*Urls[i].ErrorRateWeekOverWeek = RES_MAX_VALUE
			}
			//过滤错误率
			if Urls[i].ErrorRateDayOverDay != nil && *Urls[i].ErrorRateDayOverDay > errorThreshold {
				Urls[i].IsErrorRateExceeded = true
				Urls[i].Count += ErrorCount
			}
			//过滤延时
			if Urls[i].LatencyDayOverDay != nil && *Urls[i].LatencyDayOverDay > latencyThreshold {
				Urls[i].IsLatencyExceeded = true
				Urls[i].Count += LatencyCount
			}
			////过滤TPS 不对吞吐量进行比较
			//if Urls[i].TPSDayOverDay != nil && *Urls[i].TPSDayOverDay > tpsThreshold {
			//	Urls[i].IsTPSExceeded = true
			//	Urls[i].Count += TPSCount
			//}
		}
		sortByDODThreshold(Urls)
	case MUTATIONSORT:
		_, err = s.UrlAVG1min(&Urls, serviceNames, endTime, duration)
		sortByMutation(Urls)
	}
	//// 获取代码执行完后的时间
	//end4 := time.Now()
	//// 计算耗时
	//elapsed4 := end4.Sub(start4)
	//// 打印耗时
	//fmt.Printf("url排序代码执行时间: %s\n", elapsed4)

	//start5 := time.Now()
	//将url存到对应的service中
	Services := fillServices(Urls)
	// 获取代码执行完后的时间
	//end5 := time.Now()
	//// 计算耗时
	//elapsed5 := end5.Sub(start5)
	//// 打印耗时
	//fmt.Printf("service排序代码执行时间: %s\n", elapsed5)

	//start6 := time.Now()
	//查询service前三url的所有数据,并填充
	s.UrlRangeData(&Services, startTime, endTime, duration, step)
	// 获取代码执行完后的时间
	//end6 := time.Now()
	//// 计算耗时
	//elapsed6 := end6.Sub(start6)
	//// 打印耗时
	//fmt.Printf("曲线图查询代码执行时间: %s\n", elapsed6)

	//填充数据
	var ServicesResMsg []response.ServiceEndPointsRes
	for _, service := range Services {
		if service.ServiceName == "" {
			continue
		}
		var newServiceDetails []response.ServiceDetail
		for _, url := range service.Urls {
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
				DayOverDay:  url.TPSDayOverDay,
				WeekOverDay: url.TPSWeekOverWeek,
			}
			newtpsRate := response.TempChartObject{
				//ChartData: map[int64]float64{},
				Ratio: newtpsRadio,
			}
			if url.AvgTPS != nil && !math.IsInf(*url.AvgTPS, 0) { //为无穷大时则不赋值
				newtpsRate.Value = url.AvgTPS
			}
			if url.TPSData != nil {
				data := make(map[int64]float64)
				// 将chartData转换为map
				for _, item := range url.TPSData {
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
				Endpoint:  url.ContentKey,
				ErrorRate: newErrorRate,
				Tps:       newtpsRate,
				Latency:   newlatencyRate,
			}
			if url.DelaySource != nil && *url.DelaySource > 0.5 {
				newServiceDetail.DelaySource = "dependency"
			} else {
				newServiceDetail.DelaySource = "self"
			}
			if newServiceDetail.ErrorRate.Value == nil && newServiceDetail.Latency.Value == nil {
				continue
			}
			newServiceDetails = append(newServiceDetails, newServiceDetail)
		}

		if newServiceDetails == nil {
			continue
		}

		newServiceRes := response.ServiceEndPointsRes{
			ServiceName:    service.ServiceName,
			EndpointCount:  service.EndpointCount,
			ServiceDetails: newServiceDetails,
		}

		ServicesResMsg = append(ServicesResMsg, newServiceRes)
	}
	return ServicesResMsg, err
}
