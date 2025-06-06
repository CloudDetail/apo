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
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

func (s *service) GetInstancesNew(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error) {
	threshold, err := s.dbRepo.GetOrCreateThreshold(ctx, "", "", database.GLOBAL)
	if err != nil {
		return res, err
	}
	errorThreshold := threshold.ErrorRate
	tpsThreshold := threshold.Tps
	latencyThreshold := threshold.Latency

	filter := InstancesFilter{SrvName: serviceName, ContentKey: endPoint}
	filters := filter.ExtractFilterStr()
	// Get instance
	instanceList, err := s.promRepo.GetInstanceList(ctx, startTime.UnixMicro(), endTime.UnixMicro(), serviceName, endPoint)
	if err != nil {
		return res, err
	}
	// Fill the instance
	var instances = &InstanceMap{
		MetricGroupList: []*prometheus.InstanceMetrics{},
		MetricGroupMap:  map[prometheus.InstanceKey]*prometheus.InstanceMetrics{},
	}
	for _, instance := range instanceList.InstanceMap {
		key := prometheus.InstanceKey{
			ServiceName: instance.ServiceName,
			PID:         strconv.FormatInt(instance.Pid, 10),
			ContainerId: instance.ContainerId,
			Pod:         instance.PodName,
			Namespace:   instance.Namespace,
			NodeName:    instance.NodeName,
			NodeIP:      instance.NodeIP,
		}
		metric := &prometheus.InstanceMetrics{
			InstanceKey: key,
		}
		// deduplication, pod type instance will add containerID,PID,pod three but finally point to the same instance
		if _, ok := instances.MetricGroupMap[key]; !ok {
			instances.MetricGroupMap[key] = metric
			instances.MetricGroupList = append(instances.MetricGroupList, metric)
		}
	}
	// Fill RED metric
	s.InstanceRED(ctx, startTime, endTime, filters, instances)

	// populate chart data
	chartErr := s.InstanceRangeData(ctx, instances, startTime, endTime, step, filters)
	if chartErr.ErrorOrNil() != nil {
		log.Println("get instance range data error: ", chartErr)
	}
	// Fill log data
	logErr := s.InstanceLog(ctx, instances, startTime, endTime, step)
	if logErr.ErrorOrNil() != nil {
		log.Println("get instance log data error: ", logErr)
	}
	resData := []response.InstanceData{}
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
	return
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
