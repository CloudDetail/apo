package serviceoverview

import (
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}
func (s *service) GetServicesAlert(startTime time.Time, endTime time.Time, step time.Duration, serviceNames []string, returnData []string) (res []response.ServiceAlertRes, err error) {

	var Services []serviceDetail
	for i := 0; i < len(serviceNames); i++ {
		Services = append(Services, serviceDetail{
			ServiceName: serviceNames[i],
		})
	}
	_, err = s.QueryServicesInstanceByPod(&Services, startTime, endTime, step)
	_, err = s.QueryServicesInstanceByContainerId(&Services, startTime, endTime, step)
	_, err = s.QueryServicesInstanceByPid(&Services, startTime, endTime, step)
	if returnData == nil || contains(returnData, "logMetrics") {
		var duration string
		var stepNS = endTime.Sub(startTime).Nanoseconds()
		duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
		for i := range Services {
			var Pods []string
			for j := range Services[i].Instances {
				if Services[i].Instances[j].InstanceType == POD {
					Pods = append(Pods, Services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByPod(&Services[i].Instances, Pods, endTime, duration)
			_, err = s.LogDODByPod(&Services[i].Instances, Pods, endTime, duration)
			_, err = s.LogWOWByPod(&Services[i].Instances, Pods, endTime, duration)
			_, err = s.ServiceLogRangeDataByPod(&Services[i], Pods, startTime, endTime, duration, step)
		}
		//var Pods []string
		//var Instances []Instance
		//for i, _ := range Services {
		//	for j := range Services[i].Instances {
		//		if Services[i].Instances[j].InstanceType == POD {
		//			Pods = append(Pods, Services[i].Instances[j].ConvertName)
		//			Instances = append(Instances, Services[i].Instances[j])
		//		}
		//	}
		//	_, err = s.ServiceLogRangeDataByPod(&Services[i], Pods, startTime, endTime, duration, step)
		//}
		//_, err = s.AvgLogByPod(&Instances, Pods, endTime, duration)
		//_, err = s.LogDODByPod(&Instances, Pods, endTime, duration)
		//_, err = s.LogWOWByPod(&Instances, Pods, endTime, duration)
		//for i := range Services {
		//	for j := range Services[i].Instances {
		//		for k := range Instances {
		//			if Services[i].Instances[j].SvcName == Instances[k].SvcName && Services[i].Instances[j].InstanceType == POD && Services[i].Instances[j].ConvertName == Instances[k].ConvertName && Instances[k].Pod != "" {
		//				Services[i].Instances[j].AvgLog = Instances[k].AvgLog
		//				Services[i].Instances[j].LogDayOverDay = Instances[k].LogDayOverDay
		//				Services[i].Instances[j].LogWeekOverWeek = Instances[k].LogWeekOverWeek
		//			}
		//		}
		//
		//	}
		//}
		for i := range Services {
			var ContainerIds []string
			for j := range Services[i].Instances {
				if Services[i].Instances[j].InstanceType == NODE {
					ContainerIds = append(ContainerIds, Services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByContainerId(&Services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.LogDODByContainerId(&Services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.LogWOWByContainerId(&Services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.ServiceLogRangeDataByContainerId(&Services[i], ContainerIds, startTime, endTime, duration, step)
		}
		for i := range Services {
			var Pids []string
			for j := range Services[i].Instances {
				if Services[i].Instances[j].InstanceType == VM {
					Pids = append(Pids, Services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByPid(&Services[i].Instances, Pids, endTime, duration)
			_, err = s.LogDODByPid(&Services[i].Instances, Pids, endTime, duration)
			_, err = s.LogWOWByPid(&Services[i].Instances, Pids, endTime, duration)
			_, err = s.ServiceLogRangeDataByPid(&Services[i], Pids, startTime, endTime, duration, step)

		}
	}

	var ServicesAlertResMsg []response.ServiceAlertRes
	for _, service := range Services {
		if service.ServiceName == "" {
			continue
		}
		newlogs := response.TempChartObject{
			//ChartData: map[int64]float64{},
			Value: nil,
		}
		if service.LogData != nil {
			data := make(map[int64]float64)
			// 将chartData转换为map
			for _, item := range service.LogData {
				timestamp := item.TimeStamp
				value := item.Value
				data[timestamp] = value
			}
			newlogs.ChartData = data
		}
		if service.LogData == nil {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newlogs.ChartData = values
		}
		for _, instance := range service.Instances {
			if instance.LogDayOverDay != nil {
				if newlogs.Ratio.DayOverDay == nil {
					// 如果 newlogs.Ratio.DayOverDay 是 nil，需要先初始化
					newlogs.Ratio.DayOverDay = new(float64)
				}
				*newlogs.Ratio.DayOverDay += *instance.LogDayOverDay
			}
			if instance.LogWeekOverWeek != nil {
				if newlogs.Ratio.WeekOverDay == nil {
					// 如果 newlogs.Ratio.WeekOverDay 是 nil，需要先初始化
					newlogs.Ratio.WeekOverDay = new(float64)
				}
				*newlogs.Ratio.WeekOverDay += *instance.LogWeekOverWeek
			}
			if instance.AvgLog != nil {
				if newlogs.Value == nil {
					// 如果 newlogs.Value 是 nil，需要先初始化
					newlogs.Value = new(float64)
				}
				*newlogs.Value += *instance.AvgLog
			}
		}
		if newlogs.Value != nil && *newlogs.Value == 0 {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newlogs.ChartData = values
		}
		if newlogs.Value == nil {
			newlogs.Value = new(float64)
			*newlogs.Value = 0
		}

		newServiceRes := response.ServiceAlertRes{
			ServiceName:          service.ServiceName,
			Logs:                 newlogs,
			InfrastructureStatus: model.STATUS_NORMAL,
			NetStatus:            model.STATUS_NORMAL,
			K8sStatus:            model.STATUS_NORMAL,
		}
		var alertNames []string
		var nodeNames []string
		var pids []string
		var pods []string
		var containerIds []string
		for _, instance := range service.Instances {
			var networkNodeNames []string
			var networkPids []string
			var networkPods []string
			if instance.ConvertName != "" {
				alertNames = append(alertNames, instance.ConvertName)
			}
			if instance.NodeName != "" {
				nodeNames = append(nodeNames, instance.NodeName)
				networkNodeNames = append(networkNodeNames, instance.NodeName)
			}
			if instance.Pod != "" {
				pods = append(pods, instance.Pod)
				networkPods = append(networkPods, instance.Pod)
			}
			if instance.Pid != "" {
				pids = append(pids, instance.Pid)
				networkPids = append(networkPids, instance.Pid)
			}
			if instance.ContainerId != "" {
				containerIds = append(containerIds, instance.ContainerId)
			}

			var isAlert bool
			if returnData == nil || contains(returnData, "netStatus") {
				isAlert, err = s.chRepo.NetworkAlert(startTime, endTime, networkPods, networkNodeNames, networkPids)
				if isAlert {
					newServiceRes.NetStatus = model.STATUS_CRITICAL
				}
			}
		}
		if len(pids) > 0 || len(containerIds) > 0 {
			if returnData == nil || contains(returnData, "lastStartTime") {
				startTimeMap, _ := s.promRepo.QueryProcessStartTime(startTime, endTime, step, pids, containerIds)
				latestStartTime, found := GetLatestStartTime(startTimeMap, service.Instances)
				if found {
					newServiceRes.Timestamp = new(int64)
					*newServiceRes.Timestamp = latestStartTime * 1e6
				}
			}
		}
		if len(alertNames) > 0 {
			var isAlert bool
			if returnData == nil || contains(returnData, "k8sStatus") {
				isAlert, err = s.chRepo.K8sAlert(startTime, endTime, alertNames)
				if isAlert {
					newServiceRes.K8sStatus = model.STATUS_CRITICAL
				}
			}
		}
		if len(nodeNames) > 0 {
			var isAlert bool
			if returnData == nil || contains(returnData, "infraStatus") {
				isAlert, err = s.chRepo.InfrastructureAlert(startTime, endTime, nodeNames)
				if isAlert {
					newServiceRes.InfrastructureStatus = model.STATUS_CRITICAL
				}
			}

			if returnData == nil || contains(returnData, "k8sStatus") {
				isAlert, err = s.chRepo.K8sAlert(startTime, endTime, nodeNames)
				if isAlert {
					newServiceRes.K8sStatus = model.STATUS_CRITICAL
				}
			}
		}

		ServicesAlertResMsg = append(ServicesAlertResMsg, newServiceRes)
	}
	return ServicesAlertResMsg, err
}

func GetLatestStartTime(startTimeMap map[model.ServiceInstance]int64, instances []Instance) (int64, bool) {
	var latestStartTime int64
	for _, instance := range instances {
		// 容器只能采用containerId进行查询，采集到的容器Pid通常是1
		containerId := instance.ContainerId
		nodeName := instance.NodeName
		var queryInstance model.ServiceInstance
		if containerId != "" {
			queryInstance = model.ServiceInstance{
				ContainerId: containerId,
				NodeName:    nodeName,
			}
		} else {
			pidInt, _ := strconv.Atoi(instance.Pid)
			queryInstance = model.ServiceInstance{
				Pid:      int64(pidInt),
				NodeName: nodeName,
			}
		}
		if startTime, found := startTimeMap[queryInstance]; found {
			if startTime > latestStartTime {
				latestStartTime = startTime
			}
		}
	}
	if latestStartTime == 0 {
		return 0, false
	} else {
		return latestStartTime, true
	}
}
