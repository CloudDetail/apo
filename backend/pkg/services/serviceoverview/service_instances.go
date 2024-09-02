package serviceoverview

import (
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"

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
	var Instances []Instance
	var stepNS = endTime.Sub(startTime).Nanoseconds()
	duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
	serviceName = prom.EscapeRegexp(serviceName)
	// 查询 Prometheus 数据 svc_name 和 content_key 对应的 node_name 的日同比，周同比，平均值，曲线图
	_, err = s.InstanceAVGByPod(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceDODByPod(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceWOWByPod(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceRangeDataByPod(&Instances, endPoint, serviceName, startTime, endTime, duration, step)

	_, err = s.InstanceAVGByContainerId(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceDODByContainerId(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceWOWByContainerId(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceRangeDataByContainerId(&Instances, endPoint, serviceName, startTime, endTime, duration, step)

	_, err = s.InstanceAVGByPid(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceDODByPid(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceWOWByPid(&Instances, serviceName, endPoint, endTime, duration)
	_, err = s.InstanceRangeDataByPid(&Instances, endPoint, serviceName, startTime, endTime, duration, step)
	var allPids []string
	var containerIds []string
	var pods []string
	for _, instance := range Instances {
		allPids = append(allPids, instance.Pid)
		if instance.ContainerId != "" {
			containerIds = append(containerIds, instance.ContainerId)
		}
		if instance.Pod != "" {
			pods = append(pods, instance.Pod)
		}
	}
	instanceStartTimeMap, _ := s.promRepo.QueryProcessStartTime(startTime, endTime, step, allPids, containerIds)

	_, err = s.AvgLogByPod(&Instances, pods, endTime, duration)
	_, err = s.LogDODByPod(&Instances, pods, endTime, duration)
	_, err = s.LogWOWByPod(&Instances, pods, endTime, duration)
	_, err = s.LogRangeDataByPod(&Instances, pods, startTime, endTime, duration, step)

	_, err = s.AvgLogByContainerId(&Instances, containerIds, endTime, duration)
	_, err = s.LogDODByContainerId(&Instances, containerIds, endTime, duration)
	_, err = s.LogWOWByContainerId(&Instances, containerIds, endTime, duration)
	_, err = s.LogRangeDataByContainerId(&Instances, containerIds, startTime, endTime, duration, step)
	var vmPids []string
	for i := range Instances {
		if Instances[i].InstanceType == VM {
			vmPids = append(vmPids, Instances[i].Pid)
		}
	}
	_, err = s.AvgLogByPid(&Instances, vmPids, endTime, duration)
	_, err = s.LogDODByPid(&Instances, vmPids, endTime, duration)
	_, err = s.LogWOWByPid(&Instances, vmPids, endTime, duration)
	_, err = s.LogRangeDataByPid(&Instances, vmPids, startTime, endTime, duration, step)
	res.Status = model.STATUS_NORMAL
	for i := range Instances {
		if Instances[i].ErrorRateDayOverDay != nil && *Instances[i].ErrorRateDayOverDay > errorThreshold {
			Instances[i].IsErrorRateDODExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if Instances[i].LatencyDayOverDay != nil && *Instances[i].LatencyDayOverDay > latencyThreshold {
			Instances[i].IsLatencyDODExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if Instances[i].TPSDayOverDay != nil && *Instances[i].TPSDayOverDay > tpsThreshold {
			Instances[i].IsTPSDODExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
	}
	for i := range Instances {
		if Instances[i].ErrorRateWeekOverWeek != nil && *Instances[i].ErrorRateWeekOverWeek > errorThreshold {
			Instances[i].IsErrorRateWOWExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if Instances[i].LatencyWeekOverWeek != nil && *Instances[i].LatencyWeekOverWeek > latencyThreshold {
			Instances[i].IsLatencyWOWExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
		if Instances[i].TPSWeekOverWeek != nil && *Instances[i].TPSWeekOverWeek > tpsThreshold {
			Instances[i].IsTPSWOWExceeded = true
			res.Status = model.STATUS_CRITICAL
		}
	}
	var ResData []response.InstanceData
	for _, InstanceTmp := range Instances {
		if (InstanceTmp.InstanceName == "") || (InstanceTmp.InstanceName == "@@") {
			continue
		}
		//过滤空数据
		if (InstanceTmp.AvgLatency == nil && InstanceTmp.AvgTPS == nil) || (InstanceTmp.AvgLatency == nil && InstanceTmp.AvgTPS != nil && *InstanceTmp.AvgTPS == 0) {
			continue
		}
		newErrorRadio := response.Ratio{
			DayOverDay:  InstanceTmp.ErrorRateDayOverDay,
			WeekOverDay: InstanceTmp.ErrorRateWeekOverWeek,
		}
		newErrorRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newErrorRadio,
		}
		if InstanceTmp.AvgErrorRate != nil && !math.IsInf(*InstanceTmp.AvgErrorRate, 0) { //为无穷大时则不赋值
			newErrorRate.Value = InstanceTmp.AvgErrorRate
		}
		if InstanceTmp.ErrorRateData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range InstanceTmp.ErrorRateData {
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
			DayOverDay:  InstanceTmp.TPSDayOverDay,
			WeekOverDay: InstanceTmp.TPSWeekOverWeek,
		}
		newtpsRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newtpsRadio,
		}
		if InstanceTmp.AvgTPS != nil && !math.IsInf(*InstanceTmp.AvgTPS, 0) { //为无穷大时则不赋值
			newtpsRate.Value = InstanceTmp.AvgTPS
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
		if InstanceTmp.TPSData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range InstanceTmp.TPSData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					data[timestamp] = value
				}
			}
			newtpsRate.ChartData = data
		}
		newlatencyRadio := response.Ratio{
			DayOverDay:  InstanceTmp.LatencyDayOverDay,
			WeekOverDay: InstanceTmp.LatencyWeekOverWeek,
		}
		newlatencyRate := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Ratio: newlatencyRadio,
		}
		if InstanceTmp.AvgLatency != nil && !math.IsInf(*InstanceTmp.AvgLatency, 0) { //为无穷大时则不赋值
			newlatencyRate.Value = InstanceTmp.AvgLatency
		}
		if InstanceTmp.LatencyData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range InstanceTmp.LatencyData {
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
		newLogRadio := response.Ratio{
			DayOverDay:  InstanceTmp.LogDayOverDay,
			WeekOverDay: InstanceTmp.LogWeekOverWeek,
		}
		newlogs := response.TempChartObject{
			Value: InstanceTmp.AvgLog,
			Ratio: newLogRadio,
		}
		if newlogs.Value == nil {
			newlogs.Value = new(float64)
			*newlogs.Value = 0
		}

		if InstanceTmp.LogData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range InstanceTmp.LogData {
				timestamp := item.TimeStamp
				value := item.Value
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					data[timestamp] = value
				}
			}
			newlogs.ChartData = data
		}
		//日志曲线图没有数据则进行填充
		if InstanceTmp.LogData == nil {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newlogs.ChartData = values
		}
		newInstance := response.InstanceData{
			Name:                 InstanceTmp.InstanceName,
			Namespace:            InstanceTmp.Namespace,
			InfrastructureStatus: model.STATUS_NORMAL,
			NetStatus:            model.STATUS_NORMAL,
			K8sStatus:            model.STATUS_NORMAL,
			Timestamp:            nil,
			Latency:              newlatencyRate,
			Tps:                  newtpsRate,
			ErrorRate:            newErrorRate,
			Logs:                 newlogs,
		}

		var names []string
		var NodeNames []string
		var Pids []string
		var Pods []string
		var InfrastructureAlertNames []string
		if InstanceTmp.NodeName != "" {
			names = append(names, InstanceTmp.NodeName)
			InfrastructureAlertNames = append(InfrastructureAlertNames, InstanceTmp.NodeName)
			NodeNames = append(NodeNames, InstanceTmp.NodeName)
		}
		if InstanceTmp.Pod != "" {
			names = append(names, InstanceTmp.Pod)
			Pods = append(Pods, InstanceTmp.Pod)
		}
		if InstanceTmp.Pid != "" {
			Pids = append(Pids, InstanceTmp.Pid)
		}
		if len(InfrastructureAlertNames) > 0 {
			var isAlert bool
			isAlert, err = s.chRepo.InfrastructureAlert(startTime, endTime, InfrastructureAlertNames)
			if isAlert {
				newInstance.InfrastructureStatus = model.STATUS_CRITICAL
			}
		}
		if len(Pods) > 0 || (len(NodeNames) > 0 && len(Pids) > 0) {
			var isAlert bool
			isAlert, err = s.chRepo.NetworkAlert(startTime, endTime, Pods, NodeNames, Pids)
			if isAlert {
				newInstance.NetStatus = model.STATUS_CRITICAL
			}
		}
		if len(names) > 0 {
			var isAlert bool
			isAlert, err = s.chRepo.K8sAlert(startTime, endTime, names)
			if isAlert {
				newInstance.K8sStatus = model.STATUS_CRITICAL
			}
		}
		if InstanceTmp.ContainerId != "" && InstanceTmp.NodeName != "" {
			tmpInstance := model.ServiceInstance{
				ContainerId: InstanceTmp.ContainerId,
				NodeName:    InstanceTmp.NodeName,
			}
			startTime, ok := instanceStartTimeMap[tmpInstance]
			if ok {
				newInstance.Timestamp = new(int64)
				*newInstance.Timestamp = startTime * 1e6
			}
		} else if InstanceTmp.Pid != "" && InstanceTmp.NodeName != "" {
			pidInt, _ := strconv.Atoi(InstanceTmp.Pid)
			tmpInstance := model.ServiceInstance{
				Pid:      int64(pidInt),
				NodeName: InstanceTmp.NodeName,
			}
			startTime, ok := instanceStartTimeMap[tmpInstance]
			if ok {
				newInstance.Timestamp = new(int64)
				*newInstance.Timestamp = startTime * 1e6
			}
		}

		ResData = append(ResData, newInstance)
	}
	res.Data = ResData

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
