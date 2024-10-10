package service

import (
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

func (s *service) InstanceAVGByContainerId(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	var AvgLatencyRes []prometheus.MetricResult
	queryAvgLatency := prometheus.QueryContainerIdPromql(duration, prometheus.AvgLatency, serviceName, contentKey)
	AvgLatencyRes, err := s.promRepo.QueryLatencyData(endTime, queryAvgLatency)
	for _, result := range AvgLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		nodeIP := result.Metric.NodeIP
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].AvgLatency = &value
				}
				break
			}
		}
		if !found {
			newInstance := serviceoverview.Instance{
				ContentKey:   contentKey,
				SvcName:      serviceName,
				InstanceName: instanceName,
				ContainerId:  containerId,
				InstanceType: serviceoverview.CONTAINER,
				ConvertName:  nodeName,
				NodeName:     nodeName,
				Pid:          pid,
				NodeIP:       nodeIP,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.AvgLatency = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	var AvgErrorRateRes []prometheus.MetricResult
	queryAvgError := prometheus.QueryContainerIdPromql(duration, prometheus.AvgError, serviceName, contentKey)
	AvgErrorRateRes, err = s.promRepo.QueryErrorRateData(endTime, queryAvgError)
	for _, result := range AvgErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].AvgErrorRate = &value
				}
				break
			}
		}
	}

	var AvgTPSRes []prometheus.MetricResult
	queryAvgTPS := prometheus.QueryContainerIdPromql(duration, prometheus.AvgTPS, serviceName, contentKey)
	AvgTPSRes, err = s.promRepo.QueryData(endTime, queryAvgTPS)
	for _, result := range AvgTPSRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].AvgTPS = &value
				}
				break
			}
		}
	}
	return Instances, err
}

func (s *service) InstanceDODByContainerId(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	latencyDODquery := prometheus.QueryContainerIdPromql(duration, prometheus.LatencyDOD, serviceName, contentKey)
	latencyDoDres, err := s.promRepo.QueryData(endTime, latencyDODquery)
	for _, result := range latencyDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].LatencyDayOverDay = &value
				}
				break
			}
		}
		if !found {
			newInstance := serviceoverview.Instance{
				ContentKey:   contentKey,
				SvcName:      serviceName,
				InstanceName: instanceName,
				ContainerId:  containerId,
				InstanceType: serviceoverview.CONTAINER,
				ConvertName:  nodeName,
				NodeName:     nodeName,
				Pid:          pid,
				NodeIP:       nodeIP,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.LatencyDayOverDay = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	errorDODquery := prometheus.QueryContainerIdPromql(duration, prometheus.ErrorDOD, serviceName, contentKey)
	errorDoDres, err := s.promRepo.QueryData(endTime, errorDODquery)
	// 更新wrongUrls中的内容
	for _, result := range errorDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则赋值MaxFloat64
					(*Instances)[i].ErrorRateDayOverDay = &value
				} else {
					var value float64
					value = serviceoverview.RES_MAX_VALUE
					pointer := &value
					(*Instances)[i].ErrorRateDayOverDay = pointer
				}

				break
			}
		}
	}

	tpsDODquery := prometheus.QueryContainerIdPromql(duration, prometheus.TPSDOD, serviceName, contentKey)
	tpsResults, err := s.promRepo.QueryData(endTime, tpsDODquery)
	for _, result := range tpsResults {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].TPSDayOverDay = &value
				}
				break
			}
		}
	}
	return Instances, err
}

func (s *service) InstanceWOWByContainerId(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	var LatencyWoWRes []prometheus.MetricResult
	latencyWOWquery := prometheus.QueryContainerIdPromql(duration, prometheus.LatencyWOW, serviceName, contentKey)
	LatencyWoWRes, err := s.promRepo.QueryData(endTime, latencyWOWquery)
	for _, result := range LatencyWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].LatencyWeekOverWeek = &value
				}
				break
			}
		}
		if !found {
			newInstance := serviceoverview.Instance{
				ContentKey:   contentKey,
				SvcName:      serviceName,
				InstanceName: instanceName,
				ContainerId:  containerId,
				InstanceType: serviceoverview.CONTAINER,
				ConvertName:  nodeName,
				NodeName:     nodeName,
				Pid:          pid,
				NodeIP:       nodeIP,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.LatencyWeekOverWeek = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	var TPSWoWRes []prometheus.MetricResult
	TPSWOWquery := prometheus.QueryContainerIdPromql(duration, prometheus.TPSWOW, serviceName, contentKey)
	TPSWoWRes, err = s.promRepo.QueryData(endTime, TPSWOWquery)
	for _, result := range TPSWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].TPSWeekOverWeek = &value
				}
				break
			}
		}
	}

	var ErrorWoWRes []prometheus.MetricResult
	errorWoWquery := prometheus.QueryContainerIdPromql(duration, prometheus.ErrorWOW, serviceName, contentKey)
	ErrorWoWRes, err = s.promRepo.QueryData(endTime, errorWoWquery)
	for _, result := range ErrorWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则赋值MaxInt64
					(*Instances)[i].ErrorRateWeekOverWeek = &value
				} else {
					var value float64
					value = serviceoverview.RES_MAX_VALUE
					pointer := &value
					(*Instances)[i].ErrorRateWeekOverWeek = pointer
				}
				break
			}
		}
	}
	return Instances, err
}

// InstanceRangeDataByContainerId   查询曲线图
func (s *service) InstanceRangeDataByContainerId(Instances *[]serviceoverview.Instance, contentKey string, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]serviceoverview.Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	var stepToStr string

	stepMinutes := float64(step) / float64(time.Minute)
	// 格式化为字符串，保留一位小数
	stepToStr = fmt.Sprintf("%.1fm", stepMinutes)

	errorDataQuery := prometheus.QueryContainerIdRangePromql(stepToStr, prometheus.ErrorData, contentKey, serviceName)
	errorDataRes, err := s.promRepo.QueryRangeErrorData(startTime, endTime, errorDataQuery, step)
	for _, result := range errorDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == instanceName {
				(*Instances)[i].ErrorRateData = result.Values
				break
			}
		}
	}

	var LatencyDataRes []prometheus.MetricResult
	// 分批处理 ContentKeys

	//LatencyDataRes, err = s.promRepo.QueryRangePrometheusLatencyLast30min(searchTime)
	latencyDataQuery := prometheus.QueryContainerIdRangePromql(stepToStr, prometheus.LatencyData, contentKey, serviceName)
	LatencyDataRes, err = s.promRepo.QueryRangeLatencyData(startTime, endTime, latencyDataQuery, step)
	for _, result := range LatencyDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == instanceName {
				(*Instances)[i].LatencyData = result.Values
				break
			}
		}
	}

	var TPSLastDataRes []prometheus.MetricResult
	// 分批处理 ContentKeys
	//TPSLastDataRes, err = s.promRepo.QueryRangePrometheusTPSLast30min(searchTime)
	TPSDataQuery := prometheus.QueryContainerIdRangePromql(stepToStr, prometheus.TPSData, contentKey, serviceName)
	TPSLastDataRes, err = s.promRepo.QueryRangeData(startTime, endTime, TPSDataQuery, step)
	for _, result := range TPSLastDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		containerId := result.Metric.ContainerID
		instanceName := serviceName + "@" + nodeName + "@" + containerId
		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == instanceName {
				(*Instances)[i].TPSData = result.Values
				break
			}
		}
	}
	return Instances, err
}
