package serviceoverview

import (
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
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
	var Services []ServiceDetail
	for i := 0; i < len(serviceNames); i++ {
		Services = append(Services, ServiceDetail{
			ServiceName: serviceNames[i],
		})
	}
	instances, err := s.promRepo.GetMultiServicesInstanceList(startTime.UnixMicro(), endTime.UnixMicro(), serviceNames)
	if err != nil {
		return nil, err
	}
	for i, svc := range serviceNames {
		serviceInstances := instances[svc]
		for _, instance := range serviceInstances.GetInstances() {
			var convertName string
			var instanceType int
			if instance.PodName != "" {
				convertName = instance.PodName
				instanceType = POD
			} else if instance.ContainerId != "" {
				convertName = instance.ContainerId
				instanceType = CONTAINER
			} else {
				convertName = strconv.FormatInt(instance.Pid, 10)
				instanceType = VM
			}
			newInstance := Instance{
				ConvertName:  convertName,
				InstanceType: instanceType,
				NodeName:     instance.NodeName,
				Pod:          instance.PodName,
				ContainerId:  instance.ContainerId,
				SvcName:      instance.ServiceName,
				Pid:          strconv.FormatInt(instance.Pid, 10),
			}
			Services[i].Instances = append(Services[i].Instances, newInstance)
		}
	}

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
		for i := range Services {
			var ContainerIds []string
			for j := range Services[i].Instances {
				if Services[i].Instances[j].InstanceType == CONTAINER {
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

	var servicesAlertResMsg []response.ServiceAlertRes
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
			ServiceName: service.ServiceName,
			Logs:        newlogs,
			AlertStatus: model.NORMAL_ALERT_STATUS,
			AlertReason: model.AlertReason{},
		}

		var serviceInstances []*model.ServiceInstance
		for _, instance := range service.Instances {
			pidI64, err := strconv.ParseInt(instance.Pid, 10, 64)
			if err != nil {
				pidI64 = -1
			}
			serviceInstances = append(serviceInstances, &model.ServiceInstance{
				ServiceName: service.ServiceName,
				ContainerId: instance.ContainerId,
				PodName:     instance.Pod,
				Namespace:   instance.Namespace,
				NodeName:    instance.NodeName,
				Pid:         pidI64,
			})
		}

		// 填充告警状态
		newServiceRes.AlertStatusCH = GetAlertStatusCH(
			s.chRepo, &newServiceRes.AlertReason, nil,
			returnData, service.ServiceName, serviceInstances,
			startTime, endTime,
		)

		// 填充末次启动时间
		if returnData == nil || contains(returnData, "lastStartTime") {
			startTSmap, _ := s.promRepo.QueryProcessStartTime(startTime, endTime, serviceInstances)
			latestStartTime := getLatestStartTime(startTSmap) * 1e6
			if latestStartTime > 0 {
				newServiceRes.Timestamp = &latestStartTime
			}
		}

		servicesAlertResMsg = append(servicesAlertResMsg, newServiceRes)
	}
	return servicesAlertResMsg, err
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

// 填充来自Clickhouse的告警信息,并填充alertReason
func GetAlertStatusCH(chRepo clickhouse.Repo,
	alertReason *model.AlertReason, alertEventsCountMap *model.AlertEventLevelCountMap,
	alertTypes []string, serviceName string, instances []*model.ServiceInstance, // filter
	startTime, endTime time.Time,
) (alertStatus model.AlertStatusCH) {
	alertStatus = model.AlertStatusCH{
		InfrastructureStatus: model.STATUS_NORMAL,
		NetStatus:            model.STATUS_NORMAL,
		K8sStatus:            model.STATUS_NORMAL,
	}

	if len(alertTypes) == 0 ||
		contains(alertTypes, "infraStatus") ||
		contains(alertTypes, "netStatus") ||
		contains(alertTypes, "appStatus") ||
		contains(alertTypes, "containerStatus") {
		// 查询实例相关的告警信息
		events, _ := chRepo.GetAlertEventsSample(1, startTime, endTime,
			request.AlertFilter{Service: serviceName, Status: "firing"}, instances)

		// 按告警原因修改告警状态/
		for _, event := range events {
			alertGroup := clickhouse.AlertGroup(event.Group)
			switch alertGroup {
			case clickhouse.INFRA_GROUP:
				alertStatus.InfrastructureStatus = model.STATUS_CRITICAL
			case clickhouse.NETWORK_GROUP:
				alertStatus.NetStatus = model.STATUS_CRITICAL
			case clickhouse.APP_GROUP:
				alertStatus.AppStatus = model.STATUS_CRITICAL
			case clickhouse.CONTAINER_GROUP:
				alertStatus.ContainerStatus = model.STATUS_CRITICAL
			default:
				// 忽略 未知 告警
				continue
			}

			if alertReason != nil {
				alertReason.Add(alertGroup.GetAlertType(), model.AlertDetail{
					Timestamp:    event.ReceivedTime.UnixMicro(),
					AlertObject:  event.GetTargetObj(),
					AlertReason:  event.Name,
					AlertMessage: event.Detail,
				})
			}
			if alertEventsCountMap != nil {
				alertEventsCountMap.Add(alertGroup.GetAlertType(), event.Severity)
			}
		}
	}

	if len(alertTypes) == 0 || contains(alertTypes, "k8sStatus") {
		// 查询warning及以上级别的K8s事件
		k8sEvents, _ := chRepo.GetK8sAlertEventsSample(startTime, endTime, instances)
		if len(k8sEvents) > 0 {
			alertStatus.K8sStatus = model.STATUS_CRITICAL
			for _, event := range k8sEvents {
				alertReason.Add(model.K8sEventAlert, model.AlertDetail{
					Timestamp:    event.Timestamp.UnixMicro(),
					AlertObject:  event.GetObjName(),
					AlertReason:  event.GetReason(),
					AlertMessage: event.Body,
				})
			}
		}
	}

	return
}

func getLatestStartTime(startTSmap map[model.ServiceInstance]int64) int64 {
	var latestStartTime int64 = -1
	for _, startTime := range startTSmap {
		if startTime > latestStartTime {
			latestStartTime = startTime
		}
	}
	return latestStartTime
}
