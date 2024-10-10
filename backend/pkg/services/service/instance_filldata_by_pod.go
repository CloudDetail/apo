package service

import (
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

func (s *service) InstanceAVGByPod(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	// TODO: Even if there are multiple containers with the same service_name in one Pod, only the pod is considered as the instance.
	//       But now the containers may overwrite the metrics each other. This is not as expected.
	var AvgLatencyRes []prometheus.MetricResult
	queryAvgLatency := prometheus.QueryPodPromql(duration, prometheus.AvgLatency, serviceName, contentKey)
	AvgLatencyRes, err := s.promRepo.QueryLatencyData(endTime, queryAvgLatency)
	for _, result := range AvgLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		containerId := result.Metric.ContainerID
		nodeName := result.Metric.NodeName
		namespace := result.Metric.Namespace
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
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
				InstanceName: pod,
				ContainerId:  containerId,
				InstanceType: serviceoverview.POD,
				ConvertName:  pod,
				NodeName:     nodeName,
				Pod:          pod,
				Namespace:    namespace,
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
	queryAvgError := prometheus.QueryPodPromql(duration, prometheus.AvgError, serviceName, contentKey)
	AvgErrorRateRes, err = s.promRepo.QueryErrorRateData(endTime, queryAvgError)
	for _, result := range AvgErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].AvgErrorRate = &value
				}
				break
			}
		}
	}

	var AvgTPSRes []prometheus.MetricResult
	queryAvgTPS := prometheus.QueryPodPromql(duration, prometheus.AvgTPS, serviceName, contentKey)
	AvgTPSRes, err = s.promRepo.QueryData(endTime, queryAvgTPS)
	for _, result := range AvgTPSRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].AvgTPS = &value
				}
				break
			}
		}
	}
	return Instances, err
}

func (s *service) InstanceDODByPod(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	latencyDODquery := prometheus.QueryPodPromql(duration, prometheus.LatencyDOD, serviceName, contentKey)
	latencyDoDres, err := s.promRepo.QueryData(endTime, latencyDODquery)

	for _, result := range latencyDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		containerId := result.Metric.ContainerID
		nodeName := result.Metric.NodeName
		namespace := result.Metric.Namespace
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
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
				InstanceName: pod,
				ContainerId:  containerId,
				InstanceType: serviceoverview.POD,
				ConvertName:  pod,
				NodeName:     nodeName,
				Pod:          pod,
				Namespace:    namespace,
				Pid:          pid,
				NodeIP:       nodeIP,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newInstance.LatencyDayOverDay = &value
			}
			*Instances = append(*Instances, newInstance)
		}
	}

	errorDODquery := prometheus.QueryPodPromql(duration, prometheus.ErrorDOD, serviceName, contentKey)
	errorDoDres, err := s.promRepo.QueryData(endTime, errorDODquery)
	// 更新wrongUrls中的内容
	for _, result := range errorDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
				if !math.IsInf(value, 0) { //为无穷大时则赋值Int64
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

	tpsDODquery := prometheus.QueryPodPromql(duration, prometheus.TPSDOD, serviceName, contentKey)
	tpsResults, err := s.promRepo.QueryData(endTime, tpsDODquery)
	for _, result := range tpsResults {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].TPSDayOverDay = &value
				}
				break
			}
		}
	}
	return Instances, err
}

func (s *service) InstanceWOWByPod(Instances *[]serviceoverview.Instance, serviceName string, contentKey string, endTime time.Time, duration string) (*[]serviceoverview.Instance, error) {
	var LatencyWoWRes []prometheus.MetricResult
	latencyWOWquery := prometheus.QueryPodPromql(duration, prometheus.LatencyWOW, serviceName, contentKey)
	LatencyWoWRes, err := s.promRepo.QueryData(endTime, latencyWOWquery)
	for _, result := range LatencyWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		containerId := result.Metric.ContainerID
		pid := result.Metric.PID
		nodeIP := result.Metric.NodeIP
		found := false
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
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
				InstanceName: pod,
				ContainerId:  containerId,
				InstanceType: serviceoverview.POD,
				ConvertName:  pod,
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
	TPSWOWquery := prometheus.QueryPodPromql(duration, prometheus.TPSWOW, serviceName, contentKey)
	TPSWoWRes, err = s.promRepo.QueryData(endTime, TPSWOWquery)
	for _, result := range TPSWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Instances)[i].TPSWeekOverWeek = &value
				}
				break
			}
		}
	}

	var ErrorWoWRes []prometheus.MetricResult
	errorWoWquery := prometheus.QueryPodPromql(duration, prometheus.ErrorWOW, serviceName, contentKey)
	ErrorWoWRes, err = s.promRepo.QueryData(endTime, errorWoWquery)
	for _, result := range ErrorWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		value := result.Values[0].Value
		for i, Instance := range *Instances {
			if Instance.ContentKey == contentKey && Instance.SvcName == serviceName && Instance.InstanceName == pod {
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

// InstanceRangeDataByPod  查询曲线图
func (s *service) InstanceRangeDataByPod(Instances *[]serviceoverview.Instance, contentKey string, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]serviceoverview.Instance, error) {
	if Instances == nil {
		return nil, errors.New("instances is nil")
	}
	var stepToStr string

	stepMinutes := float64(step) / float64(time.Minute)
	// 格式化为字符串，保留一位小数
	stepToStr = fmt.Sprintf("%.1fm", stepMinutes)

	errorDataQuery := prometheus.QueryPodRangePromql(stepToStr, prometheus.ErrorData, contentKey, serviceName)
	errorDataRes, err := s.promRepo.QueryRangeErrorData(startTime, endTime, errorDataQuery, step)
	for _, result := range errorDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD

		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == pod {
				(*Instances)[i].ErrorRateData = result.Values
				break
			}
		}
	}

	var LatencyDataRes []prometheus.MetricResult
	// 分批处理 ContentKeys

	latencyDataQuery := prometheus.QueryPodRangePromql(stepToStr, prometheus.LatencyData, contentKey, serviceName)
	LatencyDataRes, err = s.promRepo.QueryRangeLatencyData(startTime, endTime, latencyDataQuery, step)
	for _, result := range LatencyDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == pod {
				(*Instances)[i].LatencyData = result.Values
				break
			}
		}
	}

	var TPSLastDataRes []prometheus.MetricResult
	// 分批处理 ContentKeys
	//TPSLastDataRes, err = s.promRepo.QueryRangePrometheusTPSLast30min(searchTime)
	TPSDataQuery := prometheus.QueryPodRangePromql(stepToStr, prometheus.TPSData, contentKey, serviceName)
	TPSLastDataRes, err = s.promRepo.QueryRangeData(startTime, endTime, TPSDataQuery, step)
	for _, result := range TPSLastDataRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		pod := result.Metric.POD
		for i, instance := range *Instances {
			if instance.ContentKey == contentKey && instance.SvcName == serviceName && instance.InstanceName == pod {
				(*Instances)[i].TPSData = result.Values
				break
			}
		}
	}
	return Instances, err
}
