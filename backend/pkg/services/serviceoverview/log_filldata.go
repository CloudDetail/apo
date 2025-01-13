// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) AvgLogByPod(Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	var LogRateRes []prometheus.MetricResult
	queryLog := prometheus.QueryLogPromql(duration, prometheus.AvgLog, pods)
	LogRateRes, err := s.promRepo.QueryData(endTime, queryLog)
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

func (s *service) LogDODByPod(Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogNow := prometheus.QueryLogPromql(duration, prometheus.LogNow, pods)
	logNow, err := s.promRepo.QueryData(endTime, queryLogNow)
	if err != nil {
		return Instances, err
	}
	queryLogYesterday := prometheus.QueryLogPromql(duration, prometheus.LogYesterday, pods)
	logYesterday, err := s.promRepo.QueryData(endTime, queryLogYesterday)
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

func (s *service) LogWOWByPod(Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogLastWeek := prometheus.QueryLogPromql(duration, prometheus.LogLastWeek, pods)
	logLastWeek, err := s.promRepo.QueryData(endTime, queryLogLastWeek)
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
func (s *service) LogRangeDataByPod(Instances *[]Instance, pods []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
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
	LogDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, LogDataQuery, step)
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

func (s *service) AvgLogByContainerId(Instances *[]Instance, containerIds []string, endTime time.Time, duration string) (*[]Instance, error) {
	var LogRateRes []prometheus.MetricResult
	queryLog := prometheus.QueryLogByContainerIdPromql(duration, prometheus.AvgLog, containerIds)
	LogRateRes, err := s.promRepo.QueryData(endTime, queryLog)
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

func (s *service) LogDODByContainerId(Instances *[]Instance, containerIds []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogNow := prometheus.QueryLogByContainerIdPromql(duration, prometheus.LogNow, containerIds)
	logNow, err := s.promRepo.QueryData(endTime, queryLogNow)
	if err != nil {
		return Instances, err
	}
	queryLogYesterday := prometheus.QueryLogByContainerIdPromql(duration, prometheus.LogYesterday, containerIds)
	logYesterday, err := s.promRepo.QueryData(endTime, queryLogYesterday)
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

func (s *service) LogWOWByContainerId(Instances *[]Instance, containerIds []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLastWeek := prometheus.QueryLogByContainerIdPromql(duration, prometheus.LogLastWeek, containerIds)
	logLastWeek, err := s.promRepo.QueryData(endTime, queryLastWeek)
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

func (s *service) LogRangeDataByContainerId(Instances *[]Instance, containerIds []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
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
	LogDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, LogDataQuery, step)
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

func (s *service) AvgLogByPid(Instances *[]Instance, pods []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLog := prometheus.QueryLogByPidPromql(duration, prometheus.AvgLog, pods)
	LogRateRes, err := s.promRepo.QueryData(endTime, queryLog)
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

func (s *service) LogDODByPid(Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogNow := prometheus.QueryLogByPidPromql(duration, prometheus.LogNow, pids)
	logNow, err := s.promRepo.QueryData(endTime, queryLogNow)
	if err != nil {
		return Instances, err
	}
	queryLogYesterday := prometheus.QueryLogByPidPromql(duration, prometheus.LogYesterday, pids)
	logYesterday, err := s.promRepo.QueryData(endTime, queryLogYesterday)
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
func (s *service) LogWOWByPid(Instances *[]Instance, pids []string, endTime time.Time, duration string) (*[]Instance, error) {
	queryLogLastWeek := prometheus.QueryLogByPidPromql(duration, prometheus.LogLastWeek, pids)
	logLastWeek, err := s.promRepo.QueryData(endTime, queryLogLastWeek)
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
func (s *service) LogRangeDataByPid(Instances *[]Instance, pods []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Instance, error) {
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
	LogDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, LogDataQuery, step)
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

func (s *service) ServiceLogRangeDataByPod(Service *ServiceDetail, pods []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
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
	LogDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, LogDataQuery, step)
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

func (s *service) ServiceLogRangeDataByContainerId(Service *ServiceDetail, containerIds []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
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
	LogDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, LogDataQuery, step)
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

func (s *service) ServiceLogRangeDataByPid(Service *ServiceDetail, pids []string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*ServiceDetail, error) {
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
	LogDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, LogDataQuery, step)
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
