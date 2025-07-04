// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"log"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"

	core "github.com/CloudDetail/apo/backend/pkg/core"
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
func (s *service) GetServicesAlert(
	ctx core.Context,
	groupID int64, clusterIDs []string,
	startTime time.Time, endTime time.Time,
	step time.Duration,
	serviceNames []string,
	returnData []string,
) (res []response.ServiceAlertRes, err error) {
	pqlFilter, err := common.GetPQLFilterByGroupID(ctx, s.dbRepo, "", groupID)
	if err != nil {
		return nil, err
	}
	if len(clusterIDs) > 0 {
		pqlFilter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(clusterIDs...))
	}
	pqlFilter.RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(serviceNames...))

	svcInstances, err := s.promRepo.GetMultiSVCInstanceListByPQLFilter(ctx, startTime.UnixMicro(), endTime.UnixMicro(), pqlFilter)
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

	if len(returnData) == 0 || contains(returnData, "logMetrics") {
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
			_, err = s.AvgLogByPod(ctx, &services[i].Instances, Pods, endTime, duration)
			_, err = s.LogDODByPod(ctx, &services[i].Instances, Pods, endTime, duration)
			_, err = s.LogWOWByPod(ctx, &services[i].Instances, Pods, endTime, duration)
			_, err = s.ServiceLogRangeDataByPod(ctx, &services[i], Pods, startTime, endTime, duration, step)
		}
		for i := range services {
			var ContainerIds []string
			for j := range services[i].Instances {
				if services[i].Instances[j].InstanceType == CONTAINER {
					ContainerIds = append(ContainerIds, services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByContainerId(ctx, &services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.LogDODByContainerId(ctx, &services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.LogWOWByContainerId(ctx, &services[i].Instances, ContainerIds, endTime, duration)
			_, err = s.ServiceLogRangeDataByContainerId(ctx, &services[i], ContainerIds, startTime, endTime, duration, step)
		}
		for i := range services {
			var Pids []string
			for j := range services[i].Instances {
				if services[i].Instances[j].InstanceType == VM {
					Pids = append(Pids, services[i].Instances[j].ConvertName)
				}
			}
			_, err = s.AvgLogByPid(ctx, &services[i].Instances, Pids, endTime, duration)
			_, err = s.LogDODByPid(ctx, &services[i].Instances, Pids, endTime, duration)
			_, err = s.LogWOWByPid(ctx, &services[i].Instances, Pids, endTime, duration)
			_, err = s.ServiceLogRangeDataByPid(ctx, &services[i], Pids, startTime, endTime, duration, step)
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
		normalNowLog := s.getNormalLog(ctx, service, startTime, endTime, 0)
		normalDayLog := s.getNormalLog(ctx, service, startTime, endTime, time.Hour*24)
		normalWeekLog := s.getNormalLog(ctx, service, startTime, endTime, time.Hour*24*7)
		var allLogNow, allLogDay, allLogWeek *float64 // Total log errors for current, yesterday, and last week
		// Integrate instance now,day,week,avg data
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

		// Calculate YoY and populate data
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
			// Yesterday's error was 0, today's error is infinite
			newLogs.Ratio.DayOverDay = maxVal
		} else if allLogNow != nil && *allLogNow == 0 && allLogDay != nil && *allLogDay > 0 {
			// Yesterday's mistake was wrong, today's mistake is 0 negative infinity
			newLogs.Ratio.DayOverDay = minVal
		} else if allLogNow != nil && allLogDay != nil && *allLogNow > 0 && *allLogDay > 0 {
			dod := new(float64)
			*dod = (*allLogNow / *allLogDay - 1) * 100
			newLogs.Ratio.DayOverDay = dod
		}

		if allLogNow != nil && *allLogNow > 0 && allLogWeek != nil && *allLogWeek == 0 {
			// Last week's error was 0, today's error is infinite
			newLogs.Ratio.WeekOverDay = maxVal
		} else if allLogNow != nil && *allLogNow == 0 && allLogWeek != nil && *allLogWeek > 0 {
			// Last week there was a mistake, today's mistake is 0 negative infinity
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

		// fill alarm status
		newServiceRes.AlertStatusCH = GetAlertStatusCH(
			ctx,
			s.chRepo, &newServiceRes.AlertReason, nil,
			returnData, service.ServiceName, serviceInstances,
			startTime, endTime,
		)

		// Fill in last start time
		if returnData == nil || contains(returnData, "lastStartTime") {
			startTSmap, _ := s.promRepo.QueryProcessStartTime(ctx, startTime, endTime, serviceInstances)
			latestStartTime := getLatestStartTime(startTSmap) * 1e6
			if latestStartTime > 0 {
				newServiceRes.Timestamp = &latestStartTime
			}
		}

		servicesAlertResMsg = append(servicesAlertResMsg, newServiceRes)
	}
	return servicesAlertResMsg, err
}

// Fill in the alarm information from the Clickhouse and fill in the alertReason
func GetAlertStatusCH(ctx core.Context, chRepo clickhouse.Repo,
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

		// Query the alarm information related to the instance
		events, _ := chRepo.GetAlertEventsSample(
			ctx,
			1, startTime, endTime,
			request.AlertFilter{Services: []string{serviceName}, Status: "firing"},
			&model.RelatedInstances{
				SIs: instances,
				MIs: []model.MiddlewareInstance{}, // TODO middleware alert status
			},
		)

		// Modify alarm status by alarm reason/
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
				// Ignore unknown alarms
				continue
			}

			alertReason.Add(alertGroup.GetAlertType(), model.AlertDetail{
				Timestamp:    event.ReceivedTime.UnixMicro(),
				AlertObject:  event.GetTargetObj(),
				AlertReason:  event.Name,
				AlertMessage: event.Detail,
			})

			if alertEventsCountMap != nil {
				alertEventsCountMap.Add(alertGroup.GetAlertType(), event.Severity, 1)
			}
		}
	}

	if len(alertTypes) == 0 || contains(alertTypes, "k8sStatus") {
		// Query K8s events of warning level and above
		k8sEvents, _ := chRepo.GetK8sAlertEventsSample(ctx, startTime, endTime, instances)
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

// getNormalLog whether all instances in the service have normal log metrics.
func (s *service) getNormalLog(ctx core.Context, service ServiceDetail, startTime, endTime time.Time, offset time.Duration) []prometheus.MetricResult {
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
	normalLog, err := s.promRepo.QueryData(ctx, endTime.Add(-offset), pql)
	if err != nil {
		log.Println(err)
	}
	return normalLog
}
