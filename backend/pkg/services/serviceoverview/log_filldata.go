// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) AvgLogByPod(ctx core.Context, Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	var LogRateRes []prometheus.MetricResult
	queryLog := prometheus.QueryLogPromql(duration, prometheus.AvgLog, pods)
	LogRateRes, err := s.promRepo.QueryData(ctx, endTime, queryLog)
	for _, result := range LogRateRes {
		podName := result.Metric.POD
		if podName == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == podName {
				(*Instances)[i].AvgLog = &value
				break
			}
		}
	}
	return Instances, err
}

func (s *service) LogDODByPod(ctx core.Context, Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogNow := prometheus.QueryLogPromql(duration, prometheus.LogNow, pods)
	logNow, err := s.promRepo.QueryData(ctx, endTime, queryLogNow)
	if err != nil {
		return Instances, err
	}
	queryLogYesterday := prometheus.QueryLogPromql(duration, prometheus.LogYesterday, pods)
	logYesterday, err := s.promRepo.QueryData(ctx, endTime, queryLogYesterday)
	if err != nil {
		return Instances, err
	}
	for _, result := range logNow {
		podName := result.Metric.POD
		if podName == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == podName {
				(*Instances)[i].LogNow = &value
				break
			}
		}
	}

	for _, result := range logYesterday {
		podName := result.Metric.POD
		if podName == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == podName {
				if Instance.LogYesterday == nil {
					(*Instances)[i].LogYesterday = new(float64)
				}
				*(*Instances)[i].LogYesterday += value
				break
			}
		}
	}
	return Instances, err
}

func (s *service) LogWOWByPod(ctx core.Context, Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogLastWeek := prometheus.QueryLogPromql(duration, prometheus.LogLastWeek, pods)
	logLastWeek, err := s.promRepo.QueryData(ctx, endTime, queryLogLastWeek)
	if err != nil {
		return Instances, err
	}
	for _, result := range logLastWeek {
		podName := result.Metric.POD
		if podName == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == podName {
				if Instance.LogLastWeek == nil {
					(*Instances)[i].LogLastWeek = new(float64)
				}
				(*Instances)[i].LogLastWeek = &value
				break
			}
		}
	}
	return Instances, err
}

// Query the graph
func (s *service) LogRangeDataByPod(ctx core.Context, Instances *[]Instance, pods []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	//LogDataRes, err = s.promRepo.QueryRangePrometheusErrorLast30min(searchTime)
	LogDataQuery := prometheus.QueryLogPromql(stepToStr, prometheus.AvgLog, pods)
	LogDataRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LogDataQuery, step)
	for _, result := range LogDataRes {
		podName := result.Metric.POD
		if podName == "" {
			continue
		}
		for i, instance := range *Instances {
			if instance.ConvertName == podName {
				(*Instances)[i].LogData = result.Values
				break
			}
		}
	}

	return Instances, err
}

func (s *service) AvgLogByContainerId(ctx core.Context, Instances *[]Instance, containerIds []string, endTime time.Time, duration string) (*[]Instance, error) {
	var LogRateRes []prometheus.MetricResult
	queryLog := prometheus.QueryLogByContainerIdPromql(duration, prometheus.AvgLog, containerIds)
	LogRateRes, err := s.promRepo.QueryData(ctx, endTime, queryLog)
	for _, result := range LogRateRes {
		containerId := result.Metric.ContainerID
		if containerId == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == containerId {
				(*Instances)[i].AvgLog = &value
				break
			}
		}
	}
	return Instances, err
}

func (s *service) LogDODByContainerId(ctx core.Context, Instances *[]Instance, containerIds []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogNow := prometheus.QueryLogByContainerIdPromql(duration, prometheus.LogNow, containerIds)
	logNow, err := s.promRepo.QueryData(ctx, endTime, queryLogNow)
	if err != nil {
		return Instances, err
	}
	queryLogYesterday := prometheus.QueryLogByContainerIdPromql(duration, prometheus.LogYesterday, containerIds)
	logYesterday, err := s.promRepo.QueryData(ctx, endTime, queryLogYesterday)
	if err != nil {
		return Instances, err
	}
	for _, result := range logNow {
		containerId := result.Metric.ContainerID
		if containerId == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == containerId {
				if Instance.LogNow == nil {
					(*Instances)[i].LogNow = new(float64)
				}
				*(*Instances)[i].LogNow += value
				break
			}
		}
	}

	for _, result := range logYesterday {
		containerId := result.Metric.ContainerID
		if containerId == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == containerId {
				if Instance.LogYesterday == nil {
					(*Instances)[i].LogYesterday = new(float64)
				}
				*(*Instances)[i].LogYesterday += value
				break
			}
		}
	}
	return Instances, err
}

func (s *service) LogWOWByContainerId(ctx core.Context, Instances *[]Instance, containerIds []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLastWeek := prometheus.QueryLogByContainerIdPromql(duration, prometheus.LogLastWeek, containerIds)
	logLastWeek, err := s.promRepo.QueryData(ctx, endTime, queryLastWeek)
	if err != nil {
		return Instances, err
	}

	for _, result := range logLastWeek {
		containerId := result.Metric.ContainerID
		if containerId == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == containerId {
				(*Instances)[i].LogLastWeek = &value
				break
			}
		}
	}
	return Instances, err
}

// Query the graph

func (s *service) LogRangeDataByContainerId(ctx core.Context, Instances *[]Instance, containerIds []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	//LogDataRes, err = s.promRepo.QueryRangePrometheusErrorLast30min(searchTime)
	LogDataQuery := prometheus.QueryLogByContainerIdPromql(stepToStr, prometheus.AvgLog, containerIds)
	LogDataRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LogDataQuery, step)
	for _, result := range LogDataRes {
		containerId := result.Metric.ContainerID
		if containerId == "" {
			continue
		}
		for i, instance := range *Instances {
			if instance.ConvertName == containerId {
				(*Instances)[i].LogData = result.Values
				break
			}
		}
	}

	return Instances, err
}

func (s *service) AvgLogByPid(ctx core.Context, Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLog := prometheus.QueryLogByPidPromql(duration, prometheus.AvgLog, pods)
	LogRateRes, err := s.promRepo.QueryData(ctx, endTime, queryLog)
	for _, result := range LogRateRes {
		pid := result.Metric.PID
		if pid == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == pid {
				(*Instances)[i].AvgLog = &value
				break
			}
		}
	}
	return Instances, err
}

func (s *service) LogDODByPid(ctx core.Context, Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogNow := prometheus.QueryLogByPidPromql(duration, prometheus.LogNow, pids)
	logNow, err := s.promRepo.QueryData(ctx, endTime, queryLogNow)
	if err != nil {
		return Instances, err
	}
	queryLogYesterday := prometheus.QueryLogByPidPromql(duration, prometheus.LogYesterday, pids)
	logYesterday, err := s.promRepo.QueryData(ctx, endTime, queryLogYesterday)
	if err != nil {
		return Instances, err
	}
	for _, result := range logNow {
		pid := result.Metric.PID
		if pid == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == pid {
				if Instance.LogNow == nil {
					(*Instances)[i].LogNow = new(float64)
				}
				*(*Instances)[i].LogNow += value
				break
			}
		}
	}

	for _, result := range logYesterday {
		pid := result.Metric.PID
		if pid == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == pid {
				if Instance.LogYesterday == nil {
					(*Instances)[i].LogYesterday = new(float64)
				}
				*(*Instances)[i].LogYesterday += value
				break
			}
		}
	}
	return Instances, err
}
func (s *service) LogWOWByPid(ctx core.Context, Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogLastWeek := prometheus.QueryLogByPidPromql(duration, prometheus.LogLastWeek, pids)
	logLastWeek, err := s.promRepo.QueryData(ctx, endTime, queryLogLastWeek)
	if err != nil {
		return Instances, err
	}

	for _, result := range logLastWeek {
		pid := result.Metric.PID
		if pid == "" {
			continue
		}
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ConvertName == pid {
				if Instance.LogLastWeek == nil {
					(*Instances)[i].LogLastWeek = new(float64)
				}
				*(*Instances)[i].LogLastWeek += value
				break
			}
		}
	}
	return Instances, err
}

// Query the graph
func (s *service) LogRangeDataByPid(ctx core.Context, Instances *[]Instance, pods []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	//LogDataRes, err = s.promRepo.QueryRangePrometheusErrorLast30min(searchTime)
	LogDataQuery := prometheus.QueryLogByPidPromql(stepToStr, prometheus.AvgLog, pods)
	LogDataRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LogDataQuery, step)
	for _, result := range LogDataRes {
		pid := result.Metric.PID
		if pid == "" {
			continue
		}
		for i, instance := range *Instances {
			if instance.ConvertName == pid {
				(*Instances)[i].LogData = result.Values
				break
			}
		}
	}

	return Instances, err
}

func (s *service) latestSourcesForVMInstances(ctx core.Context, instances *[]Instance) map[int]string {
	result := make(map[int]string)
	if !config.Get().Promethues.LogparserPIDSourceQueryEnabled || instances == nil {
		return result
	}

	pairs := make([]clickhouse.NodePID, 0)
	indexByPair := make(map[clickhouse.NodePID][]int)
	for i, instance := range *instances {
		if instance.InstanceType != VM || instance.NodeName == "" || instance.Pid == "" {
			continue
		}
		pid, err := strconv.ParseInt(instance.Pid, 10, 64)
		if err != nil || pid <= 0 {
			continue
		}
		pair := clickhouse.NodePID{Node: instance.NodeName, PID: pid}
		if _, ok := indexByPair[pair]; !ok {
			pairs = append(pairs, pair)
		}
		indexByPair[pair] = append(indexByPair[pair], i)
	}

	sources, err := s.chRepo.GetLatestSourceByNodePIDs(ctx, pairs)
	if err != nil {
		return result
	}
	for pair, source := range sources {
		for _, index := range indexByPair[pair] {
			result[index] = source
		}
	}
	return result
}

func (s *service) queryVMLogPointBySource(ctx core.Context, instance Instance, source string, endTime time.Time, duration string, queryType prometheus.QueryType) (*float64, error) {
	query := prometheus.QueryLogBySourcePromql(duration, queryType, instance.NodeName, source)
	results, err := s.promRepo.QueryData(ctx, endTime, query)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 || len(results[0].Values) == 0 {
		return nil, nil
	}
	value := results[0].Values[0].Value
	return &value, nil
}

func (s *service) AvgLogByPidSource(ctx core.Context, Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	sources := s.latestSourcesForVMInstances(ctx, Instances)
	fallbackPids := make([]string, 0)
	for i, instance := range *Instances {
		if instance.InstanceType != VM {
			continue
		}
		source := sources[i]
		if source == "" {
			fallbackPids = append(fallbackPids, instance.Pid)
			continue
		}
		value, err := s.queryVMLogPointBySource(ctx, instance, source, endTime, duration, prometheus.AvgLog)
		if err != nil {
			return Instances, err
		}
		if value != nil {
			(*Instances)[i].AvgLog = value
		}
	}
	if len(fallbackPids) > 0 {
		return s.AvgLogByPid(ctx, Instances, fallbackPids, endTime, duration)
	}
	return Instances, nil
}

func (s *service) LogDODByPidSource(ctx core.Context, Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	sources := s.latestSourcesForVMInstances(ctx, Instances)
	fallbackPids := make([]string, 0)
	for i, instance := range *Instances {
		if instance.InstanceType != VM {
			continue
		}
		source := sources[i]
		if source == "" {
			fallbackPids = append(fallbackPids, instance.Pid)
			continue
		}
		now, err := s.queryVMLogPointBySource(ctx, instance, source, endTime, duration, prometheus.LogNow)
		if err != nil {
			return Instances, err
		}
		yesterday, err := s.queryVMLogPointBySource(ctx, instance, source, endTime, duration, prometheus.LogYesterday)
		if err != nil {
			return Instances, err
		}
		if now != nil {
			if (*Instances)[i].LogNow == nil {
				(*Instances)[i].LogNow = new(float64)
			}
			*(*Instances)[i].LogNow += *now
		}
		if yesterday != nil {
			if (*Instances)[i].LogYesterday == nil {
				(*Instances)[i].LogYesterday = new(float64)
			}
			*(*Instances)[i].LogYesterday += *yesterday
		}
	}
	if len(fallbackPids) > 0 {
		return s.LogDODByPid(ctx, Instances, fallbackPids, endTime, duration)
	}
	return Instances, nil
}

func (s *service) LogWOWByPidSource(ctx core.Context, Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	sources := s.latestSourcesForVMInstances(ctx, Instances)
	fallbackPids := make([]string, 0)
	for i, instance := range *Instances {
		if instance.InstanceType != VM {
			continue
		}
		source := sources[i]
		if source == "" {
			fallbackPids = append(fallbackPids, instance.Pid)
			continue
		}
		lastWeek, err := s.queryVMLogPointBySource(ctx, instance, source, endTime, duration, prometheus.LogLastWeek)
		if err != nil {
			return Instances, err
		}
		if lastWeek != nil {
			if (*Instances)[i].LogLastWeek == nil {
				(*Instances)[i].LogLastWeek = new(float64)
			}
			*(*Instances)[i].LogLastWeek += *lastWeek
		}
	}
	if len(fallbackPids) > 0 {
		return s.LogWOWByPid(ctx, Instances, fallbackPids, endTime, duration)
	}
	return Instances, nil
}

func (s *service) LogRangeDataByPidSource(ctx core.Context, Instances *[]Instance, pids []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	stepToStr := logStepToString(step)
	sources := s.latestSourcesForVMInstances(ctx, Instances)
	fallbackPids := make([]string, 0)
	for i, instance := range *Instances {
		if instance.InstanceType != VM {
			continue
		}
		source := sources[i]
		if source == "" {
			fallbackPids = append(fallbackPids, instance.Pid)
			continue
		}
		query := prometheus.QueryLogBySourcePromql(stepToStr, prometheus.AvgLog, instance.NodeName, source)
		results, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, query, step)
		if err != nil {
			return Instances, err
		}
		if len(results) > 0 {
			(*Instances)[i].LogData = results[0].Values
		}
	}
	if len(fallbackPids) > 0 {
		return s.LogRangeDataByPid(ctx, Instances, fallbackPids, startTime, endTime, duration, step)
	}
	return Instances, nil
}

func (s *service) ServiceLogRangeDataByPidSource(ctx core.Context, Service *ServiceDetail, pids []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
	if Service == nil {
		return nil, errors.New("service is nil")
	}
	if pids == nil {
		return Service, nil
	}
	stepToStr := logStepToString(step)
	sources := s.latestSourcesForVMInstances(ctx, &Service.Instances)
	fallbackPids := make([]string, 0)
	for i, instance := range Service.Instances {
		if instance.InstanceType != VM {
			continue
		}
		source := sources[i]
		if source == "" {
			fallbackPids = append(fallbackPids, instance.Pid)
			continue
		}
		query := prometheus.QueryLogBySourcePromql(stepToStr, prometheus.AvgLog, instance.NodeName, source)
		results, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, query, step)
		if err != nil {
			return Service, err
		}
		if len(results) == 0 {
			continue
		}
		if len(Service.LogData) < len(results[0].Values) {
			Service.LogData = make([]prometheus.Points, len(results[0].Values))
		}
		for j := 0; j < len(results[0].Values); j++ {
			Service.LogData[j].TimeStamp = results[0].Values[j].TimeStamp
			Service.LogData[j].Value += results[0].Values[j].Value
		}
	}
	if len(fallbackPids) > 0 {
		return s.ServiceLogRangeDataByPid(ctx, Service, fallbackPids, startTime, endTime, duration, step)
	}
	return Service, nil
}

func logStepToString(step time.Duration) string {
	if step >= time.Hour {
		return strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	}
	if step >= time.Minute {
		return strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	}
	return strconv.FormatInt(int64(step/time.Second), 10) + "s"
}

func (s *service) ServiceLogRangeDataByPod(ctx core.Context, Service *ServiceDetail, pods []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
	if Service == nil {
		return nil, errors.New("service is nil")
	}
	if pods == nil {
		return Service, nil
	}
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}

	LogDataQuery := prometheus.QueryLogPromql(stepToStr, prometheus.AvgLog, pods)
	LogDataRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LogDataQuery, step)
	for _, result := range LogDataRes {
		if len(Service.LogData) < len(result.Values) {
			Service.LogData = append(Service.LogData, make([]prometheus.Points, len(result.Values)-len(Service.LogData))...)
		}

		for i := 0; i < len(result.Values); i++ {
			Service.LogData[i].TimeStamp = result.Values[i].TimeStamp
			Service.LogData[i].Value += result.Values[i].Value
		}
	}

	return Service, err
}

func (s *service) ServiceLogRangeDataByContainerId(ctx core.Context, Service *ServiceDetail, containerIds []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
	if Service == nil {
		return nil, errors.New("service is nil")
	}
	if containerIds == nil {
		return Service, nil
	}
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	//LogDataRes, err = s.promRepo.QueryRangePrometheusErrorLast30min(searchTime)
	LogDataQuery := prometheus.QueryLogPromql(stepToStr, prometheus.AvgLog, containerIds)
	LogDataRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LogDataQuery, step)
	for _, result := range LogDataRes {
		if len(Service.LogData) < len(result.Values) {
			Service.LogData = make([]prometheus.Points, len(result.Values))
		}

		for i := 0; i < len(result.Values); i++ {
			Service.LogData[i].TimeStamp = result.Values[i].TimeStamp
			Service.LogData[i].Value += result.Values[i].Value
		}
		break
	}

	return Service, err
}

func (s *service) ServiceLogRangeDataByPid(ctx core.Context, Service *ServiceDetail, pids []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
	if Service == nil {
		return nil, errors.New("service is nil")
	}
	if pids == nil {
		return Service, nil
	}
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	LogDataQuery := prometheus.QueryLogPromql(stepToStr, prometheus.AvgLog, pids)
	LogDataRes, err := s.promRepo.QueryRangeData(ctx, startTime, endTime, LogDataQuery, step)
	for _, result := range LogDataRes {
		if len(Service.LogData) < len(result.Values) {
			Service.LogData = make([]prometheus.Points, len(result.Values))
		}

		for i := 0; i < len(result.Values); i++ {
			Service.LogData[i].TimeStamp = result.Values[i].TimeStamp
			Service.LogData[i].Value += result.Values[i].Value
		}
		break
	}

	return Service, err
}
