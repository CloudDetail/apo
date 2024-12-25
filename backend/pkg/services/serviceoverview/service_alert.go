// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"log"
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
	svcInstances, err := s.promRepo.GetMultiServicesInstanceList(startTime.UnixMicro(), endTime.UnixMicro(), serviceNames)
	if err != nil {
		return nil, err
	}
	var services []ServiceDetail
	for svc, instances := range svcInstances {
		svcDetail := ServiceDetail{
			ServiceName: svc,
			Instances:   make([]Instance, 0),
		}
		for _, instance := range instances.GetInstances() {
			var convertName string
			var instanceType int
			if len(instance.PodName) > 0 {
				convertName = instance.PodName
				instanceType = POD
			} else if len(instance.ContainerId) > 0 {
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
				Namespace:    instance.Namespace,
				Pid:          strconv.FormatInt(instance.Pid, 10),
			}
			svcDetail.Instances = append(svcDetail.Instances, newInstance)
		}
		services = append(services, svcDetail)
	}

	if returnData == nil || contains(returnData, "logMetrics") {
		var duration string
		var stepNS = endTime.Sub(startTime).Nanoseconds()
		duration = strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
		for i := range services {
			var Pods []string
			for j := range services[i].Instances {
				if services[i].Instances[j].InstanceType == POD {
					Pods = append(Pods, services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByPod(&services[i].Instances, Pods, endTime, duration)
			_, err = s.LogDODByPod(&services[i].Instances, Pods, endTime, duration)
			_, err = s.LogWOWByPod(&services[i].Instances, Pods, endTime, duration)
			_, err = s.ServiceLogRangeDataByPod(&services[i], Pods, startTime, endTime, duration, step)
		}
		for i := range services {
			var ContainerIds []string
			for j := range services[i].Instances {
				if services[i].Instances[j].InstanceType == CONTAINER {
					ContainerIds = append(ContainerIds, services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByContainerId(&services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.LogDODByContainerId(&services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.LogWOWByContainerId(&services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.ServiceLogRangeDataByContainerId(&services[i], ContainerIds, startTime, endTime, duration, step)
		}
		for i := range services {
			var Pids []string
			for j := range services[i].Instances {
				if services[i].Instances[j].InstanceType == VM {
					Pids = append(Pids, services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByPid(&services[i].Instances, Pids, endTime, duration)
			_, err = s.LogDODByPid(&services[i].Instances, Pids, endTime, duration)
			_, err = s.LogWOWByPid(&services[i].Instances, Pids, endTime, duration)
			_, err = s.ServiceLogRangeDataByPid(&services[i], Pids, startTime, endTime, duration, step)
		}
	}

	var servicesAlertResMsg []response.ServiceAlertRes
	for _, service := range services {
		if service.ServiceName == "" {
			continue
		}
		newLogs := response.TempChartObject{}
		if service.LogData != nil {
			data := make(map[int64]float64)
			for _, item := range service.LogData {
				timestamp := item.TimeStamp
				value := item.Value
				data[timestamp] = value
			}
			newLogs.ChartData = data
		} else {
			values := make(map[int64]float64)
			for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
				values[ts] = 0
			}
			newLogs.ChartData = values
		}
		normalNowLog := s.getNormalLog(service, startTime, endTime, 0)
		normalDayLog := s.getNormalLog(service, startTime, endTime, time.Hour*24)
		normalWeekLog := s.getNormalLog(service, startTime, endTime, time.Hour*24*7)
		var allLogNow, allLogDay, allLogWeek *float64 // 当前、昨天、上周的总日志错误数
		// 整合instance的now，day，week，avg数据
		for _, instance := range service.Instances {
			if instance.LogNow != nil {
				if allLogNow == nil {
					allLogNow = new(float64)
				}
				*allLogNow += *instance.LogNow
			}

			if instance.LogYesterday != nil {
				if allLogDay == nil {
					allLogDay = new(float64)
				}
				*allLogDay += *instance.LogYesterday
			}

			if instance.LogLastWeek != nil {
				if allLogWeek == nil {
					allLogWeek = new(float64)
				}
				*allLogWeek += *instance.LogLastWeek
			}

			if instance.AvgLog != nil {
				if newLogs.Value == nil {
					newLogs.Value = new(float64)
				}
				*newLogs.Value += *instance.AvgLog
			}
		}

		// 计算同比并填充数据
		if allLogDay == nil && normalDayLog != nil {
			allLogDay = new(float64)
		}
		if allLogWeek == nil && normalWeekLog != nil {
			allLogWeek = new(float64)
		}
		if allLogNow == nil && normalNowLog != nil {
			allLogNow = new(float64)
		}
		maxVal := new(float64)
		*maxVal = prometheus.RES_MAX_VALUE
		minVal := new(float64)
		*minVal = -100
		if allLogNow != nil && *allLogNow > 0 && allLogDay != nil && *allLogDay == 0 {
			// 昨天错误为0，今天有错误 正无穷
			newLogs.Ratio.DayOverDay = maxVal
		} else if allLogNow != nil && *allLogNow == 0 && allLogDay != nil && *allLogDay > 0 {
			// 昨天错有错误，今天错误为0 负无穷
			newLogs.Ratio.DayOverDay = minVal
		} else if allLogNow != nil && allLogDay != nil && *allLogNow > 0 && *allLogDay > 0 {
			dod := new(float64)
			*dod = (*allLogNow / *allLogDay - 1) * 100
			newLogs.Ratio.DayOverDay = dod
		}

		if allLogNow != nil && *allLogNow > 0 && allLogWeek != nil && *allLogWeek == 0 {
			// 上周错误为0，今天有错误 正无穷
			newLogs.Ratio.WeekOverDay = maxVal
		} else if allLogNow != nil && *allLogNow == 0 && allLogWeek != nil && *allLogWeek > 0 {
			// 上周错有错误，今天错误为0 负无穷
			newLogs.Ratio.WeekOverDay = minVal
		} else if allLogNow != nil && allLogWeek != nil && *allLogNow > 0 && *allLogWeek > 0 {
			wow := new(float64)
			*wow = (*allLogNow / *allLogWeek - 1) * 100
			newLogs.Ratio.WeekOverDay = wow
		}

		if newLogs.Value == nil && normalNowLog != nil {
			newLogs.Value = new(float64)
		}

		newServiceRes := response.ServiceAlertRes{
			ServiceName: service.ServiceName,
			Logs:        newLogs,
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
				alertEventsCountMap.Add(alertGroup.GetAlertType(), event.Severity, 1)
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

// getNormalLog 查询service下所有实例是否有正常log指标
func (s *service) getNormalLog(service ServiceDetail, startTime, endTime time.Time, offset time.Duration) []prometheus.MetricResult {
	startTS, endTS := startTime.UnixMicro(), endTime.UnixMicro()
	var pods, pids, nodeNames []string
	for _, instance := range service.Instances {
		if len(instance.Pod) > 0 {
			pods = append(pods, instance.Pod)
		} else if len(instance.Pid) > 0 && len(instance.NodeName) > 0 {
			pids = append(pids, instance.NodeName)
			nodeNames = append(nodeNames, instance.NodeName)
		}
	}

	podFilter := make([]string, 2)
	podFilter[0] = prometheus.LogMetricPodRegexPQLFilter
	podFilter[1] = prometheus.RegexMultipleValue(pods...)
	vmFilter := make([]string, 4)
	vmFilter[0] = prometheus.LogMetricNodeRegexPQLFilter
	vmFilter[1] = prometheus.RegexMultipleValue(nodeNames...)
	vmFilter[2] = prometheus.LogMetricPidRegexPQLFilter
	vmFilter[3] = prometheus.RegexMultipleValue(pids...)
	pql, err := prometheus.PQLInstanceLog(
		prometheus.PQLNormalLogCountWithFilters,
		startTS, endTS,
		prometheus.LogGranularity,
		podFilter, vmFilter)
	if err != nil {
		return nil
	}
	normalLog, err := s.promRepo.QueryData(endTime.Add(-offset), pql)
	if err != nil {
		log.Println(err)
	}
	return normalLog
}
