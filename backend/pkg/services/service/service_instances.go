package service

import (
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetInstances(startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error) {
	threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
	if err != nil {

		return res, err
	}
	errorThreshold := threshold.ErrorRate
	tpsThreshold := threshold.Tps
	latencyThreshold := threshold.Latency
	var duration string
	var instances []serviceoverview.Instance
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
	serviceName = prom.EscapeRegexp(serviceName)
	// 查询 Prometheus 数据 svc_name 和 content_key 对应的 node_name 的日同比，周同比，平均值，曲线图
	_, err = s.InstanceAVGByPod(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceDODByPod(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceWOWByPod(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceRangeDataByPod(&instances, endPoint, serviceName, startTime, endTime, duration, step)

	_, err = s.InstanceAVGByContainerId(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceDODByContainerId(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceWOWByContainerId(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceRangeDataByContainerId(&instances, endPoint, serviceName, startTime, endTime, duration, step)

	_, err = s.InstanceAVGByPid(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceDODByPid(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceWOWByPid(&instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceRangeDataByPid(&instances, endPoint, serviceName, startTime, endTime, duration, step)

	var allPids []string
	var containerIds []string
	var pods []string
	for _, instance := range instances {
		allPids = append(allPids, instance.Pid)
		if instance.ContainerId != "" {
			containerIds = append(containerIds, instance.ContainerId)
		}
		if instance.Pod != "" {
			pods = append(pods, instance.Pod)
		}
	}

	_, err = s.AvgLogByPod(&instances, pods, endTime, duration)
	_, err = s.LogDODByPod(&instances, pods, endTime, duration)
	_, err = s.LogWOWByPod(&instances, pods, endTime, duration)
	_, err = s.LogRangeDataByPod(&instances, pods, startTime, endTime, duration, step)

	_, err = s.AvgLogByContainerId(&instances, containerIds, endTime, duration)
	_, err = s.LogDODByContainerId(&instances, containerIds, endTime, duration)
	_, err = s.LogWOWByContainerId(&instances, containerIds, endTime, duration)
	_, err = s.LogRangeDataByContainerId(&instances, containerIds, startTime, endTime, duration, step)
	var vmPids []string
	for i := range instances {
		if instances[i].InstanceType == serviceoverview.VM {
			vmPids = append(vmPids, instances[i].Pid)
		}
	}
	_, err = s.AvgLogByPid(&instances, vmPids, endTime, duration)
	_, err = s.LogDODByPid(&instances, vmPids, endTime, duration)
	_, err = s.LogWOWByPid(&instances, vmPids, endTime, duration)
	_, err = s.LogRangeDataByPid(&instances, vmPids, startTime, endTime, duration, step)
	res.Status = model.STATUS_NORMAL
	for i := range instances {
		if instances[i].ErrorRateDayOverDay != nil && *instances[i].ErrorRateDayOverDay > errorThreshold {
			instances[i].IsErrorRateDODExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if instances[i].LatencyDayOverDay != nil && *instances[i].LatencyDayOverDay > latencyThreshold {
			instances[i].IsLatencyDODExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if instances[i].TPSDayOverDay != nil && *instances[i].TPSDayOverDay > tpsThreshold {
			instances[i].IsTPSDODExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
	}
	for i := range instances {
		if instances[i].ErrorRateWeekOverWeek != nil && *instances[i].ErrorRateWeekOverWeek > errorThreshold {
			instances[i].IsErrorRateWOWExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if instances[i].LatencyWeekOverWeek != nil && *instances[i].LatencyWeekOverWeek > latencyThreshold {
			instances[i].IsLatencyWOWExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if instances[i].TPSWeekOverWeek != nil && *instances[i].TPSWeekOverWeek > tpsThreshold {
			instances[i].IsTPSWOWExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
	}
	var resData []response.InstanceData
	for _, instance := range instances {
		if (instance.InstanceName == "") || (instance.InstanceName == "@@") {
			continue
		}
		//过滤空数据
		if (instance.AvgLatency == nil && instance.AvgTPS == nil) || (instance.AvgLatency == nil && instance.AvgTPS != nil && *instance.AvgTPS == 0) {
			continue
		}
		newErrorRadio := response.Ratio{
			DayOverDay:  instance.ErrorRateDayOverDay,
			WeekOverDay: instance.ErrorRateWeekOverWeek,
		}
		newErrorRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newErrorRadio,
		}
		if instance.AvgErrorRate != nil && !math.IsInf(*instance.AvgErrorRate, 0) { //为无穷大时则不赋值
			newErrorRate.Value = instance.AvgErrorRate
		}
		if instance.ErrorRateData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range instance.ErrorRateData {
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
			DayOverDay:  instance.TPSDayOverDay,
			WeekOverDay: instance.TPSWeekOverWeek,
		}
		newtpsRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newtpsRadio,
		}
		if instance.AvgTPS != nil && !math.IsInf(*instance.AvgTPS, 0) { //为无穷大时则不赋值
			newtpsRate.Value = instance.AvgTPS
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
		if instance.TPSData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range instance.TPSData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					data[timestamp] = value
				}
			}
			newtpsRate.ChartData = data
		}
		newlatencyRadio := response.Ratio{
			DayOverDay:  instance.LatencyDayOverDay,
			WeekOverDay: instance.LatencyWeekOverWeek,
		}
		newlatencyRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newlatencyRadio,
		}
		if instance.AvgLatency != nil && !math.IsInf(*instance.AvgLatency, 0) { //为无穷大时则不赋值
			newlatencyRate.Value = instance.AvgLatency
		}
		if instance.LatencyData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range instance.LatencyData {
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
			*newErrorRate.Ratio.DayOverDay = serviceoverview.RES_MAX_VALUE
		}
		if newlatencyRadio.WeekOverDay != nil && newErrorRadio.WeekOverDay == nil && newErrorRate.Value != nil && *newErrorRate.Value != 0 {
			newErrorRate.Ratio.WeekOverDay = new(float64)
			*newErrorRate.Ratio.WeekOverDay = serviceoverview.RES_MAX_VALUE
		}
		newLogRadio := response.Ratio{
			DayOverDay:  instance.LogDayOverDay,
			WeekOverDay: instance.LogWeekOverWeek,
		}
		newlogs := response.TempChartObject{
			Value: instance.AvgLog,
			Ratio: newLogRadio,
		}
		if newlogs.Value == nil {
			newlogs.Value = new(float64)
			*newlogs.Value = 0
		}

		if instance.LogData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range instance.LogData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					data[timestamp] = value
				}
			}
			newlogs.ChartData = data
		}
		//日志曲线图没有数据则进行填充
		if instance.LogData == nil {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newlogs.ChartData = values
		}
		newInstance := response.InstanceData{
			Name:        instance.InstanceName,
			Namespace:   instance.Namespace,
			NodeName:    instance.NodeName,
			NodeIP:      instance.NodeIP,
			Timestamp:   nil,
			Latency:     newlatencyRate,
			Tps:         newtpsRate,
			ErrorRate:   newErrorRate,
			Logs:        newlogs,
			AlertStatus: model.NORMAL_ALERT_STATUS,
			AlertReason: model.AlertReason{},
		}

		pidI64, err := strconv.ParseInt(instance.Pid, 10, 64)
		if err != nil {
			pidI64 = -1
		}

		instanceSingleList := []*model.ServiceInstance{
			{
				ServiceName: serviceName,
				ContainerId: instance.ContainerId,
				PodName:     instance.Pod,
				Namespace:   instance.Namespace,
				NodeName:    instance.NodeName,
				Pid:         pidI64,
			},
		}
		// 填充告警状态
		newInstance.AlertStatusCH = serviceoverview.GetAlertStatusCH(
			s.chRepo, &newInstance.AlertReason, nil,
			nil, serviceName, instanceSingleList,
			startTime, endTime,
		)

		// 填充末次启动时间
		startTSmap, _ := s.promRepo.QueryProcessStartTime(startTime, endTime, instanceSingleList)
		latestStartTime := getLatestStartTime(startTSmap) * 1e6
		if latestStartTime > 0 {
			newInstance.Timestamp = &latestStartTime
		}

		resData = append(resData, newInstance)
	}
	res.Data = resData

	for _, data := range res.Data {
		if data.InfrastructureStatus == model.STATUS_CRITICAL {
			res.Status = model.STATUS_CRITICAL
			break
		}
		if data.NetStatus == model.STATUS_CRITICAL {
			res.Status = model.STATUS_CRITICAL
		}
		if data.K8sStatus == model.STATUS_CRITICAL {
			res.Status = model.STATUS_CRITICAL
		}

	}
	return res, err
}
