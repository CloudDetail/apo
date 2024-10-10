package service

import (
	"fmt"
	"math"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
	"github.com/pkg/errors"
)

func (s *service) QueryInstancesNodeName(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time) (*[]serviceoverview.Instance, error) {
	var Res []prometheus.MetricResult
	query := prometheus.QueryNodeName(serviceName, contentKey)
	Res, err := s.promRepo.QueryData(endTime, query)
	for _, result := range Res {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pod := result.Metric.POD

		//log.Printf("%v", pod)
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].AvgErrorRate = &value
				}
				break
			}
		}
		if !found && contentKey != "" && serviceName != "" {
			newInstance := serviceoverview.Instance{
				ContentKey: contentKey,
				SvcName:    serviceName,
				Pod:        pod,
				NodeName:   nodeName,
			}
			*Instances = append(*Instances, newInstance)
		}
	}
	return Instances, err
}

func (s *service) InstanceAVGByPid(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	var AvgLatencyRes []prometheus.MetricResult
	queryAvgLatency := prometheus.QueryPidPromql(duration, prometheus.AvgLatency, serviceName, contentKey)
	AvgLatencyRes, err := s.promRepo.QueryLatencyData(endTime, queryAvgLatency)
	for _, result := range AvgLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		nodeIp := result.Metric.NodeIP
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
				Pid:          pid,
				InstanceType: serviceoverview.VM,
				ConvertName:  pid,
				NodeName:     nodeName,
				NodeIP:       nodeIp,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.AvgLatency = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	var AvgErrorRateRes []prometheus.MetricResult
	queryAvgError := prometheus.QueryPidPromql(duration, prometheus.AvgError, serviceName, contentKey)
	AvgErrorRateRes, err = s.promRepo.QueryErrorRateData(endTime, queryAvgError)
	for _, result := range AvgErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
	queryAvgTPS := prometheus.QueryPidPromql(duration, prometheus.AvgTPS, serviceName, contentKey)
	AvgTPSRes, err = s.promRepo.QueryData(endTime, queryAvgTPS)
	for _, result := range AvgTPSRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
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

func (s *service) InstanceDODByPid(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	latencyDODquery := prometheus.QueryPidPromql(duration, prometheus.LatencyDOD, serviceName, contentKey)
	latencyDoDres, err := s.promRepo.QueryData(endTime, latencyDODquery)

	for _, result := range latencyDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
				Pid:          pid,
				InstanceType: serviceoverview.VM,
				ConvertName:  pid,
				NodeName:     nodeName,
				NodeIP:       nodeIP,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.LatencyDayOverDay = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	errorDODquery := prometheus.QueryPidPromql(duration, prometheus.ErrorDOD, serviceName, contentKey)
	errorDoDres, err := s.promRepo.QueryData(endTime, errorDODquery)
	// 更新wrongUrls中的内容
	for _, result := range errorDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == instanceName {
				if !math.IsInf(value, 0) { //为无穷大时则赋值MaxInt64
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

	tpsDODquery := prometheus.QueryPidPromql(duration, prometheus.TPSDOD, serviceName, contentKey)
	tpsResults, err := s.promRepo.QueryData(endTime, tpsDODquery)
	for _, result := range tpsResults {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
func (s *service) InstanceWOWByPid(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	var LatencyWoWRes []prometheus.MetricResult
	latencyWOWquery := prometheus.QueryPidPromql(duration, prometheus.LatencyWOW, serviceName, contentKey)
	LatencyWoWRes, err := s.promRepo.QueryData(endTime, latencyWOWquery)
	for _, result := range LatencyWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
				Pid:          pid,
				InstanceType: serviceoverview.VM,
				ConvertName:  pid,
				NodeIP:       nodeIP,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.LatencyWeekOverWeek = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	var TPSWoWRes []prometheus.MetricResult
	TPSWOWquery := prometheus.QueryPidPromql(duration, prometheus.TPSWOW, serviceName, contentKey)
	TPSWoWRes, err = s.promRepo.QueryData(endTime, TPSWOWquery)
	for _, result := range TPSWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
	errorWoWquery := prometheus.QueryPidPromql(duration, prometheus.ErrorWOW, serviceName, contentKey)
	ErrorWoWRes, err = s.promRepo.QueryData(endTime, errorWoWquery)
	for _, result := range ErrorWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
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

// InstanceRangeDataByPid   查询曲线图
func (s *service) InstanceRangeDataByPid(Instances *[]serviceoverview.Instance, contentKey string, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]serviceoverview.Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	var stepToStr string
	stepMinutes := float64(step) / float64(time.Minute)
	// 格式化为字符串，保留一位小数
	stepToStr = fmt.Sprintf("%.1fm", stepMinutes)

	//errorDataRes, err = s.promRepo.QueryRangePrometheusErrorLast30min(searchTime)
	errorDataQuery := prometheus.QueryPidRangePromql(stepToStr, prometheus.ErrorData, contentKey, serviceName)
	errorDataRes, err := s.promRepo.QueryRangeData(startTime, endTime, errorDataQuery, step)

	for _, result := range errorDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID

		instanceName := serviceName + "@" + nodeName + "@" + pid
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
	latencyDataQuery := prometheus.QueryPidRangePromql(stepToStr, prometheus.LatencyData, contentKey, serviceName)
	LatencyDataRes, err = s.promRepo.QueryRangeLatencyData(startTime, endTime, latencyDataQuery, step)

	for _, result := range LatencyDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
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
	TPSDataQuery := prometheus.QueryPidRangePromql(stepToStr, prometheus.TPSData, contentKey, serviceName)
	TPSLastDataRes, err = s.promRepo.QueryRangeData(startTime, endTime, TPSDataQuery, step)

	for _, result := range TPSLastDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		nodeName := result.Metric.NodeName
		pid := result.Metric.PID
		instanceName := serviceName + "@" + nodeName + "@" + pid
		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == instanceName {
				(*Instances)[i].TPSData = result.Values
				break
			}
		}
	}

	return Instances, err
}
