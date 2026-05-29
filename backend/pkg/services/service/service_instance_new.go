// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"log"
	"math"
	"strconv"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

func (s *service) GetInstancesNew(ctx core.Context, req *request.GetServiceInstanceRequest) (res response.InstancesRes, err error) {
	threshold, err := s.dbRepo.GetOrCreateThreshold(ctx, "", "", database.GLOBAL)
	if err != nil {
		return res, err
	}
	errorThreshold := threshold.ErrorRate
	tpsThreshold := threshold.Tps
	latencyThreshold := threshold.Latency

	var svcInstances *model.ServiceInstances
	instancesList, err := s.chRepo.GetDataplaneServiceInstances(ctx, req.StartTime, req.EndTime, "", []string{req.ServiceName})
	if err != nil {
		log.Printf("failed to query serviceInstance from Dataplane, services %s, err : %v", req.ServiceName, err)
	}

	filter, err := common.GetPQLFilterByGroupID(ctx, s.dbRepo, "", req.GroupID)
	if err != nil {
		return res, err
	}
	if len(req.ServiceName) > 0 {
		filter.Equal(prometheus.ServiceNameKey, req.ServiceName)
	}
	if len(req.Endpoint) > 0 {
		filter.Equal(prometheus.ContentKeyKey, req.Endpoint)
	}
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(req.ClusterIDs...))
	}

	if len(instancesList) > 0 {
		svcInstances = model.NewServiceInstances()
		svcInstances.AddInstances(instancesList)
	} else {

		// Get instance
		svcInstances, err = s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
		if err != nil {
			return res, err
		}
	}

	// Fill the instance
	var instances = &InstanceMap{
		MetricGroupList: []*prometheus.InstanceMetrics{},
		MetricGroupMap:  map[prometheus.InstanceKey]*prometheus.InstanceMetrics{},
	}
	seededInstanceKeys := make(map[prometheus.InstanceKey]struct{})
	for _, instance := range svcInstances.InstanceMap {
		key := prometheus.InstanceKey{
			ServiceName: instance.ServiceName,
			PID:         strconv.FormatInt(instance.Pid, 10),
			ContainerId: instance.ContainerId,
			Pod:         instance.PodName,
			Namespace:   instance.Namespace,
			NodeName:    instance.NodeName,
			NodeIP:      instance.NodeIP,
			ClusterID:   instance.ClusterID,
		}
		metric := &prometheus.InstanceMetrics{
			InstanceKey: key,
		}
		// deduplication, pod type instance will add containerID,PID,pod three but finally point to the same instance
		if _, ok := instances.MetricGroupMap[key]; !ok {
			instances.MetricGroupMap[key] = metric
			instances.MetricGroupList = append(instances.MetricGroupList, metric)
			seededInstanceKeys[key] = struct{}{}
		}
	}

	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	step := time.Duration(req.Step * 1000)
	serviceName := req.ServiceName

	// Fill RED metric
	// TODO support query InstanceRED from Dataplane
	s.InstanceRED(ctx, startTime, endTime, filter, instances)

	// populate chart data
	chartErr := s.InstanceRangeData(ctx, instances, startTime, endTime, step, filter)
	if chartErr.ErrorOrNil() != nil {
		log.Println("get instance range data error: ", chartErr)
	}
	// Fill log data
	// TODO support query instance logMetric from Dataplane
	logErr := s.InstanceLog(ctx, instances, startTime, endTime, step)
	if logErr.ErrorOrNil() != nil {
		log.Println("get instance log data error: ", logErr)
	}
	logDuplicateInstanceKeys(req, instances, seededInstanceKeys)
	resData := []response.InstanceData{}
	resDataByName := make(map[string]int)
	res.Status = model.STATUS_NORMAL
	// Status of filled instance and RED metrics
	for _, instance := range instances.MetricGroupList {
		if instance.REDMetrics.DOD.ErrorRate != nil && *instance.REDMetrics.DOD.ErrorRate > errorThreshold {
			res.Status = model.STATUS_CRITICAL
		}
		if instance.REDMetrics.DOD.Latency != nil && *instance.REDMetrics.DOD.Latency > latencyThreshold {
			res.Status = model.STATUS_CRITICAL
		}
		if instance.REDMetrics.DOD.TPM != nil && *instance.REDMetrics.DOD.TPM > tpsThreshold {
			res.Status = model.STATUS_CRITICAL
		}

		if instance.REDMetrics.WOW.ErrorRate != nil && *instance.REDMetrics.WOW.ErrorRate > errorThreshold {
			res.Status = model.STATUS_CRITICAL
		}
		if instance.REDMetrics.WOW.Latency != nil && *instance.REDMetrics.WOW.Latency > latencyThreshold {
			res.Status = model.STATUS_CRITICAL
		}
		if instance.REDMetrics.WOW.TPM != nil && *instance.REDMetrics.WOW.TPM > tpsThreshold {
			res.Status = model.STATUS_CRITICAL
		}
	}

	for _, instance := range instances.MetricGroupList {
		if IsInvalidData(instances.MetricGroupMap, instance) {
			continue
		}

		// Construct the return value of the latency
		latencyTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  instance.REDMetrics.DOD.Latency,
				WeekOverDay: instance.REDMetrics.WOW.Latency,
			},
			Value: instance.REDMetrics.Avg.Latency,
		}
		if instance.LatencyData != nil {
			latencyTempChartObject.ChartData = DataToChart(instance.LatencyData)
		} else {
			latencyTempChartObject.ChartData = FillChart(startTime, endTime, step)
		}

		// Construct the return value of error
		errorTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  instance.REDMetrics.DOD.ErrorRate,
				WeekOverDay: instance.REDMetrics.WOW.ErrorRate,
			},
			Value: instance.REDMetrics.Avg.ErrorRate,
		}
		if errorTempChartObject.Value == nil {
			zero := new(float64)
			errorTempChartObject.Value = zero
		}
		if instance.ErrorRateData != nil {
			errorTempChartObject.ChartData = DataToChart(instance.ErrorRateData)
		} else {
			errorTempChartObject.ChartData = FillChart(startTime, endTime, step)
		}

		// construct tps return value
		tpsTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  instance.REDMetrics.DOD.TPM,
				WeekOverDay: instance.REDMetrics.WOW.TPM,
			},
			Value: instance.REDMetrics.Avg.TPM,
		}
		if instance.TPMData != nil {
			tpsTempChartObject.ChartData = DataToChart(instance.TPMData)
		} else {
			tpsTempChartObject.ChartData = FillChart(startTime, endTime, step)
		}

		// Construct log return value
		logTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  instance.LogDayOverDay,
				WeekOverDay: instance.LogWeekOverWeek,
			},
			Value: instance.LogAVGData,
		}
		filters := ExtractLogFilter(instance.InstanceKey)
		var normalNowLog, normalDayLog, normalWeekLog []prometheus.MetricResult
		if len(filters) > 0 {
			normalNowLog = s.GetNormalLog(ctx, startTime, endTime, filters, 0)
			normalDayLog = s.GetNormalLog(ctx, startTime, endTime, filters, time.Hour*24)
			normalWeekLog = s.GetNormalLog(ctx, startTime, endTime, filters, time.Hour*24*7)
		}

		// normal data, log error average should be filled with 0
		if logTempChartObject.Value == nil && normalNowLog != nil {
			zero := new(float64)
			logTempChartObject.Value = zero
		}

		if logTempChartObject.Ratio.DayOverDay != nil && *logTempChartObject.Ratio.DayOverDay == prometheus.RES_MAX_VALUE && normalDayLog == nil {
			// There was no normal data yesterday, which should be nil year on year.
			logTempChartObject.Ratio.DayOverDay = nil
		} else if logTempChartObject.Ratio.DayOverDay != nil && logTempChartObject.Ratio.DayOverDay == nil && normalDayLog != nil && normalNowLog != nil {
			// normal data, year-on-year should be 0
			zero := new(float64)
			logTempChartObject.Ratio.DayOverDay = zero
		}

		if logTempChartObject.Ratio.WeekOverDay != nil && *logTempChartObject.Ratio.WeekOverDay == prometheus.RES_MAX_VALUE && normalWeekLog == nil {
			// No normal data last week, year-on-year should be nil
			logTempChartObject.Ratio.WeekOverDay = nil
		} else if logTempChartObject.Ratio.WeekOverDay != nil && logTempChartObject.Ratio.WeekOverDay == nil && normalWeekLog != nil && normalNowLog != nil {
			// normal data, year-on-year should be 0
			zero := new(float64)
			logTempChartObject.Ratio.WeekOverDay = zero
		}

		if instance.LogData != nil {
			logTempChartObject.ChartData = DataToChart(instance.LogData)
		} else {
			logTempChartObject.ChartData = FillChart(startTime, endTime, step)
		}

		newInstance := response.InstanceData{
			Name:        instance.InstanceKey.GenInstanceName(),
			Namespace:   instance.Namespace,
			NodeName:    instance.NodeName,
			NodeIP:      instance.NodeIP,
			Timestamp:   nil,
			Latency:     latencyTempChartObject,
			Tps:         tpsTempChartObject,
			ErrorRate:   errorTempChartObject,
			Logs:        logTempChartObject,
			AlertStatus: model.NORMAL_ALERT_STATUS,
			AlertReason: model.AlertReason{},
		}

		pidI64, err := strconv.ParseInt(instance.PID, 10, 64)
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
				ClusterID:   instance.ClusterID,
			},
		}
		// fill alarm status
		newInstance.AlertStatusCH = serviceoverview.GetAlertStatusCH(
			ctx,
			s.chRepo, &newInstance.AlertReason, nil,
			nil, serviceName, instanceSingleList,
			startTime, endTime,
		)

		// Fill in last start time
		startTSmap, _ := s.promRepo.QueryProcessStartTime(ctx, startTime, endTime, instanceSingleList)
		latestStartTime := getLatestStartTime(startTSmap) * 1e6
		if latestStartTime > 0 {
			newInstance.Timestamp = &latestStartTime
		}

		resData = appendInstanceDataByName(resData, resDataByName, newInstance)
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
	return
}

func appendInstanceDataByName(list []response.InstanceData, indexByName map[string]int, item response.InstanceData) []response.InstanceData {
	if item.Name == "" {
		return append(list, item)
	}
	index, exists := indexByName[item.Name]
	if !exists {
		indexByName[item.Name] = len(list)
		return append(list, item)
	}
	if instanceDataScore(item) > instanceDataScore(list[index]) {
		list[index] = item
	}
	return list
}

func instanceDataScore(item response.InstanceData) int {
	score := 0
	score += chartObjectScore(item.Latency) * 4
	score += chartObjectScore(item.Tps) * 4
	score += chartObjectScore(item.Logs) * 2
	score += chartObjectScore(item.ErrorRate)
	if item.Timestamp != nil {
		score += 5
	}
	if item.NodeIP != "" {
		score++
	}
	return score
}

func chartObjectScore(chart response.TempChartObject) int {
	score := 0
	if chart.Value != nil {
		score += 10
		if *chart.Value != 0 {
			score += 10
		}
	}
	for _, value := range chart.ChartData {
		if value != 0 {
			score += 20
			break
		}
	}
	if chart.Ratio.DayOverDay != nil {
		score += 2
	}
	if chart.Ratio.WeekOverDay != nil {
		score += 2
	}
	return score
}

func logDuplicateInstanceKeys(req *request.GetServiceInstanceRequest, instances *InstanceMap, seededKeys map[prometheus.InstanceKey]struct{}) {
	if instances == nil {
		return
	}
	metricsByDisplayName := make(map[string][]*prometheus.InstanceMetrics)
	for _, metric := range instances.MetricGroupList {
		if metric == nil {
			continue
		}
		displayName := metric.InstanceKey.GenInstanceName()
		if displayName == "" {
			continue
		}
		metricsByDisplayName[displayName] = append(metricsByDisplayName[displayName], metric)
	}

	for displayName, duplicateMetrics := range metricsByDisplayName {
		if len(duplicateMetrics) <= 1 {
			continue
		}
		log.Printf(
			"duplicate service instance display name detected: displayName=%q service=%q endpoint=%q startTime=%d endTime=%d count=%d",
			displayName,
			req.ServiceName,
			req.Endpoint,
			req.StartTime,
			req.EndTime,
			len(duplicateMetrics),
		)
		for idx, metric := range duplicateMetrics {
			source := "prometheus_metric"
			if _, ok := seededKeys[metric.InstanceKey]; ok {
				source = "dataplane_instance"
			}
			log.Printf(
				"duplicate service instance key detail: displayName=%q index=%d source=%s service=%q pid=%q containerId=%q pod=%q namespace=%q nodeName=%q nodeIP=%q clusterID=%q hasLatency=%t hasTPS=%t hasErrorRate=%t hasLog=%t",
				displayName,
				idx,
				source,
				metric.ServiceName,
				metric.PID,
				metric.ContainerId,
				metric.Pod,
				metric.Namespace,
				metric.NodeName,
				metric.NodeIP,
				metric.ClusterID,
				metric.REDMetrics.Avg.Latency != nil || hasNonZeroPoints(metric.LatencyData),
				metric.REDMetrics.Avg.TPM != nil || hasNonZeroPoints(metric.TPMData),
				metric.REDMetrics.Avg.ErrorRate != nil || hasNonZeroPoints(metric.ErrorRateData),
				metric.LogAVGData != nil || hasNonZeroPoints(metric.LogData),
			)
		}
	}
}

func hasNonZeroPoints(points []prometheus.Points) bool {
	for _, point := range points {
		if point.Value != 0 {
			return true
		}
	}
	return false
}

// DataToChart convert chart data to map
func DataToChart(data []prometheus.Points) map[int64]float64 {
	chart := make(map[int64]float64)
	for _, item := range data {
		timestamp := item.TimeStamp
		value := item.Value
		if !math.IsInf(value, 1) {
			chart[timestamp] = value
		} else {
			chart[timestamp] = prometheus.RES_MAX_VALUE
		}
	}
	return chart
}

func FillChart(startTime, endTime time.Time, step time.Duration) map[int64]float64 {
	values := make(map[int64]float64)
	for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += step.Microseconds() {
		values[ts] = 0
	}
	return values
}

func ExtractLogFilter(instance prometheus.InstanceKey) []string {
	var filters []string
	if len(instance.Pod) > 0 {
		filters = make([]string, 2)
		filters[0] = prometheus.LogMetricPodRegexPQLFilter
		filters[1] = instance.Pod
	} else if len(instance.PID) > 0 && len(instance.NodeName) > 0 {
		filters = make([]string, 4)
		filters[0] = prometheus.LogMetricPidRegexPQLFilter
		filters[1] = instance.PID
		filters[2] = prometheus.LogMetricNodeRegexPQLFilter
		filters[3] = instance.NodeName
	}
	return filters
}
